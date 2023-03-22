package shell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Utils struct{}

func Init() *Utils {
	return new(Utils)
}

func (c *Utils) SSH_to_do(user, password, host string, cmdList []string) {
	fmt.Println(host, cmdList)

	passWd := []ssh.AuthMethod{ssh.Password(password)}
	conf := ssh.ClientConfig{
		User:            user,
		Auth:            passWd,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
	}

	client, err := ssh.Dial("tcp", host, &conf)
	if err != nil {
		fmt.Println(fmt.Sprintf("出现错误:%s", err))
		return
	}
	defer client.Close()
	for _, cmd := range cmdList {
		if session, err := client.NewSession(); err == nil {
			defer session.Close()
			session.Stdout = os.Stdout
			session.Stderr = os.Stderr
			session.Run(cmd) //run是阻塞的
		}
	}
}

// /有返回值
func (c *Utils) SSH_to_do_out(user, password, host string, cmdList []string) (string, error) {
	fmt.Println(host, cmdList)

	passWd := []ssh.AuthMethod{ssh.Password(password)}
	conf := ssh.ClientConfig{
		User:            user,
		Auth:            passWd,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
	}

	client, err := ssh.Dial("tcp", host, &conf)
	if err != nil {
		fmt.Println(fmt.Sprintf("出现错误:%s", err))
		return "", err
	}
	var outline string
	//上面的代码每次都创建新的session导致每次都是新的环境，修改为下面的
	defer client.Close()
	if session, err := client.NewSession(); err == nil {
		stdout, err1 := session.StdoutPipe()
		if err1 != nil {
			fmt.Println(fmt.Sprintf("执行出现错误:%s", err1))
			fmt.Println(err1)
			return "", err1
		}
		reader := bufio.NewReader(stdout)
		for _, cmd := range cmdList {
			//session.Stdout = os.Stdout
			//session.Stderr = os.Stderr
			fmt.Println(fmt.Sprintf("执行：%s", cmd))
			session.Start(cmd)
			for {
				line, err2 := reader.ReadString('\n')
				if err2 != nil || io.EOF == err2 {
					break
				}
				outline = outline + line
				//fmt.Println(line)
			}
			session.Wait()
			fmt.Println(fmt.Sprintf("执行：%s完成", cmd))
			return outline, nil
		}
		return outline, nil
	}
	return outline, nil
}

// 判断端口是否正常，使用远程执行ssh的方式
func (c *Utils) SSh_ServiceIsOK(user, password, host string, serverport string) (bool, string) {
	out, err := c.SSH_to_do_out(user, password, host, []string{fmt.Sprintf("lsof -i:%s", serverport)})
	if err != nil {
		panic(err)
	}
	splits := strings.Split(out, "\n")
	for index, temp := range splits {
		//fmt.Println(index, temp)
		if strings.IndexAny(temp, fmt.Sprintf("*:%s", serverport)) > 0 { //首先要包含 *:端口
			//替换空格,临时方法，并不是正式的
			compile, _ := regexp.Compile("\\s{2,}")
			all := compile.ReplaceAll([]byte(splits[index]), []byte(" "))
			return true, strings.Split(string(all), " ")[1]
		}
	}
	return false, ""

}

// netstat -tlpn | grep 8843
func (c *Utils) SSh_ServiceIsOKN(user, password, host string, serverport string) (isok bool, pid string) {
	out, err := c.SSH_to_do_out(user, password, host, []string{fmt.Sprintf("netstat -tlpn | grep %s", serverport)})
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	splits := strings.Split(out, "\n")
	for index, temp := range splits {
		fmt.Println(index, temp)
		if strings.IndexAny(temp, fmt.Sprintf(":::%s", serverport)) > 0 { //首先要包含 *:端口
			//替换空格,临时方法，并不是正式的
			compile, _ := regexp.Compile("\\s{2,}")
			all := compile.ReplaceAll([]byte(splits[index]), []byte(" "))
			split := strings.Split(string(all), " ")
			return true, strings.Split(split[len(split)-2], "/")[0]
		}
	}
	return false, ""

}

func (c *Utils) ServiceIsOkTelnet() {

}

/*
port-端口
host-主机
user-用户名
src-源路径
dist-目标路径
*/
func (c *Utils) SCP(port, host, user, src, dist string) bool {
	//exec.Command(
	//	"scp",
	//	fmt.Sprintf("-P %s", port),
	//	src,
	//	fmt.Sprintf("%s@%s:%s", user, host, dist))
	//修改为阻塞调用
	//wait, _ := c.CmdWait("scp", []string{"-P", port, src, fmt.Sprintf("%s@%s:%s", user, host, dist)})
	//return wait
	//修改为原始
	body, err := c.Cmd("scp", "-P", port, src, fmt.Sprintf("%s@%s:%s", user, host, dist))
	if err != nil {
		log.Println(fmt.Sprintf("执行错误,返回：%s", body), "错误消息", err)
		return true
	}
	log.Println(fmt.Sprintf("执行成功,返回：%s", body))
	return false
	//return fmt.Sprintf("scp -P %s %s %s@%s:%s", port, src, user, host, dist)
}

func (c *Utils) MV(src, dist string) {

}

func (c *Utils) Pipleline(cmdlist []string) []string {
	return cmdlist
}

func (c *Utils) CMD(cmd string) {
	sh := exec.Command(cmd)
	e := sh.Run()
	if e != nil {
		log.Println(e)
	}
}

// 后台执行命令，这里暂时用来启动java程序的脚本的shell，使用error来判定是否错误
func (c *Utils) cmdBlackGroud(cmd string) error {
	ctx, cancel := context.WithCancel(context.Background())
	go func(cancelFunc context.CancelFunc) {
		time.Sleep(3 * time.Second)
		cancelFunc()
	}(cancel)

	c1 := exec.CommandContext(ctx, "bash", "-c", cmd) // mac linux
	stdout, err := c1.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c1.StderrPipe()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	// 因为有2个任务, 一个需要读取stderr 另一个需要读取stdout
	wg.Add(2)
	go read(ctx, &wg, stderr)
	go read(ctx, &wg, stdout)
	// 这里一定要用start,而不是run 详情请看下面的图
	err = c1.Start()
	//time.Sleep(5000)
	// 等待任务结束
	wg.Wait()
	return err
}

func (c *Utils) Cmd(commandName string, params ...string) (out string, err error) {
	cmd := exec.Command(commandName, params...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // 标准输出
	cmd.Stderr = &stderr // 标准错误
	err1 := cmd.Run()
	log.Println("执行命令", cmd.Args)
	outStr, _ := string(stdout.Bytes()), string(stderr.Bytes())
	//fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
	if err1 != nil {
		//log.Fatalf("cmd.Run() failed with %s\n", err)
		return "", err1
	}
	return outStr, err1
}

// 阻塞式调用，这里仅先用来做SCP命令
func (c *Utils) CmdWait(commandName string, params []string) (isok bool, body string) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command(commandName, params...)
	//显示运行的命令
	fmt.Println(fmt.Sprintf("准备执行:%s", cmd.Args))
	//StdoutPipe方法返回一个在命令Start后与命令标准输出关联的管道。Wait方法获知命令结束后会关闭这个管道，一般不需要显式的关闭该管道。
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(fmt.Sprintf("执行出现错误:%s", cmd.Args))
		fmt.Println(err)
		return false, ""
	}
	cmd.Start()
	//创建一个流来读取管道内内容，这里逻辑是通过一行一行的读取的
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	var outBody string
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		// 转换到gbk编码
		//srcCoder := mahonia.NewDecoder("gbk")
		//resultstr := bytes2str(out)
		//result := srcCoder.ConvertString(resultstr)
		outBody = line

		fmt.Println(line)
	}
	//阻塞直到该命令执行完成，该命令必须是被Start方法开始执行的
	cmd.Wait()
	fmt.Println(fmt.Sprintf("执行完成:%s", cmd.Args))
	return true, outBody
}

// func (c *Utils) CMDSCP(port, host, user, src, dist string){
// exec.Command()
// }

// 处理IO
func read(ctx context.Context, wg *sync.WaitGroup, std io.ReadCloser) {
	reader := bufio.NewReader(std)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			fmt.Print(readString)
		}
	}
}
