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
	var encoder zapcore.Encoder
	env := strings.ToLower(os.Getenv("APP_ENV"))
	switch env {
	case "prod", "production":
		encoder = zapcore.NewJSONEncoder(jsonEncoderConfig())
	default:
		encoder = zapcore.NewConsoleEncoder(consoleEncoderConfig(true))
	}
	// 创建核心列表
	var cores []zapcore.Core
	consoleWriter := zapcore.Lock(zapcore.AddSync(os.Stdout))
	mainCore := zapcore.NewCore(encoder, consoleWriter, level)
	cores = append(cores, mainCore)
	if cfg.Path != "" {
		fileCore := createNormalCore(cfg, encoder, level)
		cores = append(cores, fileCore)
	}
	if cfg.SeparateErrorLog && cfg.ErrorPath != "" {
		errorCore := createErrorCore(cfg, encoder, zap.ErrorLevel)
		cores = append(cores, errorCore)
	}

	// 创建 tee 核心（多路复用）
	core := zapcore.NewTee(cores...)
	// 创建 Logger
	//ZapLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0), zap.AddStacktrace(zap.ErrorLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	//SugaredLogger = ZapLogger.Sugar()
	//logger.Info("logger init success")
	zap.ReplaceGlobals(logger)
}

func createNormalCore(cfg *LogConfig, encoder zapcore.Encoder, level zapcore.Level) zapcore.Core {
	fileWriter := zapcore.Lock(zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}))
	return zapcore.NewCore(encoder, fileWriter, level)
}

func createErrorCore(cfg *LogConfig, encoder zapcore.Encoder, level zapcore.Level) zapcore.Core {
	fileWriter := zapcore.Lock(zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.ErrorPath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}))
	return zapcore.NewCore(encoder, fileWriter, level)
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

// consoleEncoderConfig 控制台编码器配置
func consoleEncoderConfig(color bool) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:   "time",
		LevelKey:  "level",
		NameKey:   "logger",
		CallerKey: "caller",
		//FunctionKey:    "function", // 添加函数名
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeTime:     customTimeEncoder,
	}

	if color {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	return encoderConfig
}

// jsonEncoderConfig JSON编码器配置
func jsonEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写level
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601格式
		EncodeDuration: zapcore.MillisDurationEncoder, // 毫秒
		EncodeCaller:   zapcore.ShortCallerEncoder,    // 短路径
	}
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
