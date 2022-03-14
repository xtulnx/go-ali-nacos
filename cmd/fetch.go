/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"

	"go-ali-nacos/pkg/config"
	"go-ali-nacos/pkg/consts"
	"go-ali-nacos/pkg/once"

	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "获取远程配置",
	Long:  `获取远程配置,获取到的数据直接显示在控制台`,
	Run: func(cmd *cobra.Command, args []string) {
		once.Fetch(config.GetCommonParams(cmd))
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringP(consts.FlagNacosEndpoint, "e", "", "需要连接的远程配置地址如:acm.aliyun.com")
	fetchCmd.Flags().StringP(consts.FlagNacosUsername, "u", "", "需要连接的远程配置地址的用户名")
	fetchCmd.Flags().StringP(consts.FlagNacosPassword, "p", "", "需要连接的远程配置地址的密码")
	fetchCmd.Flags().StringP(consts.FlagNacosAk, "", "", "远程配置连接参数,accessKey")
	fetchCmd.Flags().StringP(consts.FlagNacosSk, "", "", "远程配置连接参数,secretKey")
	fetchCmd.Flags().StringP(consts.FlagNacosNamespaceId, "n", "", "远程配置的命名空间")
	fetchCmd.Flags().StringP(consts.FlagNacosDataId, "d", "", "数据id")
	fetchCmd.Flags().StringP(consts.FlagNacosGroup, "g", "", "数据分组")
	fetchCmd.Flags().StringP(consts.FlagTargetFile, "f", "", "配置文件写入路径,如果不指定,输出到console")
	if err := fetchCmd.MarkFlagRequired(consts.FlagNacosEndpoint); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := fetchCmd.MarkFlagRequired(consts.FlagNacosAk); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := fetchCmd.MarkFlagRequired(consts.FlagNacosSk); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := fetchCmd.MarkFlagRequired(consts.FlagNacosNamespaceId); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := fetchCmd.MarkFlagRequired(consts.FlagNacosGroup); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := fetchCmd.MarkFlagRequired(consts.FlagNacosDataId); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
}
