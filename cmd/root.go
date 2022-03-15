/*
 * Copyright 2008-2022 xtulnx.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go-ali-nacos/pkg/config"
	"go-ali-nacos/pkg/sync_nacos"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var cfgFile string
var cfgQuiet bool

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "同步 nacos 配置",
	Long:  `作为守护进程监听配置中心数据变动,同步配置`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd:   true,
		DisableNoDescFlag:   false,
		DisableDescriptions: false,
		HiddenDefaultCmd:    false,
	},
	Run: func(cmd *cobra.Command, args []string) {
		bindDirect(cmd)
		type _tConfig struct {
			config.Config `mapstructure:",squash"`

			Group  string `json:"group" toml:"group" mapstructure:"group"`    // 资源组
			DataId string `json:"dataId" toml:"dataId" mapstructure:"dataId"` // 资源 ID
		}
		var cfg _tConfig
		if err := viper.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
			c.WeaklyTypedInput = true
		}); err != nil {
			zap.L().Fatal("获取本地配置文件Unmarshal出错", zap.Error(err))
		}
		if cfgQuiet && cfg.NacosCfg.LogLevel == "" {
			cfg.NacosCfg.LogLevel = "error"
		}
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sync_nacos.Main(ctx, cfg.Config, cfg.Group, cfg.DataId)
			if err != nil {
				log.Fatal(err)
			}
			cancel()
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-interrupt:
				log.Println("interrupt")
				cancel()
				break
			case <-ctx.Done():
				log.Println("cancel")
				break
			}
		}()
		wg.Wait()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolP("help", "h", false, "查看帮助")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件 (默认查找 .go-ali-nacos.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&cfgQuiet, "quiet", "q", false, "安静模式")
	initDirect(rootCmd, rootCmd.PersistentFlags())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath("./config")
		viper.AddConfigPath("./")
		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName(".ali-nacos")
	}
	viper.SetEnvPrefix("j00")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		if !cfgQuiet {
			log.Printf("config file path: %s", viper.ConfigFileUsed())
		}
	}
}

// 定义命令行参数
func initDirect(cmd *cobra.Command, flags *pflag.FlagSet) {
	flags.StringP("endpoint", "e", "", "需要连接的远程配置地址如: acm.aliyun.com (公网)")
	flags.StringP("namespaceId", "n", "", "远程配置的命名空间")
	flags.String("ak", "", "远程配置连接参数,accessKey")
	flags.String("sk", "", "远程配置连接参数,secretKey")
	flags.StringP("dataId", "d", "", "数据id")
	flags.StringP("group", "g", "", "数据分组")

	//flags.StringP("username", "u", "", "需要连接的远程配置地址的用户名")
	//flags.StringP("password", "p", "", "需要连接的远程配置地址的密码")
}

// 关联参数
func bindDirect(cmd *cobra.Command) {
	viper.SetDefault("nacos.loglevel","")
	_ = viper.BindPFlag("nacos.endpoint", cmd.Flags().Lookup("endpoint"))
	_ = viper.BindPFlag("nacos.namespaceId", cmd.Flags().Lookup("namespaceId"))
	_ = viper.BindPFlag("nacos.accessKey", cmd.Flags().Lookup("ak"))
	_ = viper.BindPFlag("nacos.secretKey", cmd.Flags().Lookup("sk"))
	_ = viper.BindPFlag("nacos.quiet", cmd.Flags().Lookup("sk"))
	_ = viper.BindPFlag("dataId", cmd.Flags().Lookup("dataId"))
	_ = viper.BindPFlag("group", cmd.Flags().Lookup("group"))
}
