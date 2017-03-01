package main

import (
	"flag"
	"fmt"
	_ "github.com/r3boot/anycast-agent/lib"
	"github.com/r3boot/anycast-agent/lib/agent"
	"os"
	"strings"
)

const (
	_d_etcdEndpoint string = "http://localhost:2379"
)

func main() {
	var (
		etcdEndpoints *string
		name          *string
		anycastAgent  *agent.AnycastAgent
		endpoints     []string
		err           error
	)

	etcdEndpoints = flag.String(
		"etcd",
		_d_etcdEndpoint,
		"Connect to etcd on these comma-separated urls",
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

	if *etcdEndpoints == _d_etcdEndpoint {
		for _, kv := range os.Environ() {
			pair := strings.Split(kv, "=")
			if pair[0] == "ETCD_ENDPOINTS" {
				*etcdEndpoints = pair[1]
			}
		}
	}

	endpoints = strings.Split(*etcdEndpoints, ",")

	anycastAgent, err = agent.NewAnycastAgent(endpoints, *name)
	if err != nil {
		fmt.Println("newclient: " + err.Error())
		os.Exit(1)
	}

	anycastAgent.RunAnycastService()

}
