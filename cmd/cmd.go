package cmd

import (
	"context"
	"errors"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"jumpjumpGo/conf"
	"jumpjumpGo/database"
	"jumpjumpGo/util"
	"jumpjumpGo/util/i8n"
	"os"
	"strconv"
	"strings"
)

func StartLocalCMD(ctx context.Context, ch context.CancelCauseFunc) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)
	term := NewTerm(util.NewRW(os.Stdin, os.Stdout), "> ")

	term.WriteTerm(i8n.Dictionary.CMD.Title)
	term.WriteTerm("commands:\nlist\tadd\tdel\tchange\n")
	for {
		cmd, err := term.ReadLine()
		if err != nil {
			return
		}
		switch cmd {
		case "list":
			term.WriteTerm(userList())
		case "add":
			addUser(term)
		case "del":
		case "change":

		case "exit":
			ch(errors.New("jumpjumpGo shutdown"))
			return
		}
	}
}

func addUser(term *Term) {
	var user = new(database.User)
	var err error
	defer func() {
		if err != nil {
			term.WriteTermColor(err.Error()+"\n", "red")
		}
	}()
	for {
		user.Username, err = term.Interaction("User Name")
		if err != nil {
			return
		}
		if user.Username == "" {
			term.WriteTermColor("Please enter the correct user name\n", "red")
		} else {
			if conf.Conf.DB.Where("username = ?", user.Username).First(&database.User{}).RowsAffected != 0 {
				term.WriteTermColor("Duplicate user name\n", "red")
				continue
			}
			break
		}
	}
	setPassword, _ := term.InteractionSelect("Set a password?", []string{"yes", "no", "exit"}, "yes")
	if setPassword == "yes" {
		for {
			user.Password, _ = term.Interaction("password")
			if user.Password != "" {
				break
			}
		}
	} else if setPassword == "exit" {
		return
	}
	setKey, _ := term.InteractionSelect("Set Public Key Certificate? (support multiple)", []string{"yes", "no", "exit"}, "no")
	if setKey == "yes" {
		for {
			c, _ := term.Interaction("Public Key Certificate (Enter 'exit' exit)\n")
			if c == "exit" {
				break
			}
			_, err = ssh.ParsePublicKey([]byte(c))
			if err != nil {
				term.WriteTermColor("Public Key Certificate Error\n", "red")
				continue
			}
			user.PublicKeys.Data = append(user.PublicKeys.Data, c)
		}
	} else if setKey == "exit" {
		return
	}
	err = conf.Conf.DB.Create(user).Error
	if err != nil {
		term.WriteTermColor("DB Error: ", "red")
		return
	}
	term.WriteTermColor("Added successfully\n", "green")
}

func userList() string {
	var tmp strings.Builder
	table := tablewriter.NewWriter(&tmp)
	table.SetHeader([]string{"ID", "Name", "Number of assets owned", "Last Logon Time"})
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
	var users []*database.User
	err := conf.Conf.DB.Preload("Servers").Find(&users).Error
	if err != nil {
		return "DB Error: " + err.Error()
	}
	for i, v := range users {
		var lastLogin string
		if v.LastLogin != nil {
			lastLogin = v.LastLogin.Format("2006-01-02 15:04:05")
		}
		table.Append([]string{strconv.Itoa(i + 1), v.Username, strconv.Itoa(len(v.Servers)), lastLogin})
	}
	table.Render()
	return tmp.String()
}
