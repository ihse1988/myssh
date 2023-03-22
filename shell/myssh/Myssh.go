package myssh

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Myssh struct {
	Client  *ssh.Client
	Session *ssh.Session
}

func InitM() *Myssh {
	return new(Myssh)
}

func (c *Myssh) Conn(user, password, host string) error {
	passWd := []ssh.AuthMethod{ssh.Password(password)}
	conf := ssh.ClientConfig{
		User:            user,
		Auth:            passWd,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
	}
	client, err := ssh.Dial("tcp", host, &conf)
	if err != nil {
		fmt.Println(fmt.Sprintf("出现错误:%s", err))
		return err
	}
	c.Client = client
	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("get session error %v\n", err)
		return err
	}
	c.Session = session
	return nil
}

func (c *Myssh) Cmd(cmd string) (string, error) {
	output, err := c.Session.Output(cmd)
	return string(output), err
}

func (c *Myssh) SSH_to_do_wait(user, password, host string, cmdList []string) {
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
	//上面的代码每次都创建新的session导致每次都是新的环境，修改为下面的
	defer client.Close()
	if session, err := client.NewSession(); err == nil {
		for _, cmd := range cmdList {
			output, err1 := session.Output(cmd)
			if err1 != nil {
				fmt.Println(err1)
			}
			fmt.Println(fmt.Sprintf("执行：%s完成%s", cmd, output))
		}
	}
}

func (c *Myssh) SSH_to_do_wait_bak(user, password, host string, cmdList []string) {
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
	//上面的代码每次都创建新的session导致每次都是新的环境，修改为下面的
	defer client.Close()
	if session, err := client.NewSession(); err == nil {
		stdout, err1 := session.StdoutPipe()
		if err1 != nil {
			fmt.Println(fmt.Sprintf("执行出现错误:%s", err1))
			fmt.Println(err1)
			return
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
				fmt.Println(line)
			}
			session.Wait()
			fmt.Println(fmt.Sprintf("执行：%s完成", cmd))
		}

	}
}

/*
port-端口
host-主机
user-用户名
src-源路径
dist-目标路径
*/
func (c *Myssh) SCP(port, host, user, src, dist string) bool {
	//exec.Command(
	//	"scp",
	//	fmt.Sprintf("-P %s", port),
	//	src,
	//	fmt.Sprintf("%s@%s:%s", user, host, dist))
	//修改为阻塞调用
	return c.cmdWait("scp", []string{"-P", port, src, fmt.Sprintf("%s@%s:%s", user, host, dist)})
	//return fmt.Sprintf("scp -P %s %s %s@%s:%s", port, src, user, host, dist)
}

func (c *Myssh) MV(src, dist string) {

}

func (c *Myssh) Pipleline(cmdlist []string) []string {
	return cmdlist
}

func (c *Myssh) CMD(cmd string) {
	sh := exec.Command(cmd)
	e := sh.Run()
	if e != nil {
		log.Println(e)
	}
}

// 后台执行命令，这里暂时用来启动java程序的脚本的shell，使用error来判定是否错误
func (c *Myssh) cmdBlackGroud(cmd string) error {
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

// 阻塞式调用，这里仅先用来做SCP命令
func (c *Myssh) cmdWait(commandName string, params []string) bool {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command(commandName, params...)
	//显示运行的命令
	fmt.Println(fmt.Sprintf("准备执行:%s", cmd.Args))
	//StdoutPipe方法返回一个在命令Start后与命令标准输出关联的管道。Wait方法获知命令结束后会关闭这个管道，一般不需要显式的关闭该管道。
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(fmt.Sprintf("执行出现错误:%s", cmd.Args))
		fmt.Println(err)
		return false
	}
	cmd.Start()
	//创建一个流来读取管道内内容，这里逻辑是通过一行一行的读取的
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		// 转换到gbk编码
		//srcCoder := mahonia.NewDecoder("gbk")
		//resultstr := bytes2str(out)
		//result := srcCoder.ConvertString(resultstr)

		fmt.Println(line)
	}
	//阻塞直到该命令执行完成，该命令必须是被Start方法开始执行的
	cmd.Wait()
	fmt.Println(fmt.Sprintf("执行完成:%s", cmd.Args))
	return true
}

// func (c *Myssh) CMDSCP(port, host, user, src, dist string){
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
