package once

import (
	"fmt"
	"io"
	"os"

	"go-ali-nacos/pkg/common"
	"go-ali-nacos/pkg/config"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go.uber.org/zap"
)

func Fetch(c config.CommonParams) {
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
		err = common.WriteFile(c.File, content)
		if err != nil {
			zap.L().Fatal("文件写入失败", zap.Error(err), zap.String("file", c.File))
		}
	} else {
		fmt.Fprintln(os.Stdout, content)
	}
}

func Push(c config.CommonParams) {
	client, err := getClient(c)
	if err != nil {
		zap.L().Fatal("初始化client出错", zap.Error(err))
	}
	content := ""
	if len(c.File) > 0 {
		contentByte, err := os.ReadFile(c.File)
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
}

func getClient(c config.CommonParams) (config_client.IConfigClient, error) {
	clientConfig := &constant.ClientConfig{
		Endpoint:             c.Endpoint + ":8080",
		NamespaceId:          c.NamespaceId,
		AccessKey:            c.AccessKey,
		SecretKey:            c.SecretKey,
		TimeoutMs:            5 * 1000,
		OpenKMS:              false,
		NotLoadCacheAtStart:  true,
		UpdateCacheWhenEmpty: true,
		Username:             c.Username,
		Password:             c.Password,
	}

	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig: clientConfig,
		},
	)
	return client, err
}
