/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go-ali-nacos/pkg/config"
	"go-ali-nacos/pkg/sync_nacos"
	"log"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "推送配置",
	Long:  `推送配置内容到远端指定配置项。
注意，
  内容为空时将删除远端配置`,
	Run: func(cmd *cobra.Command, args []string) {
		bindDirect(cmd)
		_ = viper.BindPFlag("file", cmd.Flags().Lookup("file"))
		var cfg config.DirectConfig
		if err := viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
			dc.WeaklyTypedInput = true
		}); err != nil {
			fmt.Println(err)
		}
		if cfgQuiet && cfg.NacosCfg != nil && cfg.NacosCfg.LogLevel == "" {
			cfg.NacosCfg.LogLevel = "error"
		}
		err := sync_nacos.Push(cfg)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringP("file", "f", "", "资源路径，默认使用 stdin")
}
