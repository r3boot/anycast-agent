package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/r3boot/anycast-agent/lib"
)

const (
	_d_default_etcdEndpoint string = "http://localhost:2379"
)

func main() {
	var (
		etcdEndpoints *string
		apply         *string
		get           *string
		delete        *string
		etcd          *lib.EtcdClient
		endpoints     []string
		err           error
	)

	// Global options
	etcdEndpoints = flag.String(
		"etcd",
		_d_default_etcdEndpoint,
		"Connect to etcd on these comma-separated urls",
	)

	apply = flag.String(
		"apply",
		"",
		"File containing object to apply",
	)

	get = flag.String(
		"get",
		"",
		"Get object(s) (type:name)",
	)

	delete = flag.String(
		"delete",
		"",
		"File containing object to delete",
	)

	flag.Parse()

	if *apply == "" && *delete == "" && *get == "" {
		fmt.Println("Nothing to do")
		os.Exit(1)
	}

	if *etcdEndpoints == _d_default_etcdEndpoint {
		for _, kv := range os.Environ() {
			pair := strings.Split(kv, "=")
			if pair[0] == "ETCD_ENDPOINT" {
				*etcdEndpoints = pair[1]
			}
		}
	}

	endpoints = strings.Split(*etcdEndpoints, ",")

	etcd, err = lib.NewEtcdClient(endpoints, "/am/v1")
	if err != nil {
		fmt.Println("newclient: " + err.Error())
		os.Exit(1)
	}

	if *apply != "" {
		object, err := lib.LoadFromYaml(*apply)
		if err != nil {
			fmt.Print("apply: " + err.Error())
			os.Exit(1)
		}
		if err = etcd.ApplyObject(object); err != nil {
			fmt.Print("etcd.ApplyObject: " + err.Error())
		}
	}

	if *get != "" {
		objType := *get
		name := ""
		if strings.Contains(*get, ":") {
			tokens := strings.Split(*get, ":")
			objType = tokens[0]
			name = tokens[1]
		}

		prefix := etcd.Prefix
		switch objType {
		case lib.TypeBgpPeer:
			{
				prefix = prefix + "/peers"
			}
		case lib.TypeAnycast:
			{
				prefix = prefix + "/anycast"
			}
		}

		if name != "" {
			object, err := etcd.GetObject(objType, name)
			if err != nil {
				fmt.Println("get: " + err.Error())
				os.Exit(1)
			}

			output, err := lib.DumpYaml(object)
			if err != nil {
				fmt.Println("get: " + err.Error())
				os.Exit(1)
			}

			fmt.Print(string(output))
		} else {
			all_objects, err := etcd.GetAllObjects(objType, prefix)
			if err != nil {
				fmt.Println("get: " + err.Error())
				os.Exit(1)
			}

			output, err := lib.DumpYaml(all_objects)
			if err != nil {
				fmt.Println("get: " + err.Error())
				os.Exit(1)
			}

			fmt.Print(string(output))
		}
	}

	if *delete != "" {
		objType := *delete
		name := ""
		if strings.Contains(*delete, ":") {
			tokens := strings.Split(*delete, ":")
			objType = tokens[0]
			name = tokens[1]
		}

		prefix := etcd.Prefix
		switch objType {
		case lib.TypeBgpPeer:
			{
				prefix = prefix + "/peers"
			}
		case lib.TypeAnycast:
			{
				prefix = prefix + "/anycast"
			}
		}

		var items []string

		if name != "" {
			items = append(items, prefix+"/"+name)
		} else {
			all_items, err := etcd.Ls(prefix)
			if err != nil {
				fmt.Println("delete: " + err.Error())
				os.Exit(1)
			}
			for _, item := range all_items {
				items = append(items, prefix+"/"+item)
			}
		}

		for _, item := range items {
			if err = etcd.Rm(item); err != nil {
				fmt.Println("delete: " + err.Error())
				os.Exit(1)
			}
		}

	}
}
