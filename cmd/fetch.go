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

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "获取远程配置",
	Long:  `获取到的数据默认显示在控制台`,
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
		err := sync_nacos.Fetch(cfg)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().StringP("file", "f", "", "输出路径，默认使用 stdout")
}
