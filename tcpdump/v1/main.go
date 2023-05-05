package main

import (
	"bufio"
	"fmt"
	"github.com/fragmentization/mahonia"
	"os"
	"regexp"
)

func main() {
	//readfile()
	// 打开tcpdump生成的文件
	file, err := os.Open("C:\\Users\\ihse1\\Downloads\\1234.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 使用正则表达式匹配HTTP请求和响应
	//re := regexp.MustCompile("(?s)^(GET|POST|HEAD|PUT|DELETE|OPTIONS|TRACE|CONNECT).+?\r\n\r\n|^(HTTP/1\\..+?\\r\\n\\r\\n.*$)")
	re := regexp.MustCompile("(?s)^(GET|POST|HEAD|PUT|DELETE|OPTIONS|TRACE|CONNECT).+?\r\n\r\n|^(HTTP/1\\..+?\\r\\n\\r\\n.*$)")

	// 逐行读取文件内容
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		// 使用正则表达式提取HTTP请求和响应
		match := re.FindStringSubmatch(line)
		if len(match) > 0 {
			fmt.Println(match[0])
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func readfile() string {
	f, err := os.Open("C:\\Users\\ihse1\\Downloads\\1234.csv")
	if err != nil {
		return err.Error()
	}
	defer f.Close()
	buf := make([]byte, 1024)
	//文件ex7.txt的编码是gb18030
	decoder := mahonia.NewDecoder("gb18030")
	if decoder == nil {
		return "编码不存在!"
	}
	var str string = ""
	for {
		n, _ := f.Read(buf)
		if 0 == n {
			break
		}
		//解码为UTF-8
		str += decoder.ConvertString(string(buf[:n]))
	}
	return str
}
