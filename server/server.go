package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Rehtt/Kit/i18n"
	"github.com/Rehtt/jumpjumpGo/conf"
	"github.com/Rehtt/jumpjumpGo/database"
	"github.com/Rehtt/jumpjumpGo/util"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh"
	"strconv"
	"strings"
)

func (c *Client) ServerListTable() string {
	var tmp strings.Builder
	table := tablewriter.NewWriter(&tmp)
	table.SetColWidth(c.windowWidth)
	table.SetHeader([]string{"ID", "Alias", "IP", "Port"})
	table.SetHeaderColor(
		tablewriter.Color(tablewriter.FgHiRedColor),
		tablewriter.Color(tablewriter.FgHiRedColor),
		tablewriter.Color(tablewriter.FgHiRedColor),
		tablewriter.Color(tablewriter.FgHiRedColor),
	)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	table.SetColumnColor(
		tablewriter.Color(tablewriter.FgGreenColor),
		tablewriter.Color(tablewriter.FgGreenColor),
		tablewriter.Color(tablewriter.FgGreenColor),
		tablewriter.Color(tablewriter.FgGreenColor),
	)
	err := conf.Conf.DB.Where("user_id = ?", c.user.ID).Find(&c.UserServer).Error
	if err != nil {
		return "DB Error: " + err.Error()
	}
	for i, v := range c.UserServer {
		table.Append([]string{strconv.Itoa(i + 1), v.Alias, v.ServerAddr, v.ServerPort})
	}
	table.Render()
	return tmp.String()
}

func (c *Client) GetServerByIndex(i int) *database.UserServer {
	if i > len(c.UserServer) || i < 1 {
		return nil
	}
	return c.UserServer[i-1]
}

func (c *Client) AddServer() {
	var err error
	defer func() {
		if err != nil {
			c.WriteTermColor(err.Error(), "red")
		}
	}()
	data := &database.UserServer{
		UserID: c.user.ID,
	}
	data.Alias, err = c.Interaction("Alias")
	if err != nil {
		return
	}
	for {
		data.ServerAddr, err = c.Interaction("Server Host or IP")
		if err != nil {
			return
		}
		if data.ServerAddr == "" {
			c.WriteTermColor("Please enter Host or IP correctly\n", "red")
		} else {
			break
		}
	}

	for {
		data.ServerPort, err = c.InteractionDefault("Server Port (Default: 22)", "22")
		if err != nil {
			return
		}
		port, err := strconv.Atoi(data.ServerPort)
		if err != nil || !(port > 0 && port < 0xffff) {
			c.WriteTermColor("Incorrect port number\n", "red")
		} else {
			break
		}
	}

	data.LoginUserName, err = c.InteractionDefault(fmt.Sprintf("Login User (Default:%s)", c.user.Username), c.user.Username)
	if err != nil {
		return
	}
	password, err := c.Interaction(i18n.GetText("Login Password (Skip using certificate login)"), true)
	if err != nil {
		return
	}
	if password == "" {
		var priKey string
		for {
			priKey, err = c.Interaction(i18n.GetText("Private key content, convert to base64 input (encryption certificate needs to enter the key when logging in)"))
			if err != nil {
				return
			}
			if priKey == "" {
				c.WriteTermColor(i18n.GetText("No login authentication\n"), "blue")
			} else {
				k, err := base64.StdEncoding.DecodeString(priKey)
				if err != nil {
					c.WriteTermColor(i18n.GetText("Bad Base64\n"), "red")
					continue
				}
				_, err = ssh.ParsePrivateKey(k)
				if err != nil && err.Error() != "ssh: this private key is passphrase protected" {
					c.WriteTermColor(i18n.GetText("Bad Key Certificate\n"), "red")
					continue
				}
				data.PrivateKey = &priKey
				break
			}
		}
	} else {
		data.LoginPassword = &password
	}
	err = conf.Conf.DB.Save(data).Error
	if err != nil {
		err = errors.New(i18n.GetText("Database storage error: ") + err.Error())
		return
	}
	c.WriteTermColor(i18n.GetText("Added successfully\n"), "green")
}

func (c *Client) DelServer(index string) {
	i, _ := strconv.Atoi(index)
	server := c.GetServerByIndex(i)
	if server == nil {
		c.WriteTermColor(i18n.GetText("not find server\n"), "red")
		return
	}
	err := conf.Conf.DB.Where("id = ?", server.ID).Delete(&database.UserServer{}).Error
	if err != nil {
		c.WriteTermColor(i18n.GetText("DB Error: ")+err.Error(), "red")
	}
	c.WriteTermColor(i18n.GetText("Deleted successfully\n"), "green")
}

func (c *Client) ChangeServer(index string) {
	i, _ := strconv.Atoi(index)
	server := c.GetServerByIndex(i)
	if server == nil {
		c.WriteTermColor(i18n.GetText("not find server\n"), "red")
		return
	}
	var err error
	defer func() {
		if err != nil {
			c.WriteTermColor(err.Error(), "red")
		}
	}()
	server.Alias, err = c.InteractionDefault(fmt.Sprintf(i18n.GetText("Alias (Original: %s)"), server.Alias), server.Alias)
	if err != nil {
		return
	}
	server.ServerAddr, err = c.InteractionDefault(fmt.Sprintf(i18n.GetText("Server Host or IP (Original: %s)"), server.ServerAddr), server.ServerAddr)
	if err != nil {
		return
	}

	var port string
	for {
		port, err = c.InteractionDefault(fmt.Sprintf(i18n.GetText("Server Port (Original: %s)"), server.ServerPort), server.ServerPort)
		if err != nil {
			return
		}
		portn, err := strconv.Atoi(server.ServerPort)
		if err != nil || !(portn > 0 && portn < 0xffff) {
			c.WriteTermColor(i18n.GetText("Incorrect port number\n"), "red")
		} else {
			server.ServerPort = port
			break
		}
	}

	server.LoginUserName, err = c.InteractionDefault(fmt.Sprintf(i18n.GetText("Login User (Original:%s)"), server.LoginUserName), server.LoginUserName)
	if err != nil {
		return
	}

	password, err := c.InteractionDefault(i18n.GetText("Login Password (Skip using certificate login)"), util.String(server.LoginPassword), true)
	if err != nil {
		return
	}
	if password == "" {
		var priKey string
		for {
			priKey, err = c.InteractionDefault(i18n.GetText("Private key content (encryption certificate needs to enter the key when logging in)"), util.String(server.PrivateKey))
			if err != nil {
				return
			}
			if priKey == "" {
				c.WriteTermColor(i18n.GetText("No login authentication\n"), "blue")
			} else {
				_, err := ssh.ParsePrivateKey([]byte(priKey))
				if err != nil && err.Error() != "ssh: this private key is passphrase protected" {
					c.WriteTermColor(i18n.GetText("Bad Key Certificate\n"), "red")
					continue
				}
				server.PrivateKey = &priKey
				server.LoginPassword = nil
				break
			}
		}
	} else {
		server.LoginPassword = &password
		server.PrivateKey = nil
	}
	err = conf.Conf.DB.Where("id = ?", server.ID).Updates(server).Error
	if err != nil {
		err = errors.New(i18n.GetText("Database storage error: ") + err.Error())
		return
	}
	c.WriteTermColor(i18n.GetText("Changed successfully\n"), "green")
}
