package main

import (
	myssh "com.ihse/anyssh/shell/V3"
	"fmt"
)

func main() {
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
