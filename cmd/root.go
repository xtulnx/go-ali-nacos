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
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"go-ali-nacos/pkg/config"
	"go-ali-nacos/pkg/sync_nacos"
	"golang.org/x/net/context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "go-ali-nacos",
	Short: "同步 nacos 配置",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var cfg config.Config
		err = viper.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
			c.WeaklyTypedInput = true
		})
		if err != nil {
			log.Fatal(err)
		}

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = sync_nacos.Main(ctx, cfg)
			if err != nil {
				log.Fatal(err)
			}
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-ali-nacos.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
