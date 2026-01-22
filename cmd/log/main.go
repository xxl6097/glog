package main

import (
	"errors"
	"time"

	"github.com/xxl6097/glog/pkg/z"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func test() {
	// 1. 创建 lumberjack logger
	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/server.log",
		MaxSize:    10, // 10 MB
		MaxBackups: 3,  // 最多3个备份
		MaxAge:     7,  // 保留7天
		Compress:   true,
	})

	// 2. 设置编码器
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// 3. 创建 core
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)

	// 4. 创建 logger
	logger := zap.New(core)
	defer logger.Sync()

	// 5. 打印日志
	logger.Info("这是一个日志条目", zap.String("user", "tinG"))
}

func test001() {
	// 初始化zap的生产环境配置（默认是JSON格式）
	logger, _ := zap.NewProduction()
	// 记得关闭日志器，避免内存泄漏
	defer logger.Sync()

	// 打印不同级别的日志
	logger.Debug("这是调试日志，开发时用") // 生产环境默认不打印Debug
	logger.Info("服务启动成功",
		zap.String("服务名", "user-service"),
		zap.Int("端口", 8080),
		zap.String("启动时间", time.Now().Format("2006-01-02 15:04:05")),
	)
	logger.Warn("配置文件未找到，使用默认配置",
		zap.String("配置路径", "./config.yaml"),
	)
	err := errors.New("连接数据库出错")
	logger.Error("数据库连接失败",
		zap.String("地址", "127.0.0.1:3306"),
		zap.Error(err), // 直接把错误信息带进去
	)
}
func init() {
	//zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
	//zap.ReplaceGlobals(zap.Must(internal.Logger.Create(), nil))
	//internal.InitLogger()
	//glog.LoadLogDefault()
	z.LoadLogger(func(cfg *z.LogConfig) {
		cfg.Level = "debug"
	})

}
func main() {
	//logger := internal.Logger.Create()
	//defer logger.Sync()
	//logger.Info("info...")
	//zlog.ZapLogger.Info("z============")
	z.L().Debug("测试", zap.String("username", "tinG"))
}
