package telnet

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Shell struct {
	c     *Conn
	opt   *Options
	isRun bool
	regxp *regexp.Regexp
	down  context.CancelFunc
	in    chan<- Command
	out   <-chan CommandBack
}

func NewShell(c *Conn, opt *Options) (*Shell, error) {
	initOpt(opt)
	var regxp *regexp.Regexp
	if opt.Regex != "" {
		var err error
		regxp, err = regexp.Compile(opt.Regex)
		if err != nil {
			return nil, err
		}
	}
	return &Shell{c: c, opt: opt, regxp: regxp}, nil
}

func (p *Shell) Start() error {

	if p.isRun {
		return errors.New("shell is run")
	}
	p.isRun = true

	ctx, down := context.WithCancel(context.Background())
	p.down = down
	p.in, p.out = p.shell(ctx)
	return nil
}

// 终止执行
func (p *Shell) End() error {
	p.down()
	return nil
}

func (p *Shell) Run(cmd Command) CommandBack {
	p.in <- cmd
	return <-p.out
}

func (p *Shell) shell(ctx context.Context) (chan<- Command, <-chan CommandBack) {
	in := make(chan Command, 1)
	out := make(chan CommandBack, 1)
	option := make(chan string, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				out <- commandFailed(fmt.Sprintf("panic: %v", r))
				p.End()
			}
			close(option)
			fmt.Println("down")
		}()

		for {
			select {
			case cmd, ok := <-in:
				if !ok {
					return
				}
				if _, err := p.c.Write([]byte(cmd.Cmd + "\n")); err != nil {
					out <- commandFailed(err.Error())
				}
				if len(option) > 0 {
					panic("shell error")
				}
				option <- cmd.Pattern

			case <-ctx.Done(): // 收到退出信号
				return
			}
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				out <- commandFailed(fmt.Sprintf("panic: %v", r))
				p.End()
			}
			close(in)
			close(out)
			fmt.Println("down")
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case pattern, ok := <-option:
				if !ok {
					return
				}
				p.resp(out, pattern)
			}
		}
	}()
	return in, out
}

func (p *Shell) resp(out chan CommandBack, pattern string) {

	regxp := p.regxp
	if pattern != "" {
		tRegxp, err := regexp.Compile(pattern)
		if err != nil {
			out <- commandFailed(err.Error())
			return
		}
		regxp = tRegxp
	}

	buf := []byte{}
	for {

		bytes, err := p.read()
		if err != nil {
			out <- commandFailed(string(buf) + "\n" + err.Error())
			return
		}

		if regxp == nil {
			out <- commandMore(string(bytes))
			continue
		}

		isReturn := false
		buf = append(buf, bytes...)
		if i := strings.LastIndexByte(string(buf), '\n'); i > 0 {
			lastLine := strings.TrimSpace(string(buf[i+1:]))
			if regxp.MatchString(lastLine) {
				fmt.Println(p.opt.Regex, "--1--", lastLine)
				buf = buf[:i+1]
				isReturn = true
			} else if p.handlerMore(lastLine) {
				buf = buf[:i+1]
			}
		} else {
			lastLine := strings.TrimSpace(string(buf))
			if regxp.MatchString(lastLine) {
				buf = []byte{}
				isReturn = true
			} else if p.handlerMore(lastLine) {
				buf = []byte{}
			}
		}

		if !isReturn {
			continue
		}

		if p.opt.TrimFirst {
			if len(buf) > 0 {
				if i := strings.IndexByte(string(buf), '\n'); i > 0 {
					buf = buf[i+1:]
				}
			}
		}

		out <- commandSUCCESS(string(buf))
		return
	}

}

func (p *Shell) read() ([]byte, error) {
	// 读取超时返回
	p.c.SetReadDeadline(time.Now().Add(time.Duration(p.opt.Timeout) * time.Second))

	bytes := [4096]byte{}
	n, err := p.c.Read(bytes[0:])
	if err != nil {
		return nil, err
	}
	return bytes[:n], nil
}

func (p *Shell) handlerMore(line string) bool {

	for _, str := range []string{"--More--", "---- More ----"} {
		if strings.Contains(line, str) {
			p.c.Write([]byte{' '})
			return true
		}
	}

	return false
}

func commandFailed(msg string) CommandBack {
	return CommandBack{Code: COMMAND_FAILED, Msg: msg}
}

func commandSUCCESS(msg string) CommandBack {
	return CommandBack{Code: COMMAND_SUCCESS, Msg: msg}
}

func commandMore(msg string) CommandBack {
	return CommandBack{Code: COMMAND_MORE, Msg: msg}
}

// 关闭
func (p *Shell) Close() error {

	return nil
}

func initOpt(opt *Options) {

	if opt == nil {
		opt = &Options{
			Timeout:   TIMEOUT_DEFAULT,
			TrimFirst: true,
			Regex:     REGEX_DEFAULT,
		}
		return
	}

	if opt.Timeout == 0 {
		opt.Timeout = TIMEOUT_DEFAULT
	}

}
