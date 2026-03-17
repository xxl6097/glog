package z

import "fmt"

func Debugf(format string, v ...interface{}) {
	L().Debug(fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	L().Debug(fmt.Sprintln(v...))
}

func Infof(format string, v ...interface{}) {
	L().Info(fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	L().Info(fmt.Sprintln(v...))
}

func Warnf(format string, v ...interface{}) {
	L().Warn(fmt.Sprintf(format, v...))
}

func Warn(v ...interface{}) {
	L().Warn(fmt.Sprintln(v...))
}

func Errorf(format string, v ...interface{}) {
	L().Error(fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	L().Error(fmt.Sprintln(v...))
}

func Fatalf(format string, v ...interface{}) {
	L().Fatal(fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	L().Fatal(fmt.Sprintln(v...))
}
