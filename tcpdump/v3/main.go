package main

import (
	"encoding/csv"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"regexp"
)

func main() {
	// 读取csv文件
	file, err := os.Open("C:\\Users\\ihse1\\Desktop\\tmp\\tcpdump\\output.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.LazyQuotes = true    //单引号修改为true 表示不读取双引号
	reader.FieldsPerRecord = -1 // 允许变长字段

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// 将http.response_in不为空的行记录到map中
	responseMap := make(map[string][]string) //记录response
	for _, row := range records {
		//fmt.Println(row)
		if row[3] != "" {
			responseMap[row[0]] = row
		}
	}

	// 写入新的csv文件
	outFile, err := os.OpenFile("C:\\Users\\ihse1\\Desktop\\tmp\\tcpdump\\outputlast.csv", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	//outFile, err := os.Create("C:\\Users\\ihse1\\Desktop\\tmp\\tcpdump\\outputlast.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	writer.Comma = '\t'
	f := xlsx.NewFile()
	sheet, err := f.AddSheet("Sheet1")
	if err != nil {
		panic(err)
	}

	for _, row := range records {
		strings_ := responseMap[row[2]]
		if strings_ != nil {
			excelRow := sheet.AddRow()
			row = append(row, strings_[4])
			//row[7] = strings[6] //追加到最后一行
			for tempIndex, tempString := range row {
				//fmt.Println(tempString[1 : len(tempString)-1])  && strings.Contains(tempString, "\"")
				if len(tempString) > 1 {
					// 定义手机号码和身份证号码的正则表达式
					phoneRe, _ := regexp.Compile(`1[3-9]\d{9}`)
					idCardRe, _ := regexp.Compile(`\d{17}[\d|x]|\d{15}`)

					// 对手机号码进行脱敏
					maskedText := phoneRe.ReplaceAllStringFunc(tempString, func(phone string) string {
						return maskPhone(phone)
					})

					// 对身份证号码进行脱敏
					maskedText = idCardRe.ReplaceAllStringFunc(maskedText, func(idCard string) string {
						return maskIDCard(idCard)
					})

					cell := excelRow.AddCell()
					cell.Value = maskedText[1 : len(maskedText)-1]
					row[tempIndex] = maskedText[1 : len(maskedText)-1]
					//row[tempIndex] = strconv.Quote(tempString[1 : len(tempString)-1])
				}
			}
			writer.Write(row)
		}

	}
	err = f.Save("C:\\Users\\ihse1\\Desktop\\tmp\\tcpdump\\temp1.xlsx")
	if err != nil {
		panic(err)
	}
	writer.Flush()
}

// 将手机号码脱敏，将前3位和后4位替换为*号
func maskPhone(phone string) string {
	re, _ := regexp.Compile(`(\d{3})\d{4}(\d{4})`)
	return re.ReplaceAllString(phone, "$1****$2")
}

// 将身份证号码脱敏，将前6位和后4位替换为*号
func maskIDCard(idCard string) string {
	re, _ := regexp.Compile(`(\d{6})\d{8}(\w{4})`)
	return re.ReplaceAllString(idCard, "$1********$2")
}
