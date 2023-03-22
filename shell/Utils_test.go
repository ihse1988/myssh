package shell

import (
	"fmt"
	"testing"
)

func Test_Utils(t *testing.T) {
	//"lsof -i:8848"
	utils := Init()
	bak, err := utils.SSH_to_do_out(
		"root",
		"Qaz@1234",
		"192.168.35.130:22",
		[]string{
			//"cd /root/shell/ && java -jar nacos_test-0.0.1-192.168.35.130.jar", //使用此方法反而能阻塞
			"telnet 192.168.35.134 8848",
		})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(bak)
}

func TestUtils_SSh_ServiceIsOK(t *testing.T) {
	c := Init()
	server, port, user, pwd := "192.168.35.130", "22", "root", "Qaz@1234"
	ok, s := c.SSh_ServiceIsOK(user, pwd, fmt.Sprintf("%s:%s", server, port), "8848")
	if ok {
		fmt.Println(s)
	}
}
