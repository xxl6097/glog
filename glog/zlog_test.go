package glog_test

import (
	"github.com/xxl6097go-log/glog"
	"testing"
)

func TestStdGLog(t *testing.T) {

	//测试 默认debug输出
	glog.Debug("zinx debug content1")
	glog.Debug("zinx debug content2")

	glog.Debugf(" zinx debug a = %d\n", 10)

	//设置log标记位，加上长文件名称 和 微秒 标记
	glog.ResetFlags(glog.BitDate | glog.BitLongFile | glog.BitLevel)
	glog.Info("zinx info content")

	//设置日志前缀，主要标记当前日志模块
	glog.SetPrefix("MODULE")
	glog.Error("zinx error content")

	//添加标记位
	glog.AddFlag(glog.BitShortFile | glog.BitTime)
	glog.Stack(" Zinx Stack! ")

	//设置日志写入文件
	glog.SetLogFile("./log", "testfile.log")
	glog.Debug("===> zinx debug content ~~666")
	glog.Debug("===> zinx debug content ~~888")
	glog.Error("===> zinx Error!!!! ~~~555~~~")

	//调试隔离级别
	glog.Debug("=================================>")
	//1.debug
	glog.SetLogLevel(glog.LogInfo)
	glog.Debug("===> 调试Debug：debug不应该出现")
	glog.Info("===> 调试Debug：info应该出现")
	glog.Warn("===> 调试Debug：warn应该出现")
	glog.Error("===> 调试Debug：error应该出现")
	//2.info
	glog.SetLogLevel(glog.LogWarn)
	glog.Debug("===> 调试Info：debug不应该出现")
	glog.Info("===> 调试Info：info不应该出现")
	glog.Warn("===> 调试Info：warn应该出现")
	glog.Error("===> 调试Info：error应该出现")
	//3.warn
	glog.SetLogLevel(glog.LogError)
	glog.Debug("===> 调试Warn：debug不应该出现")
	glog.Info("===> 调试Warn：info不应该出现")
	glog.Warn("===> 调试Warn：warn不应该出现")
	glog.Error("===> 调试Warn：error应该出现")
	//4.error
	glog.SetLogLevel(glog.LogPanic)
	glog.Debug("===> 调试Error：debug不应该出现")
	glog.Info("===> 调试Error：info不应该出现")
	glog.Warn("===> 调试Error：warn不应该出现")
	glog.Error("===> 调试Error：error不应该出现")
}

func TestZLogger(t *testing.T) {
}
