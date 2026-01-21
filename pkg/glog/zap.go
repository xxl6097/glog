package glog

import (
	"io"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func create(fwriter io.Writer) *zap.Logger {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	switch env {
	case "prod", "production":
		return createProductionLogger(fwriter)
	case "dev", "development":
		return createDevelopmentLogger(fwriter)
	default:
		return createDevelopmentLogger(fwriter)
	}
}

func LoadLogConfig(fwriter *lumberjack.Logger) {
	logger := create(fwriter)
	defer logger.Sync()
}

func LoadLogDefault() {
	fwriter := &lumberjack.Logger{
		Filename:   "./logs/app.log", //文件路径，不存在会自动创建
		MaxSize:    100,              //单个文件的最大大小（MB）
		MaxBackups: 10,               //保留的最大备份文件数
		MaxAge:     30,               //日志文件的最大保存天数
		Compress:   true,             //是否压缩备份文件
	}
	logger := create(fwriter)
	defer logger.Sync()
}

func createProductionLogger(fwriter io.Writer) *zap.Logger {
	level := zap.InfoLevel //zap.InfoLevel zap.NewAtomicLevel()
	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	consoleEncoder := zapcore.NewJSONEncoder(config)
	// 文件编码器（无颜色）
	fileConfig := config
	fileConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 文件编码器（无颜色）
	fileEncoder := zapcore.NewConsoleEncoder(fileConfig)

	consoleWriter := zapcore.Lock(zapcore.AddSync(os.Stdout))
	fileWriter := zapcore.Lock(zapcore.AddSync(fwriter))

	// 用 NewTee 合并不同 Core（不同编码器+不同写入器）
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriter, level),
		zapcore.NewCore(fileEncoder, fileWriter, level),
	)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	zap.ReplaceGlobals(logger)
	return logger
}

func createDevelopmentLogger(logWriter io.Writer) *zap.Logger {
	level := zap.DebugLevel
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	// 文件编码器（无颜色）
	fileConfig := config
	fileConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 文件编码器（无颜色）
	fileEncoder := zapcore.NewConsoleEncoder(fileConfig)

	consoleWriter := zapcore.Lock(zapcore.AddSync(os.Stdout))
	fileWriter := zapcore.Lock(zapcore.AddSync(logWriter))

	// 用 NewTee 合并不同 Core（不同编码器+不同写入器）
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriter, level),
		zapcore.NewCore(fileEncoder, fileWriter, level),
	)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	zap.ReplaceGlobals(logger)
	return logger
}
