package main

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"os"
	"path"
	"runtime"
	"time"
)

func hook(data []byte) {
	//newData := data[2:]
	//fmt.Println(string(newData))
}

func init() {
	//glog.LogDefaultLogSetting("app.log")
	glog.SetLogFile("./log", "app.log")
	glog.Hook(hook)
	//开启日志保存文件
	//glog.LogSaveFile()
	//glog.SetNoHeader(true)
	//拦截日志
	//glog.Hook(hook)
	//glog.SetCons(true)
	//glog.SetNoHeader(true)
	//glog.SetNoColor(true)

	//glog.SetMaxSize(1 * 1024 * 1024)
	//glog.SetMaxAge(15)
	//glog.SetCons(true)
	//glog.SetNoHeader(true)
	//glog.SetNoColor(true)

	//glog.AddFlag(glog.BitMicroSeconds)
	//glog.AddFlag(glog.BitMilliseconds)
}
func getCallerInfo(skip int) (info string) {

	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {

		info = "runtime.Caller() failed"
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	fileName := path.Base(file) // Base函数返回路径的最后一个元素
	return fmt.Sprintf("FuncName:%s, file:%s, line:%d ", funcName, fileName, lineNo)
}

func testlog() {
	// 打印出getCallerInfo函数自身的信息
	//fmt.Println(getCallerInfo(0))
	// 打印出getCallerInfo函数的调用者的信息
	fmt.Println(getCallerInfo(1))
}

func test() {
	//&两个位都是1，则结果位为1
	//|两个位中至少有一个是1，则结果位为1
	//^两个位相同，则结果位为0
	var a byte = 3
	fmt.Println(a)
	a |= 0x08 //两个位相同，则结果位为0
	b := a >> 2
	fmt.Println(a, b)
	a &= 0x07
	fmt.Println(a)
}

func main() {
	//defer glog.GlobalRecover()
	glog.Println("服务安装成功!")
	test()
	glog.Println("hello glog...")
	time.Sleep(time.Millisecond * 100)
	//glog.SetLogFile("/usr/local/AATEST/logs", "normal.log")
	glog.Info("只有使用这个log打印才能记录日志哦", time.Now().Format("2006-01-02 15:04:05"))
	//glog.Flush()
	//testlog()
	//for {
	//	glog.Info("只有使用这个log打印才能记录日志哦", time.Now().Format("2006-01-02 15:04:05"))
	//	//glog.Info("Info。。。。")
	//	//glog.Error("Error。。。。")
	//	//glog.Warn("Warn。。。")
	//	//glog.Debug("Debug。。。")
	//	time.Sleep(5 * time.Second)
	//}
	glog.Error("wahahha")
	//panic("panic test...")
	//log.Fatalln("Fatalln test....")
	//glog.Stack("Stack test.....")
	for {
		time.Sleep(time.Second * 1)
		glog.Debug(time.Now().Format("2006-01-02 15:04:05"))
	}
	glog.CloseLog()
	//fmt.Scanln()
	os.Exit(0)
}
