package server

import (
	"bytes"
	"errors"
	"github.com/Rehtt/Kit/i18n"
	"github.com/Rehtt/Kit/util"
	"github.com/Rehtt/jumpjumpGo/conf"
	"github.com/Rehtt/jumpjumpGo/database"
	util2 "github.com/Rehtt/jumpjumpGo/util"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func authKeyboard(conn ssh.ConnMetadata, client ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
	var db = conf.Conf.DB
	username := conn.User()

	password, err := client("", "",
		[]string{"password:"},
		[]bool{false})
	if err != nil {
		return nil, err
	}

	var user database.User
	err = db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(i18n.GetText("not user"))
	}
	if err != nil {
		return nil, err
	}
	if !util2.CheckBcrypt(user.Password, password[0]) {
		return nil, errors.New(i18n.GetText("password error"))
	}
	user.LastLogin = util.TimeToPrt(time.Now())

	db.Where("id = ?", user.ID).Updates(&user)
	return &ssh.Permissions{
		CriticalOptions: map[string]string{"user": username, "id": strconv.Itoa(int(user.ID))},
		Extensions:      nil,
	}, nil
}

func authPrivateKeyfunc(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	var db = conf.Conf.DB
	var user database.User
	err := db.Where("username = ?", conn.User()).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(i18n.GetText("not user"))
	}
	if err != nil {
		return nil, err
	}
	for _, v := range user.PublicKeys.Data {
		pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(v))
		if err != nil {
			continue
		}
		if bytes.Equal(pub.Marshal(), key.Marshal()) {
			user.LastLogin = util.TimeToPrt(time.Now())
			db.Where("id = ?", user.ID).Updates(&user)
			return &ssh.Permissions{
				CriticalOptions: map[string]string{"user": conn.User(), "id": strconv.Itoa(int(user.ID))},
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(key),
				},
			}, nil
		}
	}
	return nil, errors.New(i18n.GetText("not user"))
}
