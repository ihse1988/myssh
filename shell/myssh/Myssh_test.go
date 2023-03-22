package myssh

import (
	"fmt"
	"testing"
)

func TestMyssh(t *testing.T) {
	fmt.Println("123")
	myssh := InitM()
	ip := "192.168.35.130:22"
	user := "root"
	pwd := "Qaz@1234"
	myssh.Conn(user, pwd, ip)

	cmd, err := myssh.Cmd("cd shell")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cmd)
	cmd, err = myssh.Cmd("echo 中文")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cmd)
}
