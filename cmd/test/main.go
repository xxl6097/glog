package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

func main() {
	path := "./data.log"
	file, e1 := safeOpenFile(path)
	if e1 != nil {
		log.Fatal("文件打开失败:", e1)
	}
	defer file.Close() // 确保资源释放

	time.Sleep(10 * time.Second)
	// 检查文件路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("文件已被删除")
		file, e1 = safeOpenFile(path)
	}
	content := "新日志内容\n"
	if e := writeContent(file, content); e != nil {
		// 处理写入失败（如重建文件句柄）
		if errors.Is(e, syscall.EBADF) { // 文件被外部删除
			file.Close()
			newFile, _ := safeOpenFile(path)
			*file = *newFile
			_ = writeContent(file, content) // 重试写入
		} else {
			log.Fatal("写入不可恢复错误:", e)
		}
	}
}

func safeOpenFile(path string) (*os.File, error) {
	// 原子操作：尝试创建或追加文件（避免竞态）
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		if os.IsNotExist(err) { // 文件不存在（但已通过O_CREATE自动创建）
			return nil, fmt.Errorf("目录不存在或权限不足")
		}
		return nil, err
	}
	return file, nil
}

func writeContent(file *os.File, content string) error {
	writer := bufio.NewWriter(file)

	// 写入内容并检测字节数及错误
	bytesWritten, err := writer.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入失败: %w", err)
	}

	// 必须刷新缓冲区！否则数据可能未落盘
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("刷新缓冲区失败: %w", err)
	}

	fmt.Printf("成功写入 %d 字节\n", bytesWritten)
	return nil
}
