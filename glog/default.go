package glog

import (
	"context"
	"fmt"
)

var instance ILogger = new(GlogDefault)

type GlogDefault struct{}

func (log *GlogDefault) InfoF(format string, v ...interface{}) {
	StdGLog.Infof(format, v...)
}

func (log *GlogDefault) ErrorF(format string, v ...interface{}) {
	StdGLog.Errorf(format, v...)
}

func (log *GlogDefault) DebugF(format string, v ...interface{}) {
	StdGLog.Debugf(format, v...)
}

func (log *GlogDefault) InfoFX(ctx context.Context, format string, v ...interface{}) {
	fmt.Println(ctx)
	StdGLog.Infof(format, v...)
}

func (log *GlogDefault) ErrorFX(ctx context.Context, format string, v ...interface{}) {
	fmt.Println(ctx)
	StdGLog.Errorf(format, v...)
}

func (log *GlogDefault) DebugFX(ctx context.Context, format string, v ...interface{}) {
	fmt.Println(ctx)
	StdGLog.Debugf(format, v...)
}

func SetLogger(newlog ILogger) {
	instance = newlog
}

func Ins() ILogger {
	return instance
}
