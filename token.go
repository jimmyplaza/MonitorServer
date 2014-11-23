package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"time"
)

type Token struct {
	User   string
	Action string
	Value  string
	Time   string
	DF     string
	token  string
}

func (t *Token) String() string {
	return t.token
}

func (t *Token) Make() {
	h := sha1.New()
	t.Time = time.Now().Format("2006-01-02-15")
	str := fmt.Sprintf("user=%s&action=%s&value=%s&time=%s&DF=%s", t.User, t.Action, t.Value, t.Time, t.DF)
	fmt.Println("******str: %s",str)
	io.WriteString(h, str)
	data := h.Sum(nil)
	t.token = fmt.Sprintf("%x", data)
}

func (t *Token) ChangeTime() {
	h := sha1.New()
	newTime := time.Now().Add(-60 * 60 * 1 * time.Second)
	t.Time = newTime.Format("2006-01-02-15")
	str := fmt.Sprintf("user=%s&action=%s&value=%s&time=%s&DF=%s", t.User, t.Action, t.Value, t.Time, t.DF)
	io.WriteString(h, str)
	data := h.Sum(nil)
	t.token = fmt.Sprintf("%x", data)
}

func (t *Token) Test() {
	h := sha1.New()
	newTime := time.Now().Add(-60 * 60 * 2 * time.Second)
	t.Time = newTime.Format("2006-01-02-15")
	str := fmt.Sprintf("user=%s&action=%s&value=%s&time=%s&DF=%s", t.User, t.Action, t.Value, t.Time, t.DF)
	io.WriteString(h, str)
	data := h.Sum(nil)
	t.token = fmt.Sprintf("%x", data)
}

func (t *Token) Verify(tok *Token) bool {
	if t.token == tok.token && len(tok.token) > 0 {
		return true
	}
	t.ChangeTime()
	if t.token == tok.token {
		return true
	}
	return false
}

func (t *Token) VerifyString(tok string) bool {
	if t.token == tok && len(tok) > 0 {
		return true
	}
	t.ChangeTime()
	if t.token == tok {
		return true
	}
	return false
}

func TestToken() {
	token := &Token{User: "matis.hsiao@nexusguard.com", Action: "reset", Value: "1000", DF: "myself"}
	token.Make()
	fmt.Println(token.Time)
	fmt.Println(token.String())

	token2 := &Token{User: "matis.hsiao@nexusguard.com", Action: "reset", Value: "1000", DF: "myself"}
	token2.Make()
	fmt.Println(token2.Time)
	fmt.Println(token2.String())
	fmt.Println(token.Verify(token2))
	token2.ChangeTime()
	fmt.Println(token2.Time)
	fmt.Println(token2.String())
	fmt.Println(token.Verify(token2))
	token2.Test()
	fmt.Println(token2.Time)
	fmt.Println(token2.String())
	fmt.Println(token.Verify(token2))
}
