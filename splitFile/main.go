package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// 是否追加写入文件
const flag int = os.O_RDWR | os.O_CREATE //| os.O_APPEND

// mysql-slow大文件按天拆分
// 使用readline的方式读取，日期不同记录offset和时间
// 使用buffer的方式拷贝分段数据
func main() {

	srcFile := "C:\\Users\\ihse1\\Desktop\\tmp\\mysql-slow.tag\\mysql-slow.log.18"
	// 打开原始文件
	file, err := os.Open(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 创建输出目录
	err = os.MkdirAll(filepath.Dir(srcFile)+"/output", 0777)
	if err != nil {
		log.Fatal(err)
	}

	// 定义文件名格式
	outputFileFormat := filepath.Dir(srcFile) + "/output" + "/mysql-slow-%s.log"

	// 读取文件并记录分割位置
	var splitPositions []int64
	var splitTimes []string
	var prevTime string
	var offset int64 //增加buffer后的偏移量
	scanner := bufio.NewScanner(file)
	const maxScanTokenSize = 30 * 1024 * 1024 //30M
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)
	//自定义分割方法，虽然默认的分割是\n，但是\r\n结尾的解析后读取的line会缺少\r导致offset无法对齐
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, []byte("\n")); i >= 0 {
			// 找到了\n
			return i + 1, data[0:i], nil
		}
		if atEOF {
			// 已经到达文件末尾，但是没有找到\r\n
			return len(data), data, nil
		}
		// 缓冲区中还没有到达文件末尾，等待下一次读取
		return 0, nil, nil
	})
	for scanner.Scan() {
		//bytes := scanner.Bytes()
		//line := string(bytes)
		line := scanner.Text()
		//n := int64(utf8.RuneCountInString(line) + 1) // 一个中文长度为1，明显也是不对的
		n := int64(len(line) + 1) //第一次少了628   143333881
		if isTimeLine(line) {
			if prevTime == "" {
				splitPositions = append(splitPositions, 0)
				splitTimes = append(splitTimes, "")
			}
			if prevTime != "" && prevTime != getTime(line) {
				// 找到下一个时间并且与前一个时间不相等，记录位置
				splitPositions = append(splitPositions, offset) //
				splitTimes = append(splitTimes, prevTime)
				fmt.Println("时间", prevTime, "pos", offset)
			}
			prevTime = getTime(line)

		}
		// 更新文件偏移量
		offset += n
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	//最后一个文件
	splitPositions = append(splitPositions, offset) //
	splitTimes = append(splitTimes, prevTime)
	fmt.Println("时间", prevTime, "pos", offset)
	// 分割文件
	for i, pos := range splitPositions {
		if i == 0 {
			// 第一个文件不需要截取头部
			continue
		}
		prevPos := splitPositions[i-1]

		fileName := fmt.Sprintf(outputFileFormat, splitTimes[i])
		err = splitFile(srcFile, fileName, prevPos, pos)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 输出文件数量
	fileCount, err := countOutputFiles()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Split into %d files\n", fileCount)
}

// 判断是否为时间行
func isTimeLine(line string) bool {
	return strings.HasPrefix(line, "# Time:")
}

// 获取时间字符串，，修改这里即可达到修改分割粒度的问题
func getTime(line string) string {
	return strings.TrimSpace(line[7:18])
}

// 分割文件，并写入
func splitFile(srcFile, dstFile string, start, end int64) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = src.Seek(start, os.SEEK_SET)
	if err != nil {
		return err
	}

	//if isAppend {
	dst, err := os.OpenFile(dstFile, flag, 0666)
	if err != nil {
		return err
	}
	//	defer dst.Close()
	//} else {
	//	dst, err = os.Create(dstFile)
	//	if err != nil {
	//		return err
	//	}
	//	defer dst.Close()
	//}

	_, err = io.CopyN(dst, src, end-start)
	if err != nil {
		return err
	}

	return nil
}

// 统计输出目录中的文件数量
func countOutputFiles() (int, error) {
	var count int
	err := filepath.Walk("output", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
