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

var cfgFile string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "同步配置",
	Long:  `作为守护进程监听配置中心数据变动,同步配置`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
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

	runCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-ali-nacos.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".go-ali-nacos" (without extension).
		viper.AddConfigPath("./config")
		viper.AddConfigPath("./")
		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName(".ali-nacos")
	}

	viper.SetEnvPrefix("j00")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		zap.L().Debug("config file path", zap.String("file", viper.ConfigFileUsed()))
	} else {
		zap.L().Fatal("config file is not exists", zap.String("file", viper.ConfigFileUsed()))
	}
}
