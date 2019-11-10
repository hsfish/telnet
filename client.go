package telnet

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	c    *Conn
	conf *ClientConf
}

func NewClient(c *ClientConf) (*Client, error) {
	initConf(c)

	conn, err := DialTimeout("tcp", fmt.Sprintf("%v:%v", c.Host, c.Port), time.Duration(c.DialTimeout)*time.Second)
	if err != nil {
		return nil, err
	}
	p := &Client{
		c:    conn,
		conf: c,
	}

	if err := p.login(); err != nil {
		p.Close()
		return nil, err
	}

	return p, nil
}

func (p *Client) login() error {

	p.c.SetReadDeadline(time.Now().Add(time.Duration(p.conf.Timeout) * time.Second))
	enterEmpty := false
	enterUser := false
	enterPass := false
	for {
		bytes := [10240]byte{}
		n, err := p.c.Read(bytes[0:])
		if err != nil {
			return err
		}

		if p.conf.User == "" && p.conf.Password == "" {
			return nil
		}

		line := string(bytes[:n])
		if enterPass {
			if p.conf.User != "" && regexp.MustCompile(p.conf.UserRegex).MatchString(line) {
				return LOGIN_ERR
			}
			if regexp.MustCompile(p.conf.PasswordRegex).MatchString(line) {
				return LOGIN_ERR
			}
			if regexp.MustCompile(p.conf.Pattern).MatchString(strings.TrimSpace(line)) {
				return nil
			}
		}

		if i := strings.LastIndexByte(line, '\n'); i > 0 {
			line = string(bytes[i:n])
		}
		if len(line) == 0 {
			continue
		}

		if p.conf.User != "" {
			if p.match(p.conf.UserRegex, line) {
				fmt.Println("user: ---", p.conf.UserRegex, p.conf.User)
				if err := p.SendStr(p.conf.User); err != nil {
					return err
				}
				enterUser = true
				continue
			}
		}

		if p.conf.Password != "" {
			if p.match(p.conf.PasswordRegex, line) {
				fmt.Println("pass:---", p.conf.PasswordRegex, line, p.conf.Password)
				if err := p.SendStr(p.conf.Password); err != nil {
					return err
				}
				enterPass = true
				continue
			}
		}

		if !enterEmpty && !(enterUser || enterPass) {
			// fmt.Println("send space")
			// if err := p.SendStr(""); err != nil {
			// 	return err
			// }
			enterEmpty = true
		}

	}
}

func (p *Client) SendStr(s string) error {

	p.c.SetWriteDeadline(time.Now().Add(time.Duration(p.conf.Timeout) * time.Second))
	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = '\n'
	_, err := p.c.Write(buf)
	return err
}

func (p *Client) match(pattern, line string) bool {
	regxp, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("正则表达式异常: ", err.Error())
		return false
	}
	return regxp.MatchString(line)
}

func (p *Client) Close() error {
	return p.c.Close()
}

func (p *Client) NewShell(opt *Options) (*Shell, error) {
	return NewShell(p.c, opt)
}

func initConf(c *ClientConf) {
	if c == nil {
		c = &ClientConf{
			DialTimeout:   DIAL_TIMEOUT_DEFAULT,
			Timeout:       TIMEOUT_DEFAULT,
			UserRegex:     USER_REGEX_DEFAULT,
			PasswordRegex: USER_REGEX_DEFAULT,
			Pattern:       PASSWORD_REGEX_DEFAULT,
		}
		return
	}

	if c.DialTimeout == 0 {
		c.DialTimeout = DIAL_TIMEOUT_DEFAULT
	}

	if c.Timeout == 0 {
		c.Timeout = TIMEOUT_DEFAULT
	}

	if c.UserRegex == "" {
		c.UserRegex = USER_REGEX_DEFAULT
	}

	if c.PasswordRegex == "" {
		c.PasswordRegex = PASSWORD_REGEX_DEFAULT
	}
	if c.Pattern == "" {
		c.Pattern = REGEX_DEFAULT
	}
}

var (
	LOGIN_ERR = errors.New("username or password error")
)
