package main

import (
	"os"
)

func main() {
	file, err := os.OpenFile("filename.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 写入内容
	_, err = file.WriteString("Hello, world!\n")
	if err != nil {
		panic(err)
	}
}
