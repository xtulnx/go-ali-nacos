package sync_nacos

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go-ali-nacos/pkg/config"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// 写入本地磁盘
func WriteFile(file, content string) error {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		fileBase := filepath.Dir(file)
		err := os.MkdirAll(fileBase, os.ModePerm)
		if err != nil {
			zap.L().Error("创建目录出错", zap.String("目录", fileBase), zap.Error(err))
			return err
		}
	} else if err != nil {
		zap.L().Error("获取文件数据出错", zap.String("文件", file), zap.Error(err))
		return err
	}
	err = os.WriteFile(file, []byte(content), os.ModePerm)
	if err != nil {
		zap.L().Error("写文件出错", zap.Error(err), zap.String("文件", file))
		return err
	}
	return nil
}

func getClient(cfg config.DirectConfig) (config_client.IConfigClient, error) {
	if cfg.NacosCfg == nil {
		return nil, fmt.Errorf("缺少必要参数")
	}
	if cfg.NacosCfg.NamespaceId == "" {
		return nil, fmt.Errorf("缺少 namespaceId ")
	}
	if cfg.DataId == "" || cfg.Group == "" {
		return nil, fmt.Errorf("缺少目标资源")
	}
	cfg.NacosCfg.LogLevel = "error"
	client, err := NewClient(*cfg.NacosCfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func Fetch(c config.DirectConfig) error {
	client, err := getClient(c)
	if err != nil {
		zap.L().Fatal("初始化client出错", zap.Error(err))
	}
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: c.DataId,
		Group:  c.Group,
	})
	if err != nil {
		zap.L().Fatal("获取配置数据出错", zap.Error(err), zap.Any("config", c))
	}
	if len(c.File) > 0 {
		err := ioutil.WriteFile(c.File, []byte(content), os.ModePerm)
		err = WriteFile(c.File, content)
		if err != nil {
			zap.L().Fatal("文件写入失败", zap.Error(err), zap.String("file", c.File))
		}
	} else {
		fmt.Fprintln(os.Stdout, content)
	}
	return nil
}

func Push(c config.DirectConfig) error {
	client, err := getClient(c)
	if err != nil {
		zap.L().Fatal("初始化client出错", zap.Error(err))
	}
	var content string
	if len(c.File) > 0 {
		contentByte, err := ioutil.ReadFile(c.File)
		if err != nil {
			zap.L().Fatal("读取配置文件数据出错", zap.String("file", c.File), zap.Error(err))
		}
		content = string(contentByte)
	} else {
		// 检查管道是否有数据
		fileInfo, err := os.Stdin.Stat()
		if err != nil {
			zap.L().Fatal("读取管道数据出错", zap.Error(err))
		}
		if fileInfo.Mode()&os.ModeCharDevice == 0 {
			// 管道有数据
			contentByte, err := io.ReadAll(os.Stdin)
			if err != nil {
				zap.L().Fatal("读取管道数据出错", zap.Error(err))
			}
			content = string(contentByte)
		} else {
			zap.L().Fatal("请指定push的数据!")
		}
	}
	_, err = client.PublishConfig(vo.ConfigParam{
		DataId:  c.DataId,
		Group:   c.Group,
		Content: content,
	})
	if err != nil {
		zap.L().Fatal("push配置数据出错", zap.String("dataId", c.DataId), zap.String("group", c.Group), zap.Error(err))
	}
	return nil
}
