package util

import (
	"errors"
	"golang.org/x/crypto/ssh"
)

// RFC 4254 Section 6.2.
type PtyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

func PtyReq(w, h uint32, modes ...ssh.TerminalModes) []byte {
	m := ssh.TerminalModes{
		ssh.TTY_OP_ISPEED: 14400, // Input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // Output speed = 14.4kbaud
	}
	if len(modes) != 0 {
		for k, v := range modes[0] {
			m[k] = v
		}
	}
	var tm []byte
	for k, v := range m {
		kv := struct {
			Key byte
			Val uint32
		}{k, v}

		tm = append(tm, ssh.Marshal(&kv)...)
	}
	tm = append(tm, 0)
	req := PtyRequestMsg{
		Term:     "xterm",
		Columns:  w,
		Rows:     h,
		Width:    w * 8,
		Height:   h * 8,
		Modelist: string(tm),
	}
	return ssh.Marshal(req)
}
func SendPtyReq(chann ssh.Channel, w, h uint32, modes ...ssh.TerminalModes) error {
	ok, err := chann.SendRequest("pty-req", true, PtyReq(w, h, modes...))
	if err == nil && !ok {
		return errors.New("ssh: pty-req failed")
	}
	return nil
}
