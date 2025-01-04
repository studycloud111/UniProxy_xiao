package proxy

import (
	"testing"

	"github.com/studycloud111/UniProxy_xiao/v2b"
)

func TestStartProxy(t *testing.T) {
	var url []string
	url = append(url, "http://127.0.0.1:1022")
	v2b.Init("", url, "xxxxxxxx")
	s, _ := v2b.GetServers()
	t.Log(s[0])
	InPort = 1151
	GlobalMode = true
	t.Log(StartProxy("test", "xxxxxx", &s[0]))
	select {}
}
