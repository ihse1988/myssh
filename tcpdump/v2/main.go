package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("C:\\Users\\ihse1\\Downloads\\output.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	//
	//scanner := bufio.NewScanner(transform.NewReader(file, simplifiedchinese.GBK.NewDecoder()))
	//scanner.Split(bufio.ScanLines)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 处理每一行数据
		//fmt.Println(line)
		split := strings.Split(line, "\t")
		for index, value := range split {
			// 循环处理每个分隔后的值
			fmt.Println(index, value)
		}
	}
}
