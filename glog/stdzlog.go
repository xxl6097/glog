package glog

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/*
	A global Log handle is provided by default for external use, which can be called directly through the API series.
	The global log object is StdGLog.
	Note: The methods in this file do not support customization and cannot replace the log recording mode.

	If you need a custom logger, please use the following methods:
	zlog.SetLogger(yourLogger)
	zlog.Ins().InfoF() and other methods.

   全局默认提供一个Log对外句柄，可以直接使用API系列调用
   全局日志对象 StdGLog
   注意：本文件方法不支持自定义，无法替换日志记录模式，如果需要自定义Logger:

   请使用如下方法:
   zlog.SetLogger(yourLogger)
   zlog.Ins().InfoF()等方法
*/

var StdGLog = NewGLog(os.Stdout, "", BitDefault)

//var logDir string

func Flags() int {
	return StdGLog.Flags()
}

// ResetFlags sets the flags of StdGLog
func ResetFlags(flag int) {
	StdGLog.ResetFlags(flag)
}

// 设置打印时间戳到毫秒
func AddFlag(flag int) {
	StdGLog.AddFlag(flag)
}

func SetPrefix(prefix string) {
	StdGLog.SetPrefix(prefix)
}

func SetLogFile(fileDir string, fileName string) {
	StdGLog.SetLogFile(fileDir, fileName)
}

func LogSaveFile() {
	StdGLog.SetLogFile("./", "app.log")
}

func GetTempDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"))
	default:
		return filepath.Join(os.TempDir())
	}
}

// GetCrossPlatformDataDir
// 临时日志	C:\Users\xxx\AppData\Local\Temp	/tmp	os.TempDir()
// 用户级日志	C:\Users\xxx\logs	/home/username/logs	os.UserHomeDir() + 拼接目录
// 系统级日志	C:\ProgramData\app\logs	/var/log/app	固定路径 + filepath.Join()
func GetCrossPlatformDataDir(args ...string) string {
	//if home, err := os.UserHomeDir(); err == nil {
	//	logDir = filepath.Join(home, appName)
	//} else {
	//	logDir = filepath.Join(os.TempDir(), appName)
	//}

	var dirs []string
	dirs = append(dirs, GetTempDir())
	dirs = append(dirs, GetAppName())
	if len(args) > 0 {
		dirs = append(dirs, args...)
	}
	dir := filepath.Join(dirs...)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	return dir
}

func GetAppName() string {
	var appName string
	if exePath, err := os.Executable(); err == nil {
		appName = filepath.Base(exePath)
	} else {
		appName = filepath.Base(os.Args[0])
	}
	// 处理通用扩展名
	if ext := filepath.Ext(appName); ext != "" {
		appName = strings.TrimSuffix(appName, ext)
	}
	if strings.Contains(appName, "_") {
		arr := strings.Split(appName, "_")
		if arr != nil && len(arr) > 0 {
			appName = arr[0]
		}
	}
	return appName
}

func LogDefaultLogSetting(logFileName string) {
	logDir := GetCrossPlatformDataDir(GetAppName())
	StdGLog.SetLogFile(logDir, logFileName)
	SetCons(true)               //需要控制台打印
	SetMaxAge(7)                //默认保存7天
	SetMaxSize(1 * 1024 * 1024) //1MB
	AddFlag(BitMilliseconds)
}

// Hook hook log
func Hook(f func([]byte)) {
	StdGLog.SetLogHook(f)
}

// SetMaxAge 最大保留天数
func SetMaxAge(ma int) {
	StdGLog.SetMaxAge(ma)
}

// SetMaxSize 单个日志最大容量 单位：字节
func SetMaxSize(ms int64) {
	StdGLog.SetMaxSize(ms)
}

// SetCons 同时输出控制台
func SetCons(b bool) {
	StdGLog.SetCons(b)
}

// SetNoHeader
// 头指的时间，行号等信息
func SetNoHeader(b bool) {
	StdGLog.SetNoHeader(b)
}
func SetNoColor(b bool) {
	StdGLog.SetNoColor(b)
}

func SetLogLevel(logLevel int) {
	StdGLog.SetLogLevel(logLevel)
}

func Debugf(format string, v ...interface{}) {
	StdGLog.Debugf(format, v...)
}

func DebugfNoCon(format string, v ...interface{}) {
	StdGLog.DebugfNoCon(format, v...)
}

func Debug(v ...interface{}) {
	StdGLog.Debug(v...)
}
func DebugNoCon(v ...interface{}) {
	StdGLog.DebugNoCon(v...)
}

func Println(a ...any) {
	StdGLog.Info(a...)
}
func PrintlnNoCon(a ...any) {
	StdGLog.InfoNoCon(a...)
}

func Flush() error {
	return StdGLog.Flush()
}

func Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

func Print(a ...any) {
	StdGLog.Info(a...)
}
func PrintNoCon(a ...any) {
	StdGLog.InfoNoCon(a...)
}

func Printf(format string, a ...any) {
	StdGLog.Infof(format, a...)
}
func PrintfNoCon(format string, a ...any) {
	StdGLog.InfofNoCon(format, a...)
}

func Infof(format string, v ...interface{}) {
	StdGLog.Infof(format, v...)
}

func InfofNoCon(format string, v ...interface{}) {
	StdGLog.InfofNoCon(format, v...)
}

func Info(v ...interface{}) {
	StdGLog.Info(v...)
}

func InfoNoCon(v ...interface{}) {
	StdGLog.InfoNoCon(v...)
}

func Warnf(format string, v ...interface{}) {
	StdGLog.Warnf(format, v...)
}
func WarnfNoCon(format string, v ...interface{}) {
	StdGLog.WarnfNoCon(format, v...)
}

func Warn(v ...interface{}) {
	StdGLog.Warn(v...)
}
func WarnNoCon(v ...interface{}) {
	StdGLog.WarnNoCon(v...)
}

func Errorf(format string, v ...interface{}) {
	StdGLog.Errorf(format, v...)
}
func ErrorfNoCon(format string, v ...interface{}) {
	StdGLog.ErrorfNoCon(format, v...)
}
func Error(v ...interface{}) {
	StdGLog.Error(v...)
}
func ErrorNoCon(v ...interface{}) {
	StdGLog.ErrorNoCon(v...)
}
func Fatalf(format string, v ...interface{}) {
	StdGLog.Fatalf(format, v...)
}

func Fatal(v ...interface{}) {
	StdGLog.Fatal(v...)
}

func Panicf(format string, v ...interface{}) {
	StdGLog.Panicf(format, v...)
}

func Panic(v ...interface{}) {
	StdGLog.Panic(v...)
}

func Stack(v ...interface{}) {
	StdGLog.Stack(v...)
}

func init() {
	// (因为StdGLog对象 对所有输出方法做了一层包裹，所以在打印调用函数的时候，比正常的logger对象多一层调用
	// 一般的gLogger对象 calldDepth=2, StdGLog的calldDepth=3)
	StdGLog.calldDepth = 3
}

func GetDefaultLogPath() string {
	dir, f := GetLogPath("./logs", "normal.log")
	return filepath.Join(dir, f)
}

func GetLogPath(logDir, logFile string) (string, string) {
	ip := getHostIp()
	//mid := time.Now().Format("2006010215")
	mid := time.Now().Format("20060102")
	hour := time.Now().Hour()
	mid += "/"
	mid += fmt.Sprintf("%02d", hour)
	mid += "/"
	mid += ip
	logdir := logDir + "/" + mid + "/normal_logs"
	logfile := logFile
	if isProductionEnv() {
		logdir = logDir
		name, extension := getFileNameAndFileExtension(logfile)
		if name != "" && extension != "" {
			logfile = fmt.Sprintf("%s_%s%s", name, ip, extension)
		}
	}
	return logdir, logfile
}

// InitLog logFileSize 单个日志最大容量 单位：字节
func InitLog(logDir, logFile string, logFileSize int64, logSaveDays int) {
	//ip := getHostIp()
	////mid := time.Now().Format("2006010215")
	//mid := time.Now().Format("20060102")
	//hour := time.Now().Hour()
	//mid += "/"
	//mid += fmt.Sprintf("%02d", hour)
	//mid += "/"
	//mid += ip
	//logdir := logDir + "/" + mid + "/normal_logs"
	//logfile := logFile
	//if isProductionEnv() {
	//	logdir = logDir
	//	name, extension := getFileNameAndFileExtension(logfile)
	//	if name != "" && extension != "" {
	//		logfile = fmt.Sprintf("%s_%s%s", name, ip, extension)
	//	}
	//}
	//时间显示到微秒级
	AddFlag(BitMicroSeconds)
	SetMaxSize(logFileSize)
	SetMaxAge(logSaveDays)
	//SetLogFile(logdir, logfile)
	SetLogFile(GetLogPath(logDir, logFile))
	SetCons(true)
}

func InitDefault() {
	InitLog("./logs", "normal.log", 1048576, 30)
}

func Init() {
	AddFlag(BitMicroSeconds)
	SetMaxSize(1048576)
	SetMaxAge(30)
	//SetLogFile(logdir, logfile)
	SetLogFile("./logs", "app.log")
	SetCons(true)
}

func isProductionEnv() bool {
	envType := os.Getenv("ENV_TYPE")
	//fmt.Println("【ENV_TYPE】环境变量", envType)
	if strings.EqualFold(strings.ToLower(envType), strings.ToLower("itest")) {
		//fmt.Println("测试环境")
		return false
	} else {
		//fmt.Println("默认生产环境")
	}
	return true
}

func getHostIp() string {
	addrList, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("get current host ip err: ", err)
		return ""
	}
	//var ips []net.IP
	for _, address := range addrList {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.IsPrivate() {
			if ipNet.IP.To4() != nil {
				//ip = ipNet.IP.String()
				//break
				ip := ipNet.IP.To4()
				//fmt.Println(ip[0])
				switch ip[0] {
				case 10:
					return ipNet.IP.String()
				case 192:
					return ipNet.IP.String()
				}
			}
		}
	}
	//fmt.Println(ips)
	return ""
}

func getFileNameAndFileExtension(filePath string) (string, string) {
	// 使用 filepath 包提供的函数获取文件名
	fileName := filepath.Base(filePath)
	// 使用 strings 包提供的函数获取文件名和扩展名
	fileNameWithoutExtension := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	fileExtension := filepath.Ext(fileName)
	// 打印文件名和扩展名
	//Println("文件名:", fileNameWithoutExtension)
	//Println("扩展名:", fileExtension)
	return fileNameWithoutExtension, fileExtension
}
