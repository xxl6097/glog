package tool

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 声明一个日志记录器实例
var sugarLogger *zap.SugaredLogger

// https://www.jb51.net/jiaoben/353131j59.htm
// 初始化日志器
func GetLogger() *zap.Logger {
	encoder := getEncoder()       //获取编码器（定义日志格式）
	writeSyncer := getLogWriter() //获取输出目标（日志写到哪里）
	//创建一个核心日志处理器，相当于日志处理引擎
	core := zapcore.NewCore(encoder,
		zapcore.NewMultiWriteSyncer(
			writeSyncer,                              // 写入文件
			zapcore.AddSync(zapcore.Lock(os.Stdout)), // 打印到控制台
		),
		zapcore.InfoLevel) // 生产环境只打印Info及以上级别
	logger := zap.New(core,
		zap.AddCaller(),                       // 显示日志调用的文件和行号
		zap.AddStacktrace(zapcore.ErrorLevel)) //只有Error及以上级别才显示堆栈
	//从一个标准的zap.Logger实例创建并获取一个zap.SugaredLogger实例
	//sugarLogger = logger.Sugar()
	return logger
}

// 日志输出配置
func getLogWriter() zapcore.WriteSyncer {
	//创建日志文件轮转管理器实例
	//使用lumberjack库实现日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "./logs/app.log", //文件路径，不存在会自动创建
		MaxSize:    100,              //单个文件的最大大小（MB）
		MaxBackups: 10,               //保留的最大备份文件数
		MaxAge:     30,               //日志文件的最大保存天数
		Compress:   true,             //是否压缩备份文件
	}
	return zapcore.AddSync(lumberjackLogger)
}

// 使用JSON格式的编码器，并采用开发环境的默认配置，包括日志级别和时间戳等字段
func getEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	//修改时间格式
	config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(config)
	//return zapcore.NewJSONEncoder(config)
}
