package server

import (
	"fmt"
	"github.com/mgutz/ansi"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"jumpjumpGo/conf"
	"jumpjumpGo/database"
	"log"
	"strconv"
	"strings"
)

type Client struct {
	Term         *term.Terminal
	windowWidth  int
	windowHeight int
	UserServer   []*database.UserServer
	SSHClient    *SSHClient
}

func NewClient(userId string) *Client {
	c := new(Client)
	conf.Conf.DB.Preload("Server").Where("user_id = ?", userId).Find(&c.UserServer)
	return c
}

func (c *Client) Resize(width, height int) error {
	err := c.Term.SetSize(width, height)
	if err != nil {
		log.Printf("Resize failed: %dx%d", width, height)
		return err
	}
	c.windowWidth, c.windowHeight = width, height
	return nil
}

func (c *Client) HandleShell(channel ssh.Channel) {
	defer channel.Close()

	for {
		line, err := c.Term.ReadLine()
		if err != nil {
			break
		}
		switch line {
		case "exit":
			channel.Close()
		case "list":
			c.Term.Write([]byte(ansi.Color(c.ServerList(), "green")))
		default:
			index, _ := strconv.Atoi(line)
			server := c.GetServerByIndex(index)
			if server != nil {
				c.handleJump(channel, server)
			} else {
				c.Term.Write([]byte("not find\n"))
			}
		}
	}
}

func (c *Client) ServerList() string {
	var tmp strings.Builder
	for i, v := range c.UserServer {
		tmp.WriteString(fmt.Sprintf("\t%d   %s:%d\n", i+1, v.Server.Ip, v.Server.Port))
	}
	return tmp.String()
}
func (c *Client) GetServerByIndex(i int) *database.UserServer {
	if i > len(c.UserServer) || i < 1 {
		return nil
	}
	return c.UserServer[i-1]
}

func inArray(c *Client, cs []*Client) bool {
	for _, v := range cs {
		if c == v {
			return true
		}
	}
	return false
}

func (c *Client) handleJump(channel ssh.Channel, server *database.UserServer) {
	addr := fmt.Sprintf("%s:%d", server.Server.Ip, server.Server.Port)
	var auth ssh.AuthMethod
	if server.PrivateKey != nil {
		var cert = []byte(*server.PrivateKey)
		key, err := ssh.ParsePrivateKey(cert)
		if err == nil {
			auth = ssh.PublicKeys(key)
		} else if err.Error() == "ssh: this private key is passphrase protected" {
			pass, err := c.Term.ReadPassword("This private key is passphrase protected, please enter the certificate password (the password will not be recorded):")
			if err != nil {
				return
			}
			key, err = ssh.ParsePrivateKeyWithPassphrase(cert, []byte(pass))
			if err == nil {
				auth = ssh.PublicKeys(key)
			}
		}
	}
	if auth == nil {
		if server.LoginPassword != nil {
			auth = ssh.Password(*server.LoginPassword)
		}
	}
	remote, err := newSSHClient(addr, server.LoginUserName, auth)
	if err != nil {
		c.Term.Write([]byte("Cannot connect " + addr + "\n"))
		fmt.Println(err)
		return
	}
	c.SSHClient = remote
	remote.jump(channel, uint32(c.windowHeight), uint32(c.windowHeight))
}
