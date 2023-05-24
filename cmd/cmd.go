package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/Rehtt/Kit/i18n"
	"github.com/Rehtt/jumpjumpGo/conf"
	"github.com/Rehtt/jumpjumpGo/database"
	"github.com/Rehtt/jumpjumpGo/util"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
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

	term.WriteTerm(i18n.GetText("user manage\n"))
	term.WriteTerm(i18n.GetText("commands:\nlist\tadd\tdel\tchange\n"))
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
			delUser(term)
		case "change":
			changeUser(term)
		case "exit":
			ch(errors.New("jumpjumpGo shutdown"))
			return
		}
	}
}

func changeUser(term *Term) {
	var user = new(database.User)
	var err error
	defer func() {
		if err != nil {
			term.WriteTermColor(err.Error()+"\n", "red")
		}
	}()
	id, err := term.Interaction(i18n.GetText("User ID"))
	if err != nil {
		return
	}
	err = conf.Conf.DB.Where("id  = ?", id).First(user).Error
	if err != nil {
		return
	}

	for {
		user.Username, err = term.InteractionDefault(fmt.Sprintf(i18n.GetText("User Name (Original: %s)"), user.Username), user.Username)
		if err != nil {
			return
		}
		if user.Username == "" {
			term.WriteTermColor(i18n.GetText("Please enter the correct user name\n"), "red")
		} else {
			if conf.Conf.DB.Where("username = ? AND id != ?", user.Username, id).First(&database.User{}).RowsAffected != 0 {
				term.WriteTermColor("Duplicate user name\n", "red")
				continue
			}
			break
		}
	}
	setPassword, err := term.InteractionSelect(i18n.GetText("Set a password?"), []string{"yes", "no", "exit"}, "yes")
	if err != nil {
		return
	}
	if setPassword == "yes" {
		for {
			user.Password, _ = term.Interaction(i18n.GetText("password"))
			if user.Password != "" {
				user.Password = util.Bcrypt(user.Password)
				break
			}
		}
	} else if setPassword == "exit" {
		return
	}
	setKey, err := term.InteractionSelect(i18n.GetText("Set Public Key Certificate? (support multiple)"), []string{"yes", "no", "exit"}, "no")
	if err != nil {
		return
	}
	if setKey == "yes" {
		var c string
		for {
			c, err = term.Interaction(i18n.GetText("Public Key Certificate (Enter 'exit' exit)\n"))
			if err != nil {
				return
			}
			if c == "exit" {
				break
			}
			_, _, _, _, err = ssh.ParseAuthorizedKey([]byte(c))
			if err != nil {
				term.WriteTermColor(i18n.GetText("Public Key Certificate Error\n"), "red")
				continue
			}
			user.PublicKeys.Data = append(user.PublicKeys.Data, c)
		}
	} else if setKey == "exit" {
		return
	}
	err = conf.Conf.DB.Where("id = ?", id).Updates(user).Error
	if err != nil {
		term.WriteTermColor(i18n.GetText("DB Error: ")+err.Error(), "red")
		return
	}
	term.WriteTermColor(i18n.GetText("Added successfully\n"), "green")
}

func delUser(term *Term) {
	id, err := term.Interaction(i18n.GetText("User ID"))
	if err != nil {
		return
	}
	err = conf.Conf.DB.Where("id = ?", id).Delete(&database.User{}).Error
	if err != nil {
		term.WriteTermColor(i18n.GetText("DB Error: ")+err.Error(), "red")
	}
	err = conf.Conf.DB.Where("user_id = ?", id).Delete(&database.UserServer{}).Error
	if err != nil {
		term.WriteTermColor(i18n.GetText("DB Error: ")+err.Error(), "red")
	}
	term.WriteTermColor(i18n.GetText("Added successfully\n"), "green")
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
		user.Username, err = term.Interaction(i18n.GetText("User Name"))
		if err != nil {
			return
		}
		if user.Username == "" {
			term.WriteTermColor(i18n.GetText("Please enter the correct user name\n"), "red")
		} else {
			if conf.Conf.DB.Where("username = ?", user.Username).First(&database.User{}).RowsAffected != 0 {
				term.WriteTermColor("Duplicate user name\n", "red")
				continue
			}
			break
		}
	}
	setPassword, err := term.InteractionSelect(i18n.GetText("Set a password?"), []string{"yes", "no", "exit"}, "yes")
	if err != nil {
		return
	}
	if setPassword == "yes" {
		for {
			user.Password, err = term.Interaction(i18n.GetText("password"))
			if err != nil {
				return
			}
			if user.Password != "" {
				user.Password = util.Bcrypt(user.Password)
				break
			}
		}
	} else if setPassword == "exit" {
		return
	}
	setKey, err := term.InteractionSelect(i18n.GetText("Set Public Key Certificate? (support multiple)"), []string{"yes", "no", "exit"}, "no")
	if err != nil {
		return
	}
	if setKey == "yes" {
		var c string
		for {
			c, err = term.Interaction(i18n.GetText("Public Key Certificate (Enter 'exit' exit)\n"))
			if err != nil {
				return
			}
			if c == "exit" {
				break
			}
			_, _, _, _, err = ssh.ParseAuthorizedKey([]byte(c))
			if err != nil {
				term.WriteTermColor(i18n.GetText("Public Key Certificate Error\n"), "red")
				continue
			}
			user.PublicKeys.Data = append(user.PublicKeys.Data, c)
		}
	} else if setKey == "exit" {
		return
	}
	err = conf.Conf.DB.Create(user).Error
	if err != nil {
		term.WriteTermColor(i18n.GetText("DB Error: "), "red")
		return
	}
	term.WriteTermColor(i18n.GetText("Added successfully\n"), "green")
}

func userList() string {
	var tmp strings.Builder
	table := tablewriter.NewWriter(&tmp)
	table.SetHeader([]string{"ID", i18n.GetText("Name"), i18n.GetText("Number of assets owned"), i18n.GetText("Last Logon Time")})
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
	for _, v := range users {
		var lastLogin string
		if v.LastLogin != nil {
			lastLogin = v.LastLogin.Format("2006-01-02 15:04:05")
		}
		table.Append([]string{strconv.Itoa(int(v.ID)), v.Username, strconv.Itoa(len(v.Servers)), lastLogin})
	}
	table.Render()
	return tmp.String()
}
