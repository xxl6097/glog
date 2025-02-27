 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build *.go
 bash <(curl -s -S -L http://uuxia.cn:8086/up) ./main.exe