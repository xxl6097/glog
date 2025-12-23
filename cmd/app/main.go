package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/xxl6097/glog/glog"
)

func hook(data []byte) {
	//newData := data[2:]
	//fmt.Println(string(newData))
}
func initLog() {
	bindir, err := os.Executable()
	var isSrvApp bool
	if err != nil {
		glog.LogDefaultLogSetting("app.log")
	} else {
		isSrvApp = strings.HasPrefix(strings.ToLower(bindir), strings.ToLower("/Users/uuxia/Desktop/work/code/github/golang/glog"))
		if isSrvApp {
			glog.LogDefaultLogSetting("app.log")
		} else {
			glog.SetLogFile(filepath.Dir(bindir), fmt.Sprintf("install-%s.log", filepath.Base(bindir)))
		}
	}
}
func init() {
	//initLog()
	//glog.LogDefaultLogSetting("app.log")
	glog.SetLogFile("./log", "app.log")
	//glog.Hook(hook)
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

func testSaveLastLog() {
	logFile, _ := os.OpenFile("./test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer logFile.Close()

	if r := recover(); r != nil {
		stack := debug.Stack()
		//logLib.Fatal("PANIC",
		//	zap.Any("error", r),
		//	zap.ByteString("stack", stack),
		//	zap.String("time", time.Now().Format(time.RFC3339Nano))
		//)
		glog.Error("stack:", string(stack))
		//Error("err:", r)
		//Error(Stack(r))
		//Stack(r)
		go func() {
			io.Copy(logFile, os.Stderr)
		}()
	}
	// 将标准错误重定向到日志文件

}

// 记录程序退出日志
func logExit(message string, err error) {
	logFile, err := os.OpenFile("app_exit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("无法打开日志文件: %v", err)
		return
	}
	defer logFile.Close()

	// 创建日志记录器
	logger := log.New(logFile, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	// 记录退出时间
	logger.Printf("程序退出时间: %s", time.Now().Format(time.RFC3339))
	logger.Printf("退出信息: %s", message)

	// 记录堆栈跟踪
	logger.Println("堆栈跟踪:")
	logger.Println(string(debug.Stack()))
	logger.Println("----------------------------------------")
}

func testLog() {
	glog.Println("testLog...")
	//panic("panic test...")
	go func() {
		panic("panic test...")
	}()
}

func testWhile() {
	for {
		glog.Println("testWhile...")
		time.Sleep(time.Second)
	}
}
func main() {
	//defer glog.GlobalRecover()
	//defer testSaveLastLog()
	defer func() {
		if r := recover(); r != nil {
			// 记录详细的错误信息和堆栈跟踪
			log.Printf("程序异常退出: %v\n堆栈信息: %s", r, string(debug.Stack()))
			// 这里可以添加资源清理代码
			// 然后选择退出或继续（在守护服务中可能不立即退出）
			// os.Exit(1)
		}
	}()
	glog.Println("服务安装成功!")
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
	//testLog()
	//for {
	//	time.Sleep(time.Second * 1)
	//	glog.Debug(time.Now().Format("2006-01-02 15:04:05"))
	//}
	glog.CloseLog()
	//fmt.Scanln()
	testWhile()
}
