package myssh

import (
	"fmt"
	"testing"
)

func TestSsh(t *testing.T) {
	//bb := []byte("中文")
	//fmt.Println(bb)
	//gbk := Utf8ToGbk("中文")
	//fmt.Println([]byte(gbk))

	server, port, user, pwd := "192.168.35.130", "22", "root", "Qaz@1234"
	s := NewMySsh(server, port, user, pwd)
	s.UseSSH = true
	//s.InitClient()
	//s.InitCommonSession()
	//s.InitCommonTerminal()
	//s.SendCmd("cd shell")
	//withOut, err2 := s.SendCmdWithOut("ls 1.txt")
	//if err2 != nil {
	//	panic(err2)
	//}
	//fmt.Println(withOut)
	withOut, err2 := s.RunCombinedOutput("lsof -h")
	if err2 != nil {
		fmt.Println(withOut)
		panic(err2)
	}
	fmt.Println(withOut)
	return

	//withOut, err2 := s.SendCmdWithOut("cd shell")
	//if err2 != nil {
	//	panic(err2)
	//}
	//fmt.Println(withOut)

	//s.EnableSudo()
	//s.SendCmd("cd shell")
	//ok_lsof, pid := s.ServiceIsOK_netstat("8848")
	//if ok_lsof {
	//	fmt.Println(pid)
	//}
	//return
	//out, err := s.SendCmdWithOut("sh run.sh")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(out)

	//out, err := s.SendCmdWithOut("telnet 192.168.35.130 8848")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(out)
	//继续发送
	//out, err := s.SendCmdWithOut("\x03")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(out)

	//out, err := s.SendCmdWithOut("netstat -tlpn | grep 8848")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(out)
	//处理下结果
	//splits := strings.Split(out, "\r\n")
	//for index, temp := range splits {
	//	fmt.Println(index, temp)
	//}
	//if len(splits) >= 2 { //如果返回的有2个，可能存在PID
	//	//替换空格,临时方法，并不是正式的
	//	compile, _ := regexp.Compile("\\s{2,}")
	//	all := compile.ReplaceAll([]byte(splits[1]), []byte(" "))
	//	for index, temp := range strings.Split(string(all), " ") {
	//		fmt.Println(index, temp)
	//	}
	//}
	//循环监控端口状态
	//out, err := s.SendCmdWithOut("java -jar nacos_test-0.0.1-192.168.35.130.jar")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(out)
}
