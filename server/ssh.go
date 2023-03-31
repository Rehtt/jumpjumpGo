package server

import (
	"context"
	"fmt"
	"github.com/mgutz/ansi"
	"golang.org/x/crypto/ssh"
	"jumpjumpGo/cmd"
	"jumpjumpGo/conf"
	"jumpjumpGo/util"
	"log"
	"net"
	"strconv"
	"sync"
)

func StartSSH(ctx context.Context, addr string, wg *sync.WaitGroup) {
	config := &ssh.ServerConfig{
		ServerVersion:               conf.Conf.SSHServerVersion,
		KeyboardInteractiveCallback: authKeyboard,
		PublicKeyCallback:           authPrivateKeyfunc,
	}
	var err error
	// 服务端密钥
	{
		var keys []ssh.Signer
		for {
			keys, err = parseKey("key")
			if err != nil {
				log.Fatalln("parse error:", err)
			}
			if len(keys) == 0 {
				genKey("key")
			} else {
				break
			}
		}
		log.Println("loading key: " + strconv.Itoa(len(keys)))
		for _, v := range keys {
			config.AddHostKey(v)
		}
	}

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("start ssh server error:", err)
	}
	log.Println("ssh server running")
	wg.Done()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		sshConn, chann, request, err := ssh.NewServerConn(conn, config)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// 处理请求
		go ssh.DiscardRequests(request)
		go handleChannels(sshConn, chann)

	}
}
func handleChannels(sshConn *ssh.ServerConn, channels <-chan ssh.NewChannel) {
	c := NewClient(sshConn.Permissions.CriticalOptions["id"])

	for ch := range channels {
		if t := ch.ChannelType(); t != "session" {
			ch.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}
		channel, requests, err := ch.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		c.Term = cmd.NewTerm(channel, "> ")
		c.WriteTerm(ansi.Color(fmt.Sprintf("    jumpjumpGo - %s:%s\n", conf.Conf.MainVersion, conf.Conf.BuildVersion), "red"))
		c.WriteTerm(c.ServerListTable())
		// todo commands manager
		c.WriteTerm("\ncommands:\nlist\tadd\tdel\tchange\n")

		for req := range requests {
			switch (<-requests).Type {
			case "shell":
				if c.Term != nil {
					go c.HandleShell(channel)
				}
			case "pty-req": //通过如下消息可以让服务器为Session分配一个虚拟终端
				//当客户端的终端窗口大小被改变时，或许需要发送这个消息给服务器。
				width, height, ok := util.ParsePtyRequest(req.Payload)
				if ok {
					err := c.Resize(width, height)
					ok = err == nil
				}
			case "window-change":
				if c.SSHClient != nil {
					c.SSHClient.remoteChannel.SendRequest(req.Type, true, req.Payload)
				}
				width, height, ok := util.ParseWinchRequest(req.Payload)
				if ok {
					err := c.Resize(width, height)
					ok = err == nil
				}
			case "exec":
				//// ssh root@mojotv.cn whoami
				////一旦一个Session被设置完毕，在远端就会有一个程序被启动。这个程序可以是一个Shell，也可以时一个应用程序或者是一个有着独立域名的子系统。
				//command, err := c.ParseCommandLine(req) // 协议 req.Payload 里面的用户命令输出
				//if err != nil {
				//	log.Printf("error parsing ssh execMsg: %s\n", err)
				//	return
				//} else {
				//	ok = true
				//}
				////开始执行从 whoami 远程shell 命令
				//// 执行完成 结果直接返回
				//go c.HandleExec(command, channel)
			case "env":
				//在shell或command被开始时之后，或许有环境变量需要被传递过去。然而在特权程序里不受控制的设置环境变量是一个很有风险的事情，
				//所以规范推荐实现维护一个允许被设置的环境变量列表或者只有当sshd丢弃权限后设置环境变量。
				//todo set language i18n
				log.Print(string(req.Payload))
			case "subsystem":
				////一旦一个Session被设置完毕，在远端就会有一个程序被启动。这个程序可以是一个Shell，也可以时一个应用程序或者是一个有着独立域名的子系统。
				//// 实现一下功能可以实现 sftp功能
				////fmt.Fprintf(debugStream, "Subsystem: %s\n", req.Payload[4:])
				//if string(req.Payload[4:]) == "sftp" {
				//	ok = true
				//	go c.HandleSftp(channel)
				//}

			default:
				log.Println(req.Type, string(req.Payload))
			}
			if req.WantReply {
				req.Reply(true, nil)
			}
		}
	}
}

type SSHClient struct {
	serverConn    *ssh.Client
	remoteChannel ssh.Channel
}

func (s *SSHClient) jump(userChann ssh.Channel, w, h uint32) {
	defer func() {
		s.remoteChannel.Close()
		s.serverConn.Close()
	}()
	err := util.SendPtyReq(s.remoteChannel, w, h)
	if err != nil {
		log.Fatalln(err)
	}
	ok, err := s.remoteChannel.SendRequest("shell", true, nil)
	if err == nil && !ok {
		log.Fatalln(err)
	}
	go util.IoCopy(s.remoteChannel, userChann)
	util.IoCopy(userChann, s.remoteChannel)
}

func newSSHClient(remoteAddr, user string, auth ssh.AuthMethod) (remote *SSHClient, err error) {
	config := &ssh.ClientConfig{
		User:            user,
		ClientVersion:   conf.Conf.SSHClientVersion,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if auth != nil {
		config.Auth = []ssh.AuthMethod{auth}
	}
	conn, err := ssh.Dial("tcp", remoteAddr, config)
	if err != nil {
		return
	}
	remoteChannel, remoteRequests, err := conn.OpenChannel("session", nil)
	if err != nil {
		conn.Close()
		return
	}
	go ssh.DiscardRequests(remoteRequests)
	remote = &SSHClient{
		serverConn:    conn,
		remoteChannel: remoteChannel,
	}
	return
}
