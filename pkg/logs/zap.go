package logs

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitZapLogger(logFolder, logName string, minLevel zapcore.Level) {
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		result := lvl >= minLevel && lvl >= zapcore.ErrorLevel
		return result
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		result := lvl >= minLevel && lvl < zapcore.ErrorLevel
		return result
	})

	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeTime = customTimeEncoder
	fileEncoder := zapcore.NewConsoleEncoder(cfg)
	consoleEncoder := zapcore.NewConsoleEncoder(cfg)
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	logCore := []zapcore.Core{
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	}

	if logFolder != "" {
		currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			panic(err)
		}

		hook := lumberjack.Logger{
			Filename:   filepath.Join(currentPath, logFolder, logName+".log"), // 日志文件路径
			MaxSize:    128,                                                   // 最大日志大小（Mb级别）
			MaxBackups: 30,                                                    // 最多保留30个备份
			MaxAge:     7,                                                     // days
			Compress:   true,                                                  // 是否压缩 disabled by default
			LocalTime:  true,
		}
		logCore = append(logCore,
			zapcore.NewCore(fileEncoder, zapcore.AddSync(&hook), highPriority),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(&hook), lowPriority),
		)
	}
	core := zapcore.NewTee(logCore...)
	logger := zap.New(core, zap.AddStacktrace(zap.WarnLevel))
	zap.ReplaceGlobals(logger)
}

func customTimeEncoder(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(time.Format("2006-01-02 15:04:05.000000"))
}
