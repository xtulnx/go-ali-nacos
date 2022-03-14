/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go-ali-nacos/pkg/config"
	"go-ali-nacos/pkg/sync_nacos"
	"log"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push配置到远程配置中心",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		bindDirect(cmd)
		var cfg config.DirectConfig
		if err := viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
			dc.WeaklyTypedInput = true
		}); err != nil {
			fmt.Println(err)
		}
		err := sync_nacos.Push(cfg)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func pushFetchInit() (*config.DirectConfig, config_client.IConfigClient) {
	var cfg config.DirectConfig
	if err := viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.WeaklyTypedInput = true
	}); err != nil {
		fmt.Println(err)
	}
	if cfg.NacosCfg == nil {
		log.Fatal("缺少必要参数")
	}
	if cfg.NacosCfg.NamespaceId == "" {
		log.Fatal("缺少 namespaceId ")
	}
	if cfg.DataId == "" || cfg.Group == "" {
		log.Fatalf("缺少目标资源")
	}

	client, err := sync_nacos.NewClient(*cfg.NacosCfg)
	if err != nil {
		log.Fatal("初始化失败", err)
	}
	return &cfg, client
}

func init() {
	rootCmd.AddCommand(pushCmd)
	initDirect(pushCmd)
}

// 定义命令行参数
func initDirect(cmd *cobra.Command) {
	cmd.Flags().StringP("endpoint", "e", "", "需要连接的远程配置地址如:acm.aliyun.com")
	cmd.Flags().StringP("namespaceId", "n", "", "远程配置的命名空间")
	cmd.Flags().String("ak", "", "远程配置连接参数,accessKey")
	cmd.Flags().String("sk", "", "远程配置连接参数,secretKey")

	//cmd.Flags().StringP("username", "u", "", "需要连接的远程配置地址的用户名")
	//cmd.Flags().StringP("password", "p", "", "需要连接的远程配置地址的密码")

	cmd.Flags().StringP("dataId", "d", "", "数据id")
	cmd.Flags().StringP("group", "g", "", "数据分组")
	cmd.Flags().StringP("file", "f", "", "配置文件路径,如果不指定,通过管道获取")
}

// 关联参数
func bindDirect(cmd *cobra.Command) {
	_ = viper.BindPFlag("nacos.endpoint", cmd.Flags().Lookup("endpoint"))
	_ = viper.BindPFlag("nacos.namespaceId", cmd.Flags().Lookup("namespaceId"))
	_ = viper.BindPFlag("nacos.accessKey", cmd.Flags().Lookup("ak"))
	_ = viper.BindPFlag("nacos.secretKey", cmd.Flags().Lookup("sk"))
	_ = viper.BindPFlag("dataId", cmd.Flags().Lookup("dataId"))
	_ = viper.BindPFlag("group", cmd.Flags().Lookup("group"))
	_ = viper.BindPFlag("file", cmd.Flags().Lookup("file"))
}
