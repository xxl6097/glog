# go-glog的使用方式

github.com/xxl6097

```shell
git add .
git commit -m "release v0.0.6"
git tag -a v0.0.6 -m "release v0.0.6"
git push origin v0.0.6
```

## 编译

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o main.exe main.go
```
## 一、添加依赖

```go
go get -u github.com/xxl6097/glog
```

## 二、使用示例
```go

package main

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"time"
)

func hook(data []byte) {
	fmt.Println(string(data))
}

func init() {
	//开启日志保存文件
	glog.LogSaveFile()
	//拦截日志
	//glog.Hook(hook)
}

func main() {
	for {
		fmt.Println("aaa")
		glog.Info("只有使用这个log打印才能记录日志哦")
		time.Sleep(time.Second)
	}
}

```


