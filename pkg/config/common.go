package config

import (
	"go-ali-nacos/pkg/consts"

	"github.com/spf13/cobra"
)

type CommonParams struct {
	Endpoint    string
	NamespaceId string
	AccessKey   string
	SecretKey   string
	DataId      string
	Group       string
	File        string
	Username    string
	Password    string
}

// fetch push 公共参数获取
func GetCommonParams(cmd *cobra.Command) CommonParams {
	return CommonParams{
		Endpoint:    cmd.Flag(consts.FlagNacosEndpoint).Value.String(),
		Username:    cmd.Flag(consts.FlagNacosUsername).Value.String(),
		Password:    cmd.Flag(consts.FlagNacosPassword).Value.String(),
		NamespaceId: cmd.Flag(consts.FlagNacosNamespaceId).Value.String(),
		AccessKey:   cmd.Flag(consts.FlagNacosAk).Value.String(),
		SecretKey:   cmd.Flag(consts.FlagNacosSk).Value.String(),
		DataId:      cmd.Flag(consts.FlagNacosDataId).Value.String(),
		Group:       cmd.Flag(consts.FlagNacosGroup).Value.String(),
		File:        cmd.Flag(consts.FlagTargetFile).Value.String(),
	}
}
