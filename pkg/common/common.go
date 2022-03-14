package common

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// 写入本地磁盘
func WriteFile(file, content string) error {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		fileBase := filepath.Dir(file)
		err := os.MkdirAll(fileBase, os.ModePerm)
		if err != nil {
			zap.L().Error("创建目录出错", zap.String("目录", fileBase), zap.Error(err))
			return err
		}
	} else if err != nil {
		zap.L().Error("获取文件数据出错", zap.String("文件", file), zap.Error(err))
		return err
	}
	err = os.WriteFile(file, []byte(content), os.ModePerm)
	if err != nil {
		zap.L().Error("写文件出错", zap.Error(err), zap.String("文件", file))
		return err
	}
	return nil
}
