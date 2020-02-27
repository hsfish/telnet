package example

import (
	"testing"

	"github.com/hsfish/telnet"
)

func newTelnet() (*telnet.Client, error) {
	conf := &telnet.ClientConf{
		Host:        "192.168.2.13",
		Port:        23,
		User:        "admin",
		Password:    "geesunn1231",
		Timeout:     10,
		DialTimeout: 5,
	}
	return telnet.NewClient(conf)
}

func Test_NewClient(t *testing.T) {

	c, err := newTelnet()
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()
	t.Log(c)
}
