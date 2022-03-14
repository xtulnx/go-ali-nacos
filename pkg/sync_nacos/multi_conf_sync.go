package sync_nacos

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"go-ali-nacos/pkg/common"
	"go-ali-nacos/pkg/config"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pelletier/go-toml"
	"go.uber.org/zap"
)

var uniqueKey sync.Map

func isExists(namespace, group, dataId string) bool {
	key := fmt.Sprintf("%s%s%s", namespace, group, dataId)
	_, ok := uniqueKey.Load(key)
	if !ok {
		uniqueKey.Store(key, struct{}{})
	}
	return ok
}

func delUniqueKey(namespace, group, dataId string) {
	uniqueKey.Delete(fmt.Sprintf("%s%s%s", namespace, group, dataId))
}

type Node struct {
	Client      config_client.IConfigClient
	Children    []*Node // 子节点
	NamespaceId string  // 命名空间
	Jobs        []*Job  // 任务
	// root节点不存在这两个值
	DataId string
	Group  string
	sync.Mutex
}

//
func (n *Node) CheckDiff(group string, dataId string, data string) {
	for _, child := range n.Children {
		if child.Group == group && child.DataId == dataId {
			child.UnWatch()
			trees, err := toml.Load(data)
			if err != nil {
				zap.L().Error("加载数据出错", zap.Error(err))
				return
			}
			var newCfg config.Config
			err = trees.Unmarshal(&newCfg)
			if err != nil {
				zap.L().Error("获取的数据序列化出错", zap.Error(err), zap.String("data", data))
				return
			}
			jobs := make([]*Job, 0, len(newCfg.NacosJobs))
			for _, j := range newCfg.NacosJobs {
				timeout := time.Duration(j.Timeout) * time.Second
				if timeout == 0 {
					timeout = 30 * time.Second
				}
				job := &Job{
					Exec:     j.Exec,
					Params:   j.Params,
					Timeout:  time.Duration(timeout),
					outfiles: j.File,
				}
				jobs = append(jobs, job)
			}
			child.Jobs = jobs
			child.newChildren(&newCfg)
			child.Watch()
		}
	}
	// uniqueKey.Range(func(key, value interface{}) bool {
	// 	zap.L().Debug("更改后存在的文件key", zap.String("key", key.(string)))
	// 	return true
	// })
}

func (n *Node) writeFile(dataId, group, outfile string) {
	// 解析更改内容
	if len(findSchema(outfile)) <= 0 {
		data, err := n.Client.GetConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
		})
		if err != nil {
			zap.L().Error("get config file error", zap.Error(err), zap.String("data", data))
			return
		}
		if len(data) <= 0 {
			zap.L().Debug("file is empty content", zap.String("n.NamespaceId", n.NamespaceId), zap.String("group", group), zap.String("outfile", outfile))
			return
		}
		// 需要写文件
		err = common.WriteFile(outfile, data)
		if err != nil {
			zap.L().Error("write file error", zap.Error(err))
			return
		}
	}
}

func (n *Node) Watch() {
	for _, j := range n.Jobs {
		for _, f := range j.outfiles {
			// 运行当前任务
			n.writeFile(f.DataId, f.Group, f.Outfile)
			err := func(curNamespace, curDId, curGroup, outfile string, tmpJ *Job) error {
				return n.Client.ListenConfig(vo.ConfigParam{
					DataId:  curDId,
					Group:   curGroup,
					Content: "",
					DatumId: "",
					Type:    "",
					OnChange: func(namespace string, group string, dataId string, data string) {
						n.Lock()
						defer n.Unlock()
						n.writeFile(dataId, group, outfile)
						tmpJ.Run()
						n.CheckDiff(group, dataId, data)
					},
				})
			}(n.NamespaceId, f.DataId, f.Group, f.Outfile, j)
			if err != nil {
				zap.L().Error("listent config error", zap.Error(err), zap.String("n.NamespaceId", n.NamespaceId), zap.String("group", f.Group), zap.String("dataId", f.DataId))
				return
			}
			zap.L().Debug("listent config", zap.String("n.NamespaceId", n.NamespaceId), zap.String("group", f.Group), zap.String("dataId", f.DataId))
		}
		j.Run()
	}
	for _, child := range n.Children {
		child.Watch()
	}
}

func (n *Node) UnWatch() {
	for _, job := range n.Jobs {
		for _, f := range job.outfiles {
			delUniqueKey(n.NamespaceId, f.Group, f.DataId)
			err := n.Client.CancelListenConfig(vo.ConfigParam{
				DataId: f.DataId,
				Group:  f.Group,
			})
			if err != nil {
				zap.L().Error("cancel listent config error", zap.Error(err), zap.String("n.NamespaceId", n.NamespaceId), zap.String("group", f.Group), zap.String("dataId", f.DataId), zap.String("outfile", f.Outfile))
			}
			zap.L().Debug("cancel listen config", zap.String("n.NamespaceId", n.NamespaceId), zap.String("group", f.Group), zap.String("dataId", f.DataId), zap.String("outfile", f.Outfile))
		}
	}
	for _, child := range n.Children {
		child.UnWatch()
	}
}

// 需要执行的任务
type Job struct {
	sync.Mutex
	Exec     string   // 可执行程序名字
	Params   []string // 参数
	Timeout  time.Duration
	outfiles []config.NacosJobFileConfig // 文件变动需要执行的任务
}

// 执行任务,
func (j *Job) Run() {
	j.Lock()
	defer j.Unlock()
	if len(j.Exec) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), j.Timeout)
		defer cancel()
		pCmd := exec.CommandContext(ctx, j.Exec, j.Params...)
		pCmd.Stdout = os.Stdout
		zap.L().Debug("exec run", zap.String("exec", j.Exec), zap.Any("params", j.Params))
		err := pCmd.Run()
		if err != nil {
			zap.L().Error("exec run fail", zap.Error(err), zap.String("exec", j.Exec), zap.Any("params", j.Params))
			return
		}
	}
}

func NewNode(conf *config.Config, group, dataId string) *Node {
	client, err := newClient(conf.NacosCfg)
	if err != nil {
		zap.L().Fatal("nacos客户端初始化出错", zap.Error(err))
	}
	root := &Node{Client: client, NamespaceId: conf.NacosCfg.NamespaceId, Group: group, DataId: dataId}
	jobs := make([]*Job, 0, len(conf.NacosJobs))
	for _, j := range conf.NacosJobs {
		job := &Job{
			Exec:     j.Exec,
			Params:   j.Params,
			outfiles: j.File,
			Timeout:  30 * time.Second,
		}
		jobs = append(jobs, job)
	}
	root.Jobs = jobs
	root.newChildren(conf)
	return root
}

func (n *Node) newChildren(conf *config.Config) {
	children := make([]*Node, 0, len(conf.NacosJobs))
	for _, job := range conf.NacosJobs {
		for _, f := range job.File {
			if isExists(n.NamespaceId, f.Group, f.DataId) {
				// 已经存在
				zap.L().Error("监听的文件已存在,忽略处理", zap.String("exists_namespace", n.NamespaceId), zap.String("exists_group", f.Group), zap.String("exists_dataId", f.DataId), zap.String("ref_namespace", n.NamespaceId), zap.String("ref_group", n.Group), zap.String("ref_dataId", n.DataId))
				continue
			}
			schema := findSchema(f.Outfile)
			if len(schema) > 0 {
				// 需要继续解析
				var newConf config.Config
				content, err := n.Client.GetConfig(vo.ConfigParam{
					DataId: f.DataId,
					Group:  f.Group,
				})
				if err != nil {
					log.Printf("获取的配置数据出错:%+v,data:%s\n", err, content)
					return
				}
				t, err := toml.Load(content)
				if err != nil {
					log.Printf("配置数据转toml出错:%+v\n", err)
					return
				}
				err = t.Unmarshal(&newConf)
				if err != nil {
					log.Printf("toml.Unmarshal配置数据转toml出错:%+v\n", err)
					return
				}
				node := NewNode(&newConf, f.Group, f.DataId)
				if node == nil {
					return
				}
				children = append(children, node)
			}
		}
	}
	n.Children = children
}
