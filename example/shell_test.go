package example

import (
	"auth/global"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hsfish/telnet"
)

func Test_Shell(t *testing.T) {

	c, err := newTelnet()
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	s, err := c.NewShell(&telnet.Options{
		Timeout:   telnet.TIMEOUT_DEFAULT,
		TrimFirst: true,
		Regex:     telnet.REGEX_DEFAULT,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		t.Error(err)
		return
	}
	defer s.End()

	for _, command := range []telnet.Command{
		telnet.Command{
			Cmd: "terminal length 0",
		},
		telnet.Command{
			Cmd:     "show startup-config",
			Pattern: "CiscoN3K#",
		},
	} {
		data := ""
		for {
			back := s.Run(command)
			data += back.Msg
			if back.Code != telnet.COMMAND_MORE {
				fmt.Println(back.Code, data)
				break
			}
		}
		fmt.Println("---------")
		fmt.Println(formatResp(data))
	}

}

func formatResp(lines string) string {

	r, err := regexp.Compile("\\[\\d+D")
	if err != nil {
		global.Logger.Error(err.Error())
		return lines
	}
	if !r.MatchString(lines) {
		return lines
	}

	flines := []string{}
	for _, line := range strings.Split(lines, "\n") {
		// 去掉结尾空格
		line = strings.TrimRight(line, " ")
		if !r.MatchString(line) {
			flines = append(flines, line)
			continue
		}

		matchs := r.FindAllStringIndex(line, -1)
		length := len(matchs)
		if length == 0 {
			continue
		}
		flines = append(flines, line[matchs[length-1][1]:])
	}

	return strings.Join(flines, "\n")

}
