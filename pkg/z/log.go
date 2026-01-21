package z

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func LoadDefaultLogger() {
	initZapLogger(&LogConfig{
		Level: "debug",
	})
}

func LoadLogger(fn func(conf *LogConfig)) {
	cfg := &LogConfig{
		Level:            "debug",
		Path:             "./logs/app.log",
		ErrorPath:        "./logs/error.log",
		MaxSize:          100,
		MaxBackups:       30,
		MaxAge:           7,
		Compress:         true,
		SeparateErrorLog: true,
	}
	if fn != nil {
		fn(cfg)
	}
	initZapLogger(cfg)
}

// initZapLogger 初始化日志
func initZapLogger(cfg *LogConfig) {
	// 设置日志级别
	var level = getZapLevel(cfg.Level)
	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 彩色编码级别
		EncodeTime:     customTimeEncoder,                // 自定义时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 文件编码器（无颜色）
	var fileEncoder zapcore.Encoder
	fileConfig := encoderConfig
	fileConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	env := strings.ToLower(os.Getenv("APP_ENV"))
	switch env {
	case "prod", "production":
		fileEncoder = zapcore.NewJSONEncoder(fileConfig)
	default:
		fileEncoder = zapcore.NewConsoleEncoder(fileConfig)
	}

	// 创建核心列表
	var cores []zapcore.Core
	consoleWriter := zapcore.Lock(zapcore.AddSync(os.Stdout))
	mainCore := zapcore.NewCore(consoleEncoder, consoleWriter, level)
	cores = append(cores, mainCore)

	if cfg.Path != "" {
		fileWriter := zapcore.Lock(zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}))
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)
		cores = append(cores, fileCore)
	}
	if cfg.SeparateErrorLog && cfg.ErrorPath != "" {
		fileWriter := zapcore.Lock(zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.ErrorPath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}))
		core := zapcore.NewCore(fileEncoder, fileWriter, zap.ErrorLevel)
		cores = append(cores, core)
	}

	// 创建 tee 核心（多路复用）
	core := zapcore.NewTee(cores...)
	// 创建 Logger
	//ZapLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0), zap.AddStacktrace(zap.ErrorLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	//SugaredLogger = ZapLogger.Sugar()
	logger.Info("logger init success")
	zap.ReplaceGlobals(logger)
}

// getZapLevel 获取 Zap 日志级别
func getZapLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// customTimeEncoder 自定义时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// isTerminal 判断是否为终端
func isTerminal1() bool {
	fileInfo, _ := os.Stdout.Stat()
	result := fileInfo.Mode() & os.ModeCharDevice
	return result != 0
}

func Log() *zap.Logger {
	return zap.L()
}
