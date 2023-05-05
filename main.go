package main

import (
	"bufio"
	"fmt"
	myssh "gitee.com/ihse198888/mymodules/shell/V3"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func main() {
	// 打开原始文件
	file, err := os.Open("C:\\Users\\ihse1\\Desktop\\tmp\\mysql-slow.tag\\mysql-slow.log.17")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 创建输出目录
	err = os.MkdirAll("C:\\Users\\ihse1\\Desktop\\tmp\\mysql-slow.tag\\output", 0777)
	if err != nil {
		log.Fatal(err)
	}

	// 定义文件名格式
	outputFileFormat := "C:\\Users\\ihse1\\Desktop\\tmp\\mysql-slow.tag\\output/mysql-slow-%s.log"

	// 初始化变量
	var outputFile *os.File
	var lastTime time.Time

	// 逐行读取原始文件
	reader := bufio.NewReader(file)
	for {
		// 读取一行
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		// 检查是否需要打开新文件
		if isTimeLine(line) {
			t, err := parseTimestamp(line)
			if err != nil {
				log.Fatal(err)
			}
			if !t.Equal(lastTime) {
				// 关闭旧文件
				if outputFile != nil {
					outputFile.Close()
				}

				// 打开新文件
				lastTime = t
				outputFileName := fmt.Sprintf(outputFileFormat, lastTime.Format("2006-01-02T15_04_05"))
				outputFile, err = os.OpenFile(outputFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
				if err != nil {
					log.Fatal(err)
				}
				defer outputFile.Close()
			}
		}

		// 写入当前行
		_, err = outputFile.WriteString(line)
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

func isTimeLine(line string) bool {
	return len(line) > 7 && line[:7] == "# Time:"
}

// 解析日志行中的时间信息
func parseTimestamp(line string) (time.Time, error) {
	const timeFormat = "2006-01-02T15:04:05.999999Z"
	if len(line) < len("# Time: "+timeFormat) {
		return time.Time{}, fmt.Errorf("invalid line: %q", line)
	}
	timestampStr := line[len("# Time: ") : len("# Time: ")+len(timeFormat)]
	timestamp, err := time.Parse(timeFormat, timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %q", timestampStr)
	}
	return timestamp, nil
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

// 使用buff的方式分割
func buffRead() {
	const chunkSize = 1 * 1024 * 1024 * 1024 // 1GB

	filePath := "C:\\Users\\ihse1\\Desktop\\tmp\\mysql-slow.tag\\mysql-slow.log"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var fileIndex int
	for {
		buffer := make([]byte, chunkSize)
		bytesRead, err := file.Read(buffer)

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		if bytesRead == 0 {
			break
		}

		fileIndex++

		newFilePath := fmt.Sprintf("%s.%d", filePath, fileIndex)
		newFile, err := os.Create(newFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer newFile.Close()

		writer := bufio.NewWriter(newFile)
		if _, err := writer.Write(buffer[:bytesRead]); err != nil {
			log.Fatal(err)
		}

		writer.Flush()
	}

	// Rename original file
	newFilePath := fmt.Sprintf("%s.%d", filePath, fileIndex+1)
	if err := os.Rename(filePath, newFilePath); err != nil {
		log.Fatal(err)
	}

	// Create a new empty file
	if _, err := os.Create(filePath); err != nil {
		log.Fatal(err)
	}
}

func lineRead() {
	// 打开原始文件
	file, err := os.Open("C:\\Users\\ihse1\\Desktop\\tmp\\mysql-slow.tag\\mysql-slow.tag")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 创建输出目录
	err = os.MkdirAll("C:\\Users\\ihse1\\Desktop\\tmp\\output", 0777)
	if err != nil {
		log.Fatal(err)
	}

	// 定义输出文件名格式
	outputFileFormat := "C:\\Users\\ihse1\\Desktop\\tmp\\output\\mysql-slow-%d.log"

	// 定义文件大小（字节）
	maxFileSize := int64(2 * 1024 * 1024 * 1024)

	// 初始化变量
	fileCount := 1
	byteCount := int64(0)

	// 逐行读取原始文件
	reader := bufio.NewReader(file)
	for {
		// 读取一行
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		// 如果当前文件大小超过了最大值，关闭当前文件，打开一个新文件
		if byteCount+int64(len(line)) > maxFileSize {
			byteCount = 0
			fileCount++
		}

		// 打开输出文件
		outputFileName := fmt.Sprintf(outputFileFormat, fileCount)
		outputFile, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFile.Close()

		// 写入当前行
		_, err = outputFile.WriteString(line)
		if err != nil {
			log.Fatal(err)
		}

		// 更新变量
		byteCount += int64(len(line))
	}

	// 输出文件数量
	fmt.Printf("Split into %d files\n", fileCount)
}

func test() {
	//测试golang读写
	fileName := "C:\\Users\\ihse1\\Desktop\\tmp\\外网NG日志\\443.access.log_2023033023\\443.access.log_2023033023"
	file, err := os.OpenFile(fileName, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Open file error!", err)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}
	bT := time.Now()
	var size = stat.Size()
	fmt.Println("file size=", size)

	buf := bufio.NewReader(file)
	pattern := `"uri":"(.*?)",.*?upstream_response_time":(.*?),`
	//re := regexp.MustCompile(pattern)
	compile, err := regexp.Compile(pattern)
	for {
		line, err := buf.ReadString('\n')
		//line = strings.TrimSpace(line)
		//match := re.FindStringSubmatch(line)
		//if len(match) == 3 {
		//	//fmt.Printf(match[1])
		//}
		matchString := compile.FindStringSubmatch(line)
		if len(matchString) > 0 {

		}
		//fmt.Println(line)
		if err != nil {
			if err == io.EOF {
				fmt.Println("File read ok!")
				break
			} else {
				fmt.Println("Read file error!", err)
				return
			}
		}
	}
	eT := time.Since(bT) // 从开始到当前所消耗的时间

	fmt.Println("Run time: ", eT)
}

func temp() {
	// 目标主机的地址和端口
	host := "192.168.100.85"
	port := "60022"

	// SSH登陆的用户名和密码
	userName := "xdw"
	password := "YnDQ@189"

	my := myssh.NewMySsh(host, port, userName, password, password)
	my.InitCommonTerminal()
	output, err := my.SendCmdWithOut("ls")
	fmt.Println(output)
	if err != nil {
		panic(err)
	}
	output, err = my.SendCmdWithOut("find / -name temp ")
	fmt.Println(output)
	if err != nil {
		panic(err)
	}
}

// 提示用户输入密码，并返回输入的密码
func promptForPassword() string {
	//fmt.Print("Enter root password: ")
	//bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	//if err != nil {
	//	fmt.Println("Failed to read password: ", err)
	//	os.Exit(1)
	//}
	//password := string(bytePassword)
	//return password
	return "YnDQ@189"
}
