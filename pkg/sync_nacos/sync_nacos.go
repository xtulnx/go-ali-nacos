package sync_nacos

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/logger"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pelletier/go-toml"
	"go-ali-nacos/pkg/config"
	"go-ali-nacos/pkg/logs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SyncNacos 每个实例只对接一个命名空间
type SyncNacos struct {
	root, parent *SyncNacos
	key          jobKey

	cfgMain     config.Config               // 当前配置
	namespaceId string                      // 命名空间
	client      config_client.IConfigClient // 连接器
	jobs        map[jobKey]*jobStatus       // 任务
	children    map[jobKey]*SyncNacos       // 子配置节点
}

type jobKey struct {
	group, dataId string
}

type jobFile struct {
	outfile string // 文件路径
	hash    string // 校验
	schema  string
}

type jobStatus struct {
	mc sync.RWMutex
	rc chan struct{}

	outfile map[jobKey]jobFile
	exec    string
	params  []string
}

// -o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-o-

func (J *SyncNacos) Close() error {
	//
	for k, _ := range J.jobs {
		_ = J.client.CancelListenConfig(vo.ConfigParam{
			DataId: k.dataId,
			Group:  k.group,
		})
	}

	return nil
}

func Hash(content string) (md string) {
	h := md5.New()
	_, _ = io.WriteString(h, content)
	md = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func findSchema(s string) string {
	if p := strings.Index(s, "://"); p > 0 {
		return s[:p+3]
	}
	return ""
}

func NewClient(cfgNacos config.NacosConfig) (config_client.IConfigClient, error) {
	if cfgNacos.NamespaceId == "" {
		return nil, fmt.Errorf("缺少 namespaceId")
	}
	if cfgNacos.LogLevel != "" {
		_ = logger.InitLogger(
			logger.BuildLoggerConfig(constant.ClientConfig{
				TimeoutMs:            0,
				ListenInterval:       0,
				BeatInterval:         0,
				NamespaceId:          "",
				AppName:              "",
				Endpoint:             "",
				RegionId:             "",
				AccessKey:            "",
				SecretKey:            "",
				OpenKMS:              false,
				CacheDir:             "",
				UpdateThreadNum:      0,
				NotLoadCacheAtStart:  false,
				UpdateCacheWhenEmpty: false,
				Username:             "",
				Password:             "",
				LogDir:               "",
				LogLevel:             "",
				LogSampling:          nil,
				ContextPath:          "",
				LogRollingConfig:     nil,
			}),
		)
	}
	clientConfig := &constant.ClientConfig{
		//
		Endpoint:    cfgNacos.Endpoint + ":8080",
		NamespaceId: cfgNacos.NamespaceId,
		AccessKey:   cfgNacos.AccessKey,
		SecretKey:   cfgNacos.SecretKey,

		TimeoutMs: 5 * 1000,
		//ListenInterval: 30 * 1000,

		AppName:     cfgNacos.AppName,
		RegionId:    cfgNacos.RegionId,
		Username:    cfgNacos.Username,
		Password:    cfgNacos.Password,
		ContextPath: cfgNacos.ContextPath,

		LogLevel: cfgNacos.LogLevel,

		OpenKMS:              false,
		NotLoadCacheAtStart:  true,
		UpdateCacheWhenEmpty: true,
	}

	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig: clientConfig,
		},
	)
	return client, err
}

// 检查所有任务
func newJobs(cfgJobs []config.NacosJobConfig, namespaceId string) map[jobKey]*jobStatus {
	var jobs = make(map[jobKey]*jobStatus)
	for _, cj := range cfgJobs {
		outfile := make(map[jobKey]jobFile)
		for _, f := range cj.File {
			if f.Group == "" || f.DataId == "" || f.Outfile == "" {
				zap.S().Debugf("[SKIP]Invalid %s/%s/%s : %s", namespaceId, f.Group, f.DataId, f.Outfile)
				continue
			}
			k := jobKey{f.Group, f.DataId}
			if _, ok := jobs[k]; ok {
				zap.S().Debugf("[SKIP]Exists %s/%s/%s : %s", namespaceId, f.Group, f.DataId, f.Outfile)
				continue
			}
			if _, ok := outfile[k]; ok {
				zap.S().Debugf("[SKIP]Exists %s/%s/%s : %s", namespaceId, f.Group, f.DataId, f.Outfile)
				continue
			}

			// FIXME: 检查输出路径是否为系统文件
			// ...

			_hash, _outfile := "", f.Outfile
			_schema := findSchema(f.Outfile)
			switch _schema {
			case config.SCHEME_SYS_CONFIG, config.SCHEME_SYS_MEMORY:
				_outfile = _outfile[len(_schema):]
			default:
				if b1, err := ioutil.ReadFile(f.Outfile); err == nil {
					_hash = Hash(string(b1))
				}
			}

			outfile[k] = jobFile{outfile: _outfile, hash: _hash, schema: _schema}
		}
		if len(outfile) > 0 {
			st := &jobStatus{
				outfile: outfile,
				exec:    cj.Exec,
				params:  cj.Params,
				rc:      make(chan struct{}, 1),
			}
			for k, _ := range outfile {
				jobs[k] = st
			}
		}
	}
	return jobs
}

func NewSyncNacos(cfg config.Config) (*SyncNacos, error) {
	//
	client, err := NewClient(cfg.NacosCfg)
	if err != nil {
		return nil, err
	}
	// 添加配置
	var jobs = newJobs(cfg.NacosJobs, cfg.NacosCfg.NamespaceId)

	n := &SyncNacos{
		root:        nil,
		parent:      nil,
		cfgMain:     cfg,
		namespaceId: cfg.NacosCfg.NamespaceId,
		client:      client,
		jobs:        jobs,
		children:    make(map[jobKey]*SyncNacos),
	}

	return n, nil
}

func (J *SyncNacos) tryLoadSubConfig(dataId, data string) (*config.Config, error) {
	c1 := config.Config{}
	ext := strings.ToLower(filepath.Ext(dataId))
	if data == "" {
		return nil, fmt.Errorf("无效配置")
	}
	var err error
	switch ext {
	case ".yml", ".yaml":
		err = yaml.Unmarshal([]byte(data), &c1)
	case ".json":
		err = json.Unmarshal([]byte(data), &c1)
	case ".toml":
		err = toml.Unmarshal([]byte(data), &c1)
	default:
		err = fmt.Errorf("不支持格式:%s", ext)
	}
	if err == nil {
		return &c1, nil
	}
	return nil, err
}

// ExtendSub 远端资源是「配置」，扩展所有子节点
// 如果该子节点存在，则先检查有无更新
func (J *SyncNacos) ExtendSub(ctx context.Context, k jobKey, data string) error {
	if data == "" {
		zap.S().Debug("释放子任务 %s/%s/%s", J.namespaceId, k.group, k.dataId)
		if last, ok := J.children[k]; ok {
			last.free()
			delete(J.children, k)
		}
		return nil
	}

	cfg, err := J.tryLoadSubConfig(k.dataId, data)
	if err != nil {
		return err
	}

	// 继承授权配置
	if cfg.NacosCfg.NamespaceId == "" {
		cfg.NacosCfg = J.cfgMain.NacosCfg
	}

	sub, err := NewSyncNacos(*cfg)
	if err != nil {
		zap.S().Warnf("创建子节点失败 %s/%s/%s: %s", J.namespaceId, k.group, k.dataId, err)
		return err
	}
	sub.parent, sub.root = J, J.root

	if last, ok := J.children[k]; ok {
		// TODO: 检查是否需要清理子树，这里省事直接释放
		last.free()
		delete(J.children, k)
	}
	J.children[k] = sub
	go func(ctx context.Context, s *SyncNacos, k jobKey) {
		s.refresh(ctx)
		s.watch(ctx)
		zap.S().Debugf("[SUCC]init sub %s/%s/%s", s.namespaceId, k.group, k.dataId)
	}(ctx, sub, k)
	return nil
}

func (J *SyncNacos) free() {
	for _, sb := range J.children {
		sb.free()
	}
	if J.client != nil {
		for k, _ := range J.jobs {
			err := J.client.CancelListenConfig(vo.ConfigParam{DataId: k.dataId, Group: k.group})
			if err != nil {
			}
		}
	}
	J.children = make(map[jobKey]*SyncNacos)
	J.jobs = make(map[jobKey]*jobStatus)
}

// 内容更新
func (J *SyncNacos) onChange(ctx context.Context, group, dataId, data string) {
	zap.S().Debugf("onChange: %s/%s/%s", J.namespaceId, group, dataId)
	k := jobKey{group, dataId}
	if st, ok := J.jobs[k]; ok {
		if of, ok1 := st.outfile[k]; ok1 {
			hash := Hash(data)
			if hash == of.hash {
				// 文件相同，跳过
				zap.S().Infof("[SKIP] %s/%s/%s: %s", J.namespaceId, group, dataId, hash)
				return
			}
			env := os.Environ()
			env = append(env, "OUTFILE="+of.outfile)
			switch of.schema {
			case config.SCHEME_SYS_CONFIG:
				err := J.ExtendSub(ctx, k, data)
				if err != nil {
					zap.S().Warnf("[FAIL]ExtendSub %s/%s/%s: %v", J.namespaceId, group, dataId, err)
					return
				}
				env = append(env, "CONTENT="+data)
			case config.SCHEME_SYS_MEMORY:
				env = append(env, "CONTENT="+data)
			default:
				err := WriteFile(of.outfile, data)
				if err != nil {
					zap.S().Warnf("[FAIL]Write file %s: %v", of.outfile, err)
					return
				}
			}
			of.hash = hash
			// 启动任务
			go J.exec(ctx, st, k, env)
			return
		}
	}
	zap.S().Warnf("[MISS] %s/%s/%s", J.namespaceId, group, dataId)
}

func (J *SyncNacos) exec(ctx context.Context, st *jobStatus, k jobKey, env []string) {
	if st == nil || st.exec == "" {
		return
	}
	select {
	case st.rc <- struct{}{}:
	default:
		zap.S().Infof("[EXEC] cancel  %s/%s/%s", J.namespaceId, k.group, k.dataId)
		return
	}
	defer func() {
		<-st.rc
		zap.S().Infof("[EXEC] end %s/%s/%s", J.namespaceId, k.group, k.dataId)
	}()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	pCmd := exec.CommandContext(ctx, st.exec, st.params...)
	//pCmd.Stdin = bytes.NewReader(content.Bytes())
	pCmd.Stdout = os.Stdout
	pCmd.Env = append(pCmd.Env, env...)
	zap.S().Infof("[EXEC] start %s/%s/%s", J.namespaceId, k.group, k.dataId)
	err := pCmd.Run()
	if err != nil {
		//return nil, "", err
		zap.S().Warnf("[FAIL]exec %s/%s/%s: %v", J.namespaceId, k.group, k.dataId, err)
		return
	}
	zap.S().Infof("[EXEC]finish %s/%s/%s", J.namespaceId, k.group, k.dataId)
}

// refresh 主动刷新所有配置（配置有变化才执行外部命令）
func (J *SyncNacos) refresh(ctx context.Context) {
	for k, _ := range J.jobs {
		select {
		case <-ctx.Done():
			return
		default:
			content, err := J.client.GetConfig(vo.ConfigParam{DataId: k.dataId, Group: k.group})
			if err != nil {
				zap.S().Warnf("[FAIL]get for %s/%s/%s: %v", J.namespaceId, k.group, k.dataId, err)
			} else {
				J.onChange(ctx, k.group, k.dataId, content)
			}
		}
	}
}

func (J *SyncNacos) watch(ctx context.Context) {
	onChange := func(namespace, group, dataId, data string) {
		if namespace == J.namespaceId {
			J.onChange(ctx, group, dataId, data)
		} else {
			zap.S().Warnf("[INVALID]onChange %s/%s/%s: %v", namespace, group, dataId, data)
		}
	}
	for k, _ := range J.jobs {
		err := J.client.ListenConfig(vo.ConfigParam{
			DataId:   k.dataId,
			Group:    k.group,
			OnChange: onChange,
		})

		if err != nil {
			zap.S().Warnf("[FAIL]listen for %s/%s/%s : %v", J.namespaceId, k.group, k.dataId, err)
		}
	}
}

// 启动
func (J *SyncNacos) Run(ctx context.Context) error {
	J.refresh(ctx)
	J.watch(ctx)
	for {
		select {
		case <-ctx.Done():
			//log.Printf("[RUN] end")
			return nil
		}
	}
}

func Main(ctx context.Context, cfg config.Config, group, dataId string) error {
	var logLevel = zapcore.DebugLevel
	if cfg.NacosCfg.LogLevel != "" {
		_ = logLevel.UnmarshalText([]byte(cfg.NacosCfg.LogLevel))
	}
	logs.InitZapLogger("", "", logLevel)

	if group != "" && dataId != "" {
		cfg.NacosJobs = append(cfg.NacosJobs, config.NacosJobConfig{
			File: []config.NacosJobFileConfig{{
				DataId:  dataId,
				Group:   group,
				Outfile: config.SCHEME_SYS_CONFIG,
			}},
		})
	}
	J, err := NewSyncNacos(cfg)
	if err != nil {
		return err
	}
	if len(cfg.NacosJobs) == 0 {
		zap.S().Warnf("需要至少一个任务")
		return nil
	}
	err = J.Run(ctx)
	return err
}
