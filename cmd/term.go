package cmd

import (
	"errors"
	"fmt"
	"github.com/mgutz/ansi"
	"golang.org/x/term"
	"io"
	"jumpjumpGo/util"
	"strings"
)

type Term struct {
	*term.Terminal
}

func (t *Term) Interaction(query string, isPassword ...bool) (answer string, err error) {
	if len(isPassword) != 0 && isPassword[0] {
		return t.ReadPassword(query)
	}
	t.WriteTerm(query)
	answer, err = t.ReadLine()
	return
}
func (t *Term) InteractionDefault(query string, defaultValue string, isPassword ...bool) (answer string, err error) {
	answer, err = t.Interaction(query, isPassword...)
	if err != nil {
		return "", err
	}
	if answer == "" {
		answer = defaultValue
	}
	return
}
func (t *Term) InteractionSelect(query string, selectV []string, defaultValue string) (string, error) {
	if !util.InStringArray(defaultValue, selectV) {
		return "", errors.New("selectV not defaultValue")
	}
	query = fmt.Sprintf("%s [%s] (Default: %s)", query, strings.Join(selectV, "/"), defaultValue)
	for {
		a, err := t.InteractionDefault(query, defaultValue)
		if err != nil {
			return "", err
		}
		if util.InStringArray(a, selectV) {
			return a, nil
		}
	}
}

func (t *Term) WriteTerm(str string) {
	t.Write([]byte(str))
}
func (t *Term) WriteTermColor(str string, style string) {
	t.WriteTerm(ansi.Color(str, style))
}
func NewTerm(c io.ReadWriter, p string) *Term {
	return &Term{Terminal: term.NewTerminal(c, p)}
}
