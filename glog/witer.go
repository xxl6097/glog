package glog

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var colors = []string{
	Green,
	Blue,
	Magenta,
	Red,
	BlueBold,
	Yellow,
}

const (
	sizeMiB    = 1024 * 1024
	defMaxAge  = 31
	defMaxSize = 64 //MiB
)

var _ io.WriteCloser = (*Writer)(nil)

type Writer struct {
	maxAge       int       // 最大保留天数
	daemonSecond int       // 日志多久一次写文件，单位秒
	maxSize      int64     // 单个日志最大容量 默认 64MB
	size         int64     // 累计大小
	fpath        string    // 文件目录 完整路径 fpath=fdir+fname+fsuffix
	fdir         string    //
	fname        string    // 文件名
	fsuffix      string    // 文件后缀名 默认 .log
	zipsuffix    string    // 文件后缀名 默认 .log
	created      time.Time // 文件创建日期
	creates      []byte    // 文件创建日期
	cons         bool      // 标准输出  默认 false
	nocolor      bool      // 颜色输出
	noHeader     bool      // 日志头
	file         *os.File
	bw           *bufio.Writer
	mu           sync.Mutex
	out          io.Writer //日志输出的文件描述符
}

func New(out io.Writer, path string) *Writer {
	w := &Writer{
		fpath:        path, //dir1/dir2/app.log
		mu:           sync.Mutex{},
		out:          out,
		cons:         true,
		daemonSecond: 5,
	}
	w.fdir = filepath.Dir(w.fpath)                                  //dir1/dir2
	w.fsuffix = filepath.Ext(w.fpath)                               //.log
	w.fname = strings.TrimSuffix(filepath.Base(w.fpath), w.fsuffix) //app
	if w.fsuffix == "" {
		w.fsuffix = ".log"
	}
	if w.zipsuffix == "" {
		w.zipsuffix = ".zip"
	}
	w.maxSize = sizeMiB * defMaxSize
	w.maxAge = defMaxAge
	err := os.MkdirAll(filepath.Dir(w.fpath), 0755)
	if err != nil {
		fmt.Println(err)
	}
	go w.daemon()
	return w
}
func (w *Writer) daemon() {
	for range time.NewTicker(time.Second * time.Duration(w.daemonSecond)).C {
		err := w.flush()
		if err != nil {
		}
		_, e := os.Stat(w.fpath)
		if e != nil && os.IsNotExist(e) {
			_ = os.MkdirAll(w.fdir, 0755)
			fout, e2 := os.OpenFile(w.fpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
			if e2 == nil {
				w.file = fout
				w.bw = bufio.NewWriter(w.file)
			}

		}
	}
}

// SetMaxAge 最大保留天数
func (w *Writer) SetMaxAge(ma int) {
	w.mu.Lock()
	w.maxAge = ma
	w.mu.Unlock()
}

// SetMaxSize 单个日志最大容量
func (w *Writer) SetMaxSize(ms int64) {
	if ms < 1 {
		return
	}
	w.mu.Lock()
	w.maxSize = ms
	w.mu.Unlock()
}

// SetCons 同时输出控制台
func (w *Writer) SetCons(b bool) {
	w.mu.Lock()
	w.cons = b
	w.mu.Unlock()
}

// SetDaemonSecond 日志写文件周期时钟
func (w *Writer) SetDaemonSecond(second int) {
	w.mu.Lock()
	w.daemonSecond = second
	w.mu.Unlock()
}

func (w *Writer) SetNoColor(b bool) {
	w.mu.Lock()
	w.nocolor = b
	w.mu.Unlock()
}

// SetNoHeader 头指的时间，行号等信息
func (w *Writer) SetNoHeader(b bool) {
	w.mu.Lock()
	w.noHeader = b
	w.mu.Unlock()
}

//func (w *Writer) WriteInConsole(level int, p []byte) {
//	w.mu.Lock()
//	defer w.mu.Unlock()
//	if w.cons {
//		var buf bytes.Buffer
//		if !w.nocolor {
//			buf.WriteString(colors[level])
//		}
//		buf.Write(p)
//		if !w.nocolor {
//			buf.WriteString(Reset)
//		}
//		w.out.Write(buf.Bytes())
//	}
//}

func (w *Writer) Write(buffer []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := buffer[2:] //第一字节是level，第二字节是header长度
	index := buffer[0]
	isCons := index >> 3
	header_size := buffer[1]
	if w.cons {
		var buf bytes.Buffer
		if !w.nocolor {
			index &= 0x07
			buf.WriteString(colors[index])
		}
		if w.noHeader {
			//header_size := int((uint(buffer[1]) << 24) |
			//	(uint(buffer[2]) << 16) |
			//	(uint(buffer[3]) << 8) |
			//	uint(buffer[4]))
			buf.Write(p[header_size:])
		} else {
			buf.Write(p)
		}
		if !w.nocolor {
			buf.WriteString(Reset)
		}
		if isCons == 0 {
			w.out.Write(buf.Bytes())
		} else {
			w.out.Write(buf.Bytes()[:header_size+(50)])
			w.out.Write([]byte{10, 13}) //回车换行
		}
	}
	if w.file == nil {
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
	if w.size+int64(len(p)) >= w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	// n, err = w.file.Write(p)
	n, err = w.bw.Write(p)
	w.size += int64(n)
	if err != nil {
		return n, err
	}
	return
}

// rotate 切割文件
func (w *Writer) rotate() error {
	now := time.Now()
	if w.file != nil {
		w.bw.Flush()
		w.file.Sync()
		w.file.Close()
		// 保存
		fbak := w.fname + w.time2name(w.created)
		fbakname := fbak + w.fsuffix
		err := os.Rename(w.fpath, filepath.Join(w.fdir, fbakname))
		if err == nil {
			err1 := ZipToFile(filepath.Join(w.fdir, fbak+".zip"), filepath.Join(w.fdir, fbakname))
			if err1 == nil {
				os.Remove(filepath.Join(w.fdir, fbakname))
			} else {
				fmt.Println(err1)
			}
		}

		w.size = 0
	}
	finfo, err := os.Stat(w.fpath)
	w.created = now
	if err == nil {
		w.size = finfo.Size()
		w.created = finfo.ModTime()
	}
	w.creates = w.created.AppendFormat(nil, time.RFC3339)
	fout, err := os.OpenFile(w.fpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		//fmt.Println("witer.go rotate", w.fpath, err)
		return err
	}
	w.file = fout
	w.bw = bufio.NewWriter(w.file)
	return nil
}

// 删除旧日志
func (w *Writer) delete() {
	if w.maxAge <= 0 {
		return
	}
	dir := filepath.Dir(w.fpath)
	fakeNow := time.Now().AddDate(0, 0, -w.maxAge)
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
func (w *Writer) name2time(name string) (time.Time, error) {
	name = strings.TrimPrefix(name, filepath.Base(w.fname))
	name = strings.TrimSuffix(name, w.zipsuffix)
	//return time.Parse(".2006-01-02-150405", name)
	return time.Parse(".2006-01-02", name)
}

func (w *Writer) time2name(t time.Time) string {
	//return t.Format(".2006-01-02-150405")
	return t.Format(".2006-01-02")
}

func (w *Writer) Close() error {
	err := w.flush()
	err = w.close()
	if err != nil {
		//fmt.Println("Writer Close", err)
	}
	return err
}

// close closes the file if it is open.
func (w *Writer) close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return errors.New("file is nil")
	}
	err := w.file.Sync()
	err = w.file.Close()
	w.file = nil
	return err
}

func (w *Writer) flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.bw == nil {
		//fmt.Println("Writer flush w.bw is nil")
		return nil
	}
	return w.bw.Flush()
}

func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.bw == nil {
		return nil
	}
	return w.bw.Flush()
}

// appendInt appends the decimal form of x to b and returns the result.
// If the decimal form (excluding sign) is shorter than width, the result is padded with leading 0's.
// Duplicates functionality in strconv, but avoids dependency.
func appendInt(b []byte, x int, width int) []byte {
	u := uint(x)
	if x < 0 {
		b = append(b, '-')
		u = uint(-x)
	}

	// 2-digit and 4-digit fields are the most common in time formats.
	utod := func(u uint) byte { return '0' + byte(u) }
	switch {
	case width == 2 && u < 1e2:
		return append(b, utod(u/1e1), utod(u%1e1))
	case width == 4 && u < 1e4:
		return append(b, utod(u/1e3), utod(u/1e2%1e1), utod(u/1e1%1e1), utod(u%1e1))
	}

	// Compute the number of decimal digits.
	var n int
	if u == 0 {
		n = 1
	}
	for u2 := u; u2 > 0; u2 /= 10 {
		n++
	}

	// Add 0-padding.
	for pad := width - n; pad > 0; pad-- {
		b = append(b, '0')
	}

	// Ensure capacity.
	if len(b)+n <= cap(b) {
		b = b[:len(b)+n]
	} else {
		b = append(b, make([]byte, n)...)
	}

	// Assemble decimal in reverse order.
	i := len(b) - 1
	for u >= 10 && i > 0 {
		q := u / 10
		b[i] = utod(u - q*10)
		u = q
		i--
	}
	b[i] = utod(u)
	return b
}

// ZipToFile 压缩至文件
// @params dst string 压缩文件目标路径
// @params src string 待压缩源文件/目录路径
// @return     error  错误信息
func ZipToFile(dst, src string) error {
	// 创建一个ZIP文件
	fw, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return err
	}
	defer fw.Close()

	// 执行压缩
	return Zip(fw, src)
}

// Zip 压缩文件或目录
// @params dst io.Writer 压缩文件可写流
// @params src string    待压缩源文件/目录路径
func Zip(dst io.Writer, src string) error {
	// 强转一下路径
	src = filepath.Clean(src)
	// 提取最后一个文件或目录的名称
	baseFile := filepath.Base(src)
	// 判断src是否存在
	_, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 通文件流句柄创建一个ZIP压缩包
	zw := zip.NewWriter(dst)
	// 延迟关闭这个压缩包
	defer zw.Close()

	// 通过filepath封装的Walk来递归处理源路径到压缩文件中
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		// 是否存在异常
		if err != nil {
			return err
		}

		// 通过原始文件头信息，创建zip文件头信息
		zfh, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 赋值默认的压缩方法，否则不压缩
		zfh.Method = zip.Deflate

		// 移除绝对路径
		tmpPath := path
		index := strings.Index(tmpPath, baseFile)
		if index > -1 {
			tmpPath = tmpPath[index:]
		}
		// 替换文件名，并且去除前后 "\" 或 "/"
		tmpPath = strings.Trim(tmpPath, string(filepath.Separator))
		// 替换一下分隔符，zip不支持 "\\"
		zfh.Name = strings.ReplaceAll(tmpPath, "\\", "/")
		// 目录需要拼上一个 "/" ，否则会出现一个和目录一样的文件在压缩包中
		if info.IsDir() {
			zfh.Name += "/"
		}

		// 写入文件头信息，并返回一个ZIP文件写入句柄
		zfw, err := zw.CreateHeader(zfh)
		if err != nil {
			return err
		}

		// 仅在他是标准文件时进行文件内容写入
		if zfh.Mode().IsRegular() {
			// 打开要压缩的文件
			sfr, err := os.Open(path)
			if err != nil {
				return err
			}
			defer sfr.Close()

			// 将srcFileReader拷贝到zipFilWrite中
			_, err = io.Copy(zfw, sfr)
			if err != nil {
				return err
			}
		}

		// 搞定
		return nil
	})
}
