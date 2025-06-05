package glog

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	sizeMiB    = 1024 * 1024
	defMaxAge  = 31
	defMaxSize = 64 //MiB
)

type LogWriter struct {
	logPath       string
	logFile       *os.File
	logFileName   string
	logZipsuffix  string
	logFileBuffer *bufio.Writer
	created       time.Time // 文件创建日期
	MaxAge        int
	size          int64  // 累计大小
	MaxSize       int64  // 单个日志最大容量 默认 64MB
	creates       []byte // 文件创建日期
	DaemonSecond  int    // 日志多久一次写文件，单位秒
}

func NewFileWriter(logPath string) *LogWriter {
	fsuffix := filepath.Ext(logPath)                             //.log
	fname := strings.TrimSuffix(filepath.Base(logPath), fsuffix) //app
	if fsuffix == "" {
		fsuffix = ".log"
	}
	err := os.MkdirAll(filepath.Dir(logPath), 0755)
	if err != nil {
		fmt.Println(err)
	}

	logWriter := &LogWriter{
		logPath:       logPath,
		logFileName:   fname,
		MaxSize:       sizeMiB * defMaxSize,
		MaxAge:        defMaxAge,
		logZipsuffix:  ".zip",
		DaemonSecond:  5,
		logFileBuffer: bufio.NewWriter(os.Stdout),
	}
	go logWriter.daemon()
	return logWriter
}

func (w *LogWriter) daemon() {
	for range time.NewTicker(time.Second * time.Duration(w.DaemonSecond)).C {
		err := w.Flush()
		if err != nil {
			Error("daemon err", err)
			//TextToFile("./tmp/glog.log", err.Error())
		}
	}
}

func (w *LogWriter) Write(p []byte) (int, error) {
	// 检查文件路径是否存在
	if _, err := os.Stat(w.logPath); os.IsNotExist(err) {
		fmt.Println("文件已被删除")
		if err1 := w.rotate(); err1 != nil {
			//w.out.Write(p)
			return 0, err1
		}
	}
	if w.logFile == nil {
		if err1 := w.rotate(); err1 != nil {
			//w.out.Write(p)
			return 0, err1
		}
	}

	t := time.Now()
	var b []byte
	year, month, day := t.Date()
	b = appendInt(b, year, 4)
	b = append(b, '-')
	b = appendInt(b, int(month), 2)
	b = append(b, '-')
	b = appendInt(b, day, 2)

	// 按天切割
	if !bytes.Equal(w.creates[:10], b) { //2023-04-05
		go w.delete() // 每天检测一次旧文件
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	// 按大小切割
	if w.size+int64(len(p)) >= w.MaxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.logFileBuffer.Write(p)
	w.size += int64(n)
	return n, err
}

// rotate 切割文件
func (w *LogWriter) rotate() error {
	now := time.Now()
	fsuffix := filepath.Ext(w.logPath)
	fdir := filepath.Dir(w.logPath)
	fname := strings.TrimSuffix(filepath.Base(w.logPath), fsuffix) //app
	if w.logFile != nil {
		w.logFileBuffer.Flush()
		w.logFile.Sync()
		w.logFile.Close()
		// 保存
		fbak := fname + w.time2name(w.created)
		fbakname := fbak + fsuffix
		err := os.Rename(w.logPath, filepath.Join(fdir, fbakname))
		if err == nil {
			err1 := ZipToFile(filepath.Join(fdir, fbak+".zip"), filepath.Join(fdir, fbakname))
			if err1 == nil {
				_ = os.Remove(filepath.Join(fdir, fbakname))
			} else {
				fmt.Println(err1)
			}
		}

		w.size = 0
	}

	if _, err := os.Stat(fdir); os.IsNotExist(err) {
		// 目录不存在，创建它
		e := os.MkdirAll(filepath.Clean(fdir), 0755)
		if e != nil {
			fmt.Printf("创建目录失败: %w\n", e)
		}
		fmt.Printf("目录 %s 创建成功\n", fdir)
	}

	fInfo, err := os.Stat(w.logPath)
	w.created = now
	if err == nil {
		w.size = fInfo.Size()
		w.created = fInfo.ModTime()
	}
	w.creates = w.created.AppendFormat(nil, time.RFC3339)
	fout, err := os.OpenFile(w.logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		//fmt.Println("witer.go rotate", w.fpath, err)
		return err
	}
	w.logFile = fout
	w.logFileBuffer = bufio.NewWriter(w.logFile)
	return nil
}

// 删除旧日志
func (w *LogWriter) delete() {
	if w.MaxAge <= 0 {
		return
	}
	dir := filepath.Dir(w.logPath)
	fakeNow := time.Now().AddDate(0, 0, -w.MaxAge)
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, path := range dirs {
		name := path.Name()
		if path.IsDir() {
			continue
		}
		t, err := w.name2time(name)
		// 只删除满足格式的文件
		if err == nil && t.Before(fakeNow) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

func (w *LogWriter) Close() error {
	if w.logFile == nil {
		return errors.New("file is nil")
	}
	err := w.logFile.Sync()
	err = w.logFile.Close()
	w.logFile = nil
	return err
}

func (w *LogWriter) Flush() error {
	if w.logFileBuffer == nil {
		//fmt.Println("Writer flush w.bw is nil")
		return fmt.Errorf("writer flush w.bw is nil")
	}
	// 必须刷新缓冲区！否则数据可能未落盘
	if err := w.logFileBuffer.Flush(); err != nil {
		return fmt.Errorf("刷新缓冲区失败: %w", err)
	}
	return nil
}

func (w *LogWriter) name2time(name string) (time.Time, error) {
	name = strings.TrimPrefix(name, filepath.Base(w.logFileName))
	name = strings.TrimSuffix(name, w.logZipsuffix)
	return time.Parse(".2006-01-02", name)
}
func (w *LogWriter) time2name(t time.Time) string {
	return t.Format(".2006-01-02")
}
