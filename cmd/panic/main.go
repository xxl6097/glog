package main

import (
	"github.com/xxl6097/glog/glog"
	"time"
)

func main() {
	glog.LogDefaultLogSetting("aaa.log")
	//defer glog.GlobalRecover()
	glog.Debug("AppHome", glog.AppHome())
	glog.Debug("AppName", glog.AppName())
	glog.Debug("test is now...")
	glog.Debug("bbbbbbbbbb")
	glog.LogToFile("aaa.log")
	time.Sleep(5 * time.Second)
	glog.Debug("test is now...1")
	//panic("err.ooiefrer")
}
