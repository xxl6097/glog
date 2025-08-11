package main

import (
	"github.com/xxl6097/glog/glog"
	"time"
)

func main() {
	defer glog.GlobalRecover()
	glog.Debug("AppHome", glog.AppHome())
	glog.Debug("AppName", glog.AppName())
	glog.Debug("test is now...")
	time.Sleep(5 * time.Second)
	glog.Debug("test is now...1")
	panic("err.ooiefrer")
}
