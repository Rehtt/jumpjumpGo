package server

import (
	"encoding/base64"
	"fmt"
	"github.com/Rehtt/Kit/i18n"
	"golang.org/x/crypto/ssh"
	"jumpjumpGo/cmd"
	"jumpjumpGo/conf"
	"jumpjumpGo/database"
	"log"
	"strconv"
)

type Client struct {
	*cmd.Term
	windowWidth  int
	windowHeight int
	UserServer   []*database.UserServer
	SSHClient    *SSHClient
	user         *database.User
}

func NewClient(userId string) *Client {
	c := new(Client)
	conf.Conf.DB.Preload("Server").Where("user_id = ?", userId).Find(&c.UserServer)
	c.user = new(database.User)
	conf.Conf.DB.Where("id = ?", userId).Find(c.user)
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
			c.WriteTerm(c.ServerListTable())
		case "add":
			c.AddServer()
		case "del":
			index, _ := c.Interaction("ID")
			c.DelServer(index)
		case "change":
			index, _ := c.Interaction("ID")
			c.ChangeServer(index)
		default:
			index, _ := strconv.Atoi(line)
			server := c.GetServerByIndex(index)
			if server != nil {
				c.handleJump(channel, server)
			} else {
				c.Term.Write([]byte(i18n.GetText("not find\n")))
			}
		}
	}
}

func (c *Client) handleJump(channel ssh.Channel, server *database.UserServer) {
	addr := fmt.Sprintf("%s:%s", server.ServerAddr, server.ServerPort)
	var auth ssh.AuthMethod
	if server.PrivateKey != nil {
		cert, _ := base64.StdEncoding.DecodeString(*server.PrivateKey)
		key, err := ssh.ParsePrivateKey(cert)
		if err == nil {
			auth = ssh.PublicKeys(key)
		} else if err.Error() == i18n.GetText("ssh: this private key is passphrase protected") {
			pass, err := c.Interaction(i18n.GetText("This private key is passphrase protected, please enter the certificate password (the password will not be recorded):"), true)
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
		c.WriteTerm(i18n.GetText("Cannot connect ") + addr + "\n")
		return
	}
	c.SSHClient = remote
	remote.jump(channel, uint32(c.windowHeight), uint32(c.windowHeight))
}
