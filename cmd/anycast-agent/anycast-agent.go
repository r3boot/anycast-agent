package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/r3boot/anycast-agent/lib"
	"github.com/r3boot/anycast-agent/lib/agent"
)

const (
	_d_consulEndpoint string = "http://localhost:8500"
)

func main() {
	var (
		consulEndpoint *string
		name           *string
		err            error
	)

	consulEndpoint = flag.String(
		"consul",
		_d_consulEndpoint,
		"Connect to consul on this url",
	)

	name = flag.String(
		"name",
		"",
		"Anycast profile to use",
	)

	flag.Parse()

	if *name == "" {
		fmt.Println("Nothing to do")
		os.Exit(1)
	}

	if *consulEndpoint == _d_consulEndpoint {
		for _, kv := range os.Environ() {
			pair := strings.Split(kv, "=")
			if pair[0] == "CONSUL_HTTP_ADDR" {
				*consulEndpoint = pair[1]
			}
		}
	}

	anycastAgent, err := agent.NewAnycastAgent(*consulEndpoint, *name)
	if err != nil {
		fmt.Println("newclient: " + err.Error())
		os.Exit(1)
	}

	anycastAgent.RunAnycastService()

}
