/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"go-ali-nacos/pkg/config"
	"go-ali-nacos/pkg/sync_nacos"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)


// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "同步配置",
	Long:  `作为守护进程监听配置中心数据变动,同步配置`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var cfg config.Config
		err = viper.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
			c.WeaklyTypedInput = true
		})
		if err != nil {
			zap.L().Fatal("获取本地配置文件Unmarshal出错", zap.Error(err))
		}
		root := sync_nacos.NewNode(&cfg, "", "")
		root.Watch()
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		<-interrupt
		root.UnWatch()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
