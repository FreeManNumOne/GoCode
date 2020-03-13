package main

/*
用go语言重写发布的脚本
主要用于配合jenkins服务发布
version: 0.1
auth: FreeMan
date: 20200310
*/
import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

//Cli 初始化连接服务器的信息是结构体类型
type Cli struct {
	user        string //登录操作系统的用户
	pwd         string //操作系统密码
	addr        string //操作系统地址需要把ssh 连接的端口也同时写进来
	serviceName string //Service  Name
	jarName     string //Jar package name
	client      *gossh.Client
	session     *gossh.Session
	LastResult  string
}

//Connect 建立SSH连接函数，结构体连接的方法
func (c *Cli) Connect() (*Cli, error) {
	config := &gossh.ClientConfig{}
	config.SetDefaults()
	config.User = c.user
	config.Auth = []gossh.AuthMethod{gossh.Password(c.pwd)}
	config.HostKeyCallback = func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }
	client, err := gossh.Dial("tcp", c.addr, config)
	if nil != err {
		return c, err
	}
	c.client = client
	return c, nil
}

//Run 远程运行操作系统命令
func (c Cli) Run(shell string) (string, error) {
	if c.client == nil {
		if _, err := c.Connect(); err != nil {
			return "", err
		}
	}
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	buf, err := session.CombinedOutput(shell)
	c.LastResult = string(buf)
	return c.LastResult, err
}

//启动服务
func (c Cli) start() {
	cmd := "supervisorctl start " + c.serviceName
	ret, err := c.Run(cmd)
	if err != nil {
		fmt.Println("start failed", err)
		os.Exit(0)
	}
	fmt.Println(ret)
	//判断服务之前是否有启动
	if strings.Contains(ret, "ERROR") {
		fmt.Println(c.serviceName, " service not run")
		os.Exit(0)
	}
}

//停止目标服务
func (c Cli) stop() {
	cmd := "supervisorctl stop " + c.serviceName
	ret, err := c.Run(cmd)
	if err != nil {
		fmt.Println("stop failed", err)
		os.Exit(0)
	}
	fmt.Println(ret)
	if strings.Contains(ret, "ERROR") {
		fmt.Println(c.serviceName, " before service not run")
	}
}

func sftpconnect(user, password, host string, port int) (*sftp.Client, error) {
	var (
		auth         []gossh.AuthMethod
		addr         string
		clientConfig *gossh.ClientConfig
		sshClient    *gossh.Client
		sftpClient   *sftp.Client
		err          error
	)

	// get auth method
	auth = make([]gossh.AuthMethod, 0)
	auth = append(auth, gossh.Password(password))
	clientConfig = &gossh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
	}
	//自动加载ssh key
	clientConfig.HostKeyCallback = func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = gossh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}

//translef parckage jar
func (c Cli) fileTranself() {
	var (
		err        error
		sftpClient *sftp.Client
	)
	// ssh sftp creater conn
	sftpRemoteHost := c.addr
	//get  remote host IP
	sftpRemoteHost2 := strings.Split(sftpRemoteHost, ":")
	sftpRemoteHost = sftpRemoteHost2[0]
	fmt.Println(sftpRemoteHost)
	sftpClient, err = sftpconnect(c.user, c.pwd, sftpRemoteHost, 22)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpClient.Close()

	// 用来测试的本地文件路径 和 远程机器上的文件夹
	var localFilePath = "/data/package/" + c.jarName
	serviceNamePath := strings.Split(c.serviceName, ".")
	var remoteDir = "/data/" + serviceNamePath[0]
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	var remoteFileName = path.Base(localFilePath)
	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	buf := make([]byte, 4096)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf[0:n])
	}

	fmt.Println("copy file to remote server finished!")
}

//传输完成jar包进行md5校验
func (c Cli) md5Package() {
	cmd := "md5sum /data/" + c.serviceName + "/" + c.jarName
	fmt.Println(cmd)
	ret, err := c.Run(cmd)
	if err != nil {
		fmt.Println("md5sum  failed", err)
		os.Exit(0)
	}
	fmt.Println(ret)
	//把目标服务器的md5码取回并与本地的md5码进行对比
	localPackageFile := "/data/package/" + c.jarName
	h := md5.New()
	f, err := os.Open(localPackageFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	io.Copy(h, f)
	getLocalMd5 := hex.EncodeToString(h.Sum(nil))
	if strings.Contains(ret, getLocalMd5) {
		fmt.Println(c.jarName, "md5sum check sucess ")
	} else {
		fmt.Println(c.jarName, "md5sum check failed")
		os.Exit(0)
	}
}

//检查服务是否启动完成
func (c Cli) checkService() {}

func main() {
	paramater := os.Args
	host := paramater[1]
	jarName := paramater[2]
	serviceName := strings.Split(jarName, ".")
	hosts := strings.Split(host, ":")
	for _, v := range hosts {
		remoteHost := v + ":22"
		cli := Cli{
			user:        "root",
			pwd:         "k8s123456",
			addr:        remoteHost,
			serviceName: serviceName[0],
			jarName:     jarName,
		}
		fmt.Println(cli)
		cli.stop()
		cli.fileTranself()
		cli.md5Package()
		cli.start()
		//cli.checkService()  get code 200
	}
}
