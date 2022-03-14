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

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push配置到远程配置中心",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		once.Push(config.GetCommonParams(cmd))
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringP(consts.FlagNacosEndpoint, "e", "", "需要连接的远程配置地址如:acm.aliyun.com")
	pushCmd.Flags().StringP(consts.FlagNacosUsername, "u", "", "需要连接的远程配置地址的用户名")
	pushCmd.Flags().StringP(consts.FlagNacosPassword, "p", "", "需要连接的远程配置地址的密码")
	pushCmd.Flags().StringP(consts.FlagNacosAk, "", "", "远程配置连接参数,accessKey")
	pushCmd.Flags().StringP(consts.FlagNacosSk, "", "", "远程配置连接参数,secretKey")
	pushCmd.Flags().StringP(consts.FlagNacosNamespaceId, "n", "", "远程配置的命名空间")
	pushCmd.Flags().StringP(consts.FlagNacosDataId, "d", "", "数据id")
	pushCmd.Flags().StringP(consts.FlagNacosGroup, "g", "", "数据分组")
	pushCmd.Flags().StringP(consts.FlagTargetFile, "f", "", "配置文件路径,如果不指定,通过管道获取")
	if err := pushCmd.MarkFlagRequired(consts.FlagNacosEndpoint); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := pushCmd.MarkFlagRequired(consts.FlagNacosAk); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := pushCmd.MarkFlagRequired(consts.FlagNacosSk); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := pushCmd.MarkFlagRequired(consts.FlagNacosNamespaceId); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := pushCmd.MarkFlagRequired(consts.FlagNacosGroup); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
	if err := pushCmd.MarkFlagRequired(consts.FlagNacosDataId); err != nil {
		log.Fatalf("flag标记required出错:%v", err)
	}
}
