package glog

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var colors = []string{
	Green,
	Blue,
	Magenta,
	Red,
	BlueBold,
	Yellow,
}

var _ io.WriteCloser = (*Writer)(nil)

type Writer struct {
	cons      bool // 标准输出  默认 false
	nocolor   bool // 颜色输出
	noHeader  bool // 日志头
	mu        sync.Mutex
	out       io.Writer //日志输出的文件描述符
	logWriter *LogWriter
}

func New(out io.Writer) *Writer {
	w := &Writer{
		mu:   sync.Mutex{},
		out:  out,
		cons: true,
	}
	return w
}

func (w *Writer) SetLogFile(logFile string) {
	w.logWriter = NewFileWriter(logFile)
}

// SetMaxAge 最大保留天数
func (w *Writer) SetMaxAge(ma int) {
	if w.logWriter == nil {
		return
	}
	w.mu.Lock()
	w.logWriter.MaxAge = ma
	w.mu.Unlock()
}

// SetMaxSize 单个日志最大容量
func (w *Writer) SetMaxSize(ms int64) {
	if ms < 1 {
		return
	}
	if w.logWriter == nil {
		return
	}
	w.mu.Lock()
	w.logWriter.MaxSize = ms
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
	if w.logWriter == nil {
		return
	}
	w.mu.Lock()
	w.logWriter.DaemonSecond = second
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

func (w *Writer) Write(buffer []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := buffer[2:] //第一字节是level，第二字节是header长度
	index := buffer[0]
	isCons := index >> 3
	printToConsole := index >> 1
	header_size := buffer[1]
	if w.cons && printToConsole == 0 {
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
			n, err = buf.Write(p[header_size:])
		} else {
			n, err = buf.Write(p)
		}
		if !w.nocolor {
			n, err = buf.WriteString(Reset)
		}
		if isCons == 0 {
			n, err = w.out.Write(buf.Bytes())
		} else {
			n, err = w.out.Write(buf.Bytes()[:header_size+(50)])
			n, err = w.out.Write([]byte{10, 13}) //回车换行
		}
	}
	if w.logWriter != nil {
		n, err = w.logWriter.Write(p)
		if printToConsole == 1 {
			_ = w.logWriter.Flush()
		}
	}
	return
}

func (w *Writer) IsLogSave() bool {
	if w.logWriter != nil {
		return true
	}
	return false
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
	if w.logWriter == nil {
		return errors.New("logWriter is nil")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.logWriter.Close()
}

func (w *Writer) flush() error {
	if w.logWriter == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.logWriter.Flush()
}

func (w *Writer) Flush() error {
	return w.flush()
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
