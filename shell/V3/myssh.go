package myssh

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

var ESC_pattern string = "^[ -/]*([0-Z\\-~]|\\[[ -/]*[0-?]*[@-~])$" //重写全定义
// V3版本 ，buff里面应该是session不重连则会一直从开始读取，无法rest，暂时使用后置处理的方式
type MySsh struct {
	SshBuffer   *SshBuffer
	SshTerminal *SshTerminal
	Client      *ssh.Client
	Session     *ssh.Session
	Start       bool
	User        string
	Pwd         string
	Server      string
	Port        string
	//使用公钥
	UseSSH bool

	byteslist []byte
	count     int
	lastcount int
	RootPwd   string
}

type SshBuffer struct {
	outBuf   *bytes.Buffer
	stdinBuf io.WriteCloser
}

type SshTerminal struct {
	in  chan string
	out chan string
}

func (c *MySsh) Close() {
	if c.Session != nil {
		defer c.Session.Close()
	}
	if c.Client != nil {
		defer c.Client.Close()
	}
}

// 构造SSH
// 构造完成后需要再次执行InitCommonTerminal 进入自定义终端，自定义终端可以使用SendCmd和SendCmdWithOut来持续操作
// 也可以使用另外的命令执行操作。
// 如果rootPwd密码为空，则不切换为root，如果root密码不为空则执行切换，这个逻辑有点奇怪，需要修改
// 2023/03/22
func NewMySsh(server string, port string, user string, pwd string, rootPwd string) *MySsh {
	var stdinBuf io.WriteCloser
	return &MySsh{
		SshBuffer: &SshBuffer{
			bytes.NewBuffer(make([]byte, 0)),
			stdinBuf,
		},
		SshTerminal: &SshTerminal{
			make(chan string, 1),
			make(chan string, 1),
		},
		Server:  server,
		Port:    port,
		User:    user,
		Pwd:     pwd,
		RootPwd: rootPwd,
	}
}

func (c *MySsh) InitCommonTerminal() error {
	if c.Start {
		return fmt.Errorf("session is start terminal")
	}
	err := c.InitCommonSession()
	if err != nil {
		return err
	}
	session := c.Session
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	//term vt100,xtrem  貌似对比输入都可以。。。需要对字符串进行过滤操作
	if err = session.RequestPty("vt100", 18, 41, modes); err != nil {
		fmt.Printf("get pty error:%v\n", err)
		return err
	}
	stdinBuf, err := session.StdinPipe()
	if err != nil {
		log.Printf("get stdin pipe error%v\n", err)
		return err
	}
	c.SshBuffer.stdinBuf = stdinBuf
	session.Stdout = c.SshBuffer.outBuf

	err = session.Shell()
	if err != nil {
		fmt.Printf("shell session error%v", err)
		return err
	}
	ch := make(chan struct{})
	time.Sleep(time.Millisecond * 200)
	if c.User == "root" {
		go resetOutBuf(c.SshBuffer.outBuf, ch, '#')
	} else {
		go resetOutBuf(c.SshBuffer.outBuf, ch, '$')
	}
	<-ch
	log.Println("after reset buf", c.SshBuffer.outBuf.String())
	c.Start = true
	log.Println("start wait session")
	//切换为root
	c.EnableSudo()
	//在这里发送一条
	c.SendCmd("cd")

	go session.Wait()
	return err
}

// 使用新的session执行命令，命令可以合并为多个
func (c *MySsh) RunCombinedOutput(cmd string) (string, error) {
	if c.Client == nil {
		if err := c.InitClient(); err != nil {
			return "", err
		}
	}
	//log.Println(fmt.Sprintf("newSession执行：%s", cmd))
	session, err := c.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create new session error: %w", err)
	}
	defer session.Close()
	buf, err := session.CombinedOutput(cmd)
	return string(buf), err
}
func (c *MySsh) RunOutput(cmd string) (string, error) {
	if c.Client == nil {
		if err := c.InitClient(); err != nil {
			return "", err
		}
	}
	//log.Println(fmt.Sprintf("newSession执行：%s", cmd))
	session, err := c.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create new session error: %w", err)
	}
	defer session.Close()
	buf, err := session.Output(cmd)
	return string(buf), err
}

func resetOutBuf(outBuf *bytes.Buffer, ch chan struct{}, terminator byte) {
	buf := make([]byte, 8192)
	var t int
	for {
		if outBuf != nil {
			_, err := outBuf.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("read out buffer err:%v", err)
				break
			}
			t = bytes.LastIndexByte(buf, terminator)
			if t > 0 {
				ch <- struct{}{}
				break
			}
		}
	}
}

// 读取本地公钥
func (c *MySsh) readSshRsaKey() (ssh.Signer, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	key, err := os.ReadFile(path.Join(homePath, ".ssh", "id_rsa"))
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return signer, err
}

func (c *MySsh) InitCommonSession() error {
	c.InitClient()
	session, err := c.Client.NewSession()
	if err != nil {
		fmt.Printf("get session error %v\n", err)
		return err
	}
	c.Session = session
	return nil
}

// 触发这个执行普通的SHH
func (c *MySsh) InitClient() error {
	myAuthMethod := []ssh.AuthMethod{ssh.Password(c.Pwd)}
	if c.UseSSH {
		log.Println("尝试使用公钥登陆")
		signer, err := c.readSshRsaKey()
		if err != nil {
			log.Println("公钥读取失败，继续使用密码登陆错误消息", err)
		} else {
			myAuthMethod = []ssh.AuthMethod{ssh.PublicKeys(signer)}
		}
	} else {
		log.Println("密码登陆")
	}
	config := &ssh.ClientConfig{
		User:            c.User,
		Auth:            myAuthMethod,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", c.Server, c.Port), config)
	if err != nil {
		fmt.Printf("dial ssh error :%v\n", err)
		return err
	}
	c.Client = client
	c.User = config.User
	return nil
}

func (c *MySsh) getChar() (byte, error) {
	readByte, err := c.SshBuffer.outBuf.ReadByte()
	c.count += 1
	return readByte, err
	//b := make([]byte, 1)
	//read, err := c.SshBuffer.outBuf.Read(b)
	//if read > 0 {
	//	return b[0], err
	//}
	//return 0, err
}

func (c *MySsh) handleESCSequences() {
	i := 0
	mlist := make([]byte, 0)
	for {
		x, _ := c.getChar()
		c.count -= 1 //这些要排除了，不然rang超了
		if x == 27 {
			continue
		}
		mlist = append(mlist, x)
		match, _ := regexp.MatchString(ESC_pattern, string(mlist))
		if match {
			break
		}
		i = i + 1
		if i > 20 {
			fmt.Println("We meet invalid Escape Sequences, skip the first ESC")
			//pushChar(mlist)
			c.count -= 1
			c.SshBuffer.outBuf.UnreadByte()
			break
		}
	}
}

func (c *MySsh) pushChar() {
	//c.SshBuffer.outBuf.UnreadRune()
	c.count -= 1
	c.SshBuffer.outBuf.UnreadByte()
}
func (c *MySsh) isASCII(b byte) bool {
	if 0x20 <= b && b <= 0x7f {
		return true
	}

	if b == 0x0a || b == 0x0d {
		return true
	}

	return false
}

func (c *MySsh) BytesComin(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}
func (c *MySsh) internWrites(b []byte) {
	c.byteslist = c.BytesComin(c.byteslist, b)
	//byteslist = bytes.Join(byteslist, []byte(s))
	//byteslist = append(byteslist, []byte(s))
}

func (c *MySsh) internWrite(b byte) {
	c.byteslist = append(c.byteslist, b)
	//byteslist = BytesComin(byteslist, [] byteb)
	//byteslist = bytes.Join(byteslist, []byte(s))
	//byteslist = append(byteslist, []byte(s))
}
func (c *MySsh) handleASCIISequences() {
	b, err2 := c.getChar()
	tempbytes := make([]byte, 0)
	if (b & 0x80) != 0 {
		c.internWrite(b)
		b, _ := c.getChar()
		c.internWrite(b)
		//internWrite(new String(new byte[]{b, getChar()}));
	} else if c.isASCII(b) {
		for c.isASCII(b) && io.EOF != err2 {
			tempbytes = append(tempbytes, b)
			b, err2 = c.getChar()
			if err2 == io.EOF {
				break
			}
		}

		if c.isASCII(b) {
			tempbytes = append(tempbytes, b)
			c.internWrites(tempbytes)
			//byteslist = append(byteslist, tempbytes)
		} else {
			c.internWrites(tempbytes)
			c.pushChar()
		}
	}
	//c.SshBuffer.outBuf.E
}

// 使用上一个session继续执行命令
func (c *MySsh) listenMessages(terminator int32) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		cmd := <-c.SshTerminal.in
		_, _ = c.SshBuffer.stdinBuf.Write([]byte(fmt.Sprintf("%v\n", cmd)))
		//fmt.Printf("send cmd %v\n", cmd)
		log.Println(fmt.Sprintf("执行:%v", cmd))
		wg.Done()
	}()
	wg.Wait()
	var wg1 sync.WaitGroup
	wg1.Add(1)

	//timer := time.AfterFunc(10*time.Second, myFunc) //使用这种方式比较low
	go func() {
		//buf := make([]byte, 8192)
		c.byteslist = make([]byte, 0)
		var t int
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		//这里还需要一个定时器
		//ESC_pattern := "[ -/]*([0-Z\\-~]|\\[[ -/]*[0-?]*[@-~])"  //和java不一样，java严格匹配
		//判断
		for {
			time.Sleep(time.Millisecond * 200)
			//n, err := c.SshBuffer.outBuf.Read(buf)
			//重写读的方法排除乱码
			select {
			case <-ctx.Done():
				c.SshTerminal.out <- string(c.byteslist)
				wg1.Done()
				//超时退出，这里无论如何都返回并且
				return
			default:
				for {
					b, err := c.getChar()
					if err != nil {
						break
					}
					//fmt.Println(b)  //保留debug
					if b == 0 {
						continue
					}
					// ESC
					if b == 0x1b {
						c.handleESCSequences()
						continue
					}
					if b == 0x07 { // bel ^G
						continue
					}
					if b == 0x08 {
						continue
					}
					if b == 0x09 { // ht(^I)
						continue
					}

					if b == 0x0f { // rmacs ^O // end alternate character set (P)
						continue
					}

					if b == 0x0e { // smacs ^N // start alternate character set
						// (P)
						continue
					} else {
						c.pushChar()
						c.handleASCIISequences()
						continue
					}
				}

				//if err != nil && err != io.EOF {
				//	fmt.Printf("read out buffer err:%v", err)
				//	break
				//}
				//fmt.Println(n, terminator, t)
				//compile, err := regexp.Compile("(\\x1b\\[|\\x9b)[^@-_]*[@-_]|\\x1b[@-_]")
				//all := compile.ReplaceAll(buf, []byte(""))
				//c.SshTerminal.out <- string(all)
				//break
				if len(c.byteslist) > 0 {
					t = bytes.LastIndexByte(c.byteslist, byte(terminator))
					if t > 0 {
						//fmt.Println("t>0", c.byteslist[:t])
						//fmt.Println("转string", string(c.byteslist))
						//设计为永远获取最后一条
						//取倒数第二个#号到用户之间的数据
						t1 := bytes.LastIndexByte(c.byteslist[0:t], byte(terminator))
						t2 := strings.LastIndex(string(c.byteslist[0:t]), "["+c.User) //- len(c.User) - 1
						//fmt.Println("测试：", string(c.byteslist))
						if t2 > t1 && t1 > 0 {

							c.SshTerminal.out <- string(c.byteslist[t1+1 : t2])
						} else {
							c.SshTerminal.out <- string(c.byteslist[t:])
						}
						break
					}
				} else {
					//fmt.Println(c.byteslist)
					//fmt.Println("转string", string(c.byteslist))
					c.SshTerminal.out <- string(c.byteslist)
					break
				}
			}
			//fmt.Println(fmt.Sprintf("读取个数%d", c.count))
			wg1.Done()
		}

	}()
	wg1.Wait()
	return nil
}

func (c *MySsh) EnableSudo() error {
	err := c.SendCmdTerminator("su", ':')
	c.User = "root"
	err = c.SendCmd(c.RootPwd)
	return err
}

func (c *MySsh) SendCmd(cmd string) error {
	c.SshTerminal.in <- cmd
	terminator := '$'
	if c.User == "root" {
		terminator = '#'
	}
	err := c.listenMessages(terminator)
	<-c.SshTerminal.out
	return err
}

func (c *MySsh) SendCmdTerminator(cmd string, terminator int32) error {
	c.SshTerminal.in <- cmd
	err := c.listenMessages(terminator)
	<-c.SshTerminal.out
	return err
}

// 使用上一个session继续执行命令
func (c *MySsh) SendCmdWithOut(cmd string) (string, error) {
	c.SshTerminal.in <- cmd
	terminator := '$'
	if c.User == "root" {
		terminator = '#'
	}
	err := c.listenMessages(terminator)
	out := <-c.SshTerminal.out
	//fmt.Println(out)
	return strings.TrimSpace(strings.Split(out, "["+c.User)[0]), err
}

// 判断端口是否正常，使用远程执行ssh的方式
func (c *MySsh) ServiceIsOK_lsof(serverport string) (isok bool, pid string) {
	out, err := c.SendCmdWithOut(fmt.Sprintf("lsof -i:%s", serverport))
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
func (c *MySsh) ServiceIsOK_netstat(serverport string) (isok bool, pid string) {
	out, err := c.SendCmdWithOut(fmt.Sprintf("netstat -tlpn | grep %s", serverport))
	if err != nil {
		panic(err)
	}
	//fmt.Println(out)
	splits := strings.Split(out, "\n")
	for index, temp := range splits {
		//fmt.Println(index, temp)
		if strings.IndexAny(temp, fmt.Sprintf(":::%s", serverport)) > 0 { //首先要包含 *:端口
			//替换空格,临时方法，并不是正式的
			compile, _ := regexp.Compile("\\s{2,}")
			all := compile.ReplaceAll([]byte(splits[index]), []byte(" "))
			split := strings.Split(string(all), " ")
			//fmt.Println(split)
			return true, strings.Split(split[len(split)-1], "/")[0]
		}
	}
	return false, ""

}

// 使用进阶版本
func (c *MySsh) ServiceIsOK_netstatAd(serverport string) (isok bool, pid string) {
	time.Sleep(3 * time.Second)
	out, err := c.SendCmdWithOut(fmt.Sprintf("netstat -tlpn | grep %s | awk '{print $7}' | awk -F\"/\" '{ print $1}'", serverport))
	if err != nil {
		panic(err)
	}
	if out != "" {
		return true, out
	}
	return false, ""

}
