package sync_nacos

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/logger"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go-ali-nacos/pkg/config"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// SyncNacos 每个实例只对接一个命名空间
type SyncNacos struct {
	cfgMain     config.Config
	namespaceId string // 命名空间
	client      config_client.IConfigClient
	jobs        map[jobKey]*jobStatus
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
		_ = logger.InitLogger(logger.Config{
			Level: cfgNacos.LogLevel,
		})
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
				log.Printf("[SKIP]Invalid %s/%s/%s : %s", namespaceId, f.Group, f.DataId, f.Outfile)
				continue
			}
			k := jobKey{f.Group, f.DataId}
			if _, ok := jobs[k]; ok {
				log.Printf("[SKIP]Exists %s/%s/%s : %s", namespaceId, f.Group, f.DataId, f.Outfile)
				continue
			}
			if _, ok := outfile[k]; ok {
				log.Printf("[SKIP]Exists %s/%s/%s : %s", namespaceId, f.Group, f.DataId, f.Outfile)
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
		cfgMain:     cfg,
		namespaceId: cfg.NacosCfg.NamespaceId,
		client:      client,
		jobs:        jobs,
	}

	return n, nil
}

// 内容更新
func (J *SyncNacos) onChange(ctx context.Context, group, dataId, data string) {
	log.Printf("onChange: %s/%s/%s\n%v", J.namespaceId, group, dataId, data)
	k := jobKey{group, dataId}
	if st, ok := J.jobs[k]; ok {
		if of, ok1 := st.outfile[k]; ok1 {
			hash := Hash(data)
			if hash == of.hash {
				// 文件相同，跳过
				log.Printf("[SKIP] %s/%s/%s: %s", J.namespaceId, group, dataId, hash)
				return
			}

			switch of.schema {
			case config.SCHEME_SYS_CONFIG:
				// TODO:
			case config.SCHEME_SYS_MEMORY:
				// TODO:
			default:
				err := ioutil.WriteFile(of.outfile, []byte(data), fs.ModePerm)
				if err != nil {
					log.Printf("[FAIL]Write file %s: %v", of.outfile, err)
					return
				}
			}
			of.hash = hash
			// 启动任务
			go J.exec(ctx, st)
			return
		}
	}
	log.Printf("[MISS] %s/%s/%s", J.namespaceId, group, dataId)
}

func (J *SyncNacos) exec(ctx context.Context, st *jobStatus) {
	select {
	case st.rc <- struct{}{}:
	default:
		log.Printf("[EXEC] cancel")
		return
	}
	defer func() {
		<-st.rc
		log.Printf("[EXEC] end")
	}()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	pCmd := exec.CommandContext(ctx, st.exec, st.params...)
	//pCmd.Stdin = bytes.NewReader(content.Bytes())
	pCmd.Stdout = os.Stdout
	//pCmd.Env = append(pCmd.Env, pb.Env...)
	log.Printf("[EXEC] start")
	err := pCmd.Run()
	if err != nil {
		//return nil, "", err
		log.Printf("[FAIL]exec %v", err)
		return
	}
	log.Printf("[EXEC] finish")
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
				log.Printf("[FAIL]get for %s/%s/%s: %v", J.namespaceId, k.group, k.dataId, err)
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
			log.Printf("[INVALID]onChange %s/%s/%s: %v", namespace, group, dataId, data)
		}
	}
	for k, _ := range J.jobs {
		err := J.client.ListenConfig(vo.ConfigParam{
			DataId:   k.dataId,
			Group:    k.group,
			OnChange: onChange,
		})

		if err != nil {
			log.Printf("[FAIL]listen for %s/%s/%s : %v", J.namespaceId, k.group, k.dataId, err)
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

func Main(ctx context.Context, cfg config.Config) error {
	J, err := NewSyncNacos(cfg)
	if err != nil {
		return err
	}
	err = J.Run(ctx)
	return err
}
