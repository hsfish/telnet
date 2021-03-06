package example

import (
	"fmt"
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
	}

}
