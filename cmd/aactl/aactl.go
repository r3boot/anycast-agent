package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/r3boot/anycast-agent/lib"
	"github.com/r3boot/anycast-agent/lib/consul"
	"github.com/r3boot/anycast-agent/lib/structs"
)

const (
	_d_default_consulEndpoint string = "http://localhost:8500"
)

func main() {
	var (
		consulEndpoint *string
		apply          *string
		get            *string
		delete         *string
		Consul         *consul.Consul
		err            error
	)

	// Global options
	consulEndpoint = flag.String(
		"consul",
		_d_default_consulEndpoint,
		"Connect to consul on this url",
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

	if *consulEndpoint == _d_default_consulEndpoint {
		for _, kv := range os.Environ() {
			pair := strings.Split(kv, "=")
			if pair[0] == "CONSUL_HTTP_ADDR" {
				*consulEndpoint = pair[1]
			}
		}
	}

	Consul, err = consul.NewConsul(*consulEndpoint)
	if err != nil {
		fmt.Println("newclient: " + err.Error())
		os.Exit(1)
	}

	fmt.Printf("Consul.Prefix: %v\n", Consul.Prefix)

	if *apply != "" {
		object, err := structs.LoadFromYaml(*apply)
		if err != nil {
			fmt.Print("apply: " + err.Error())
			os.Exit(1)
		}
		if err = Consul.ApplyObject(object); err != nil {
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

		prefix := Consul.Prefix
		switch objType {
		case structs.TypeBgpPeer:
			{
				prefix = prefix + "/peers"
			}
		case structs.TypeAnycast:
			{
				prefix = prefix + "/anycast"
			}
		}

		if name != "" {
			object, err := Consul.GetObject(objType, name)
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
			all_objects, err := Consul.GetAllObjects(objType, prefix)
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

	/*
		if *delete != "" {
			objType := *delete
			name := ""
			if strings.Contains(*delete, ":") {
				tokens := strings.Split(*delete, ":")
				objType = tokens[0]
				name = tokens[1]
			}

			prefix := Consul.Prefix
			switch objType {
			case structs.TypeBgpPeer:
				{
					prefix = prefix + "/peers"
				}
			case structs.TypeAnycast:
				{
					prefix = prefix + "/anycast"
				}
			}

			var items []string

			if name != "" {
				items = append(items, prefix+"/"+name)
			} else {
				all_items, err := Consul.Ls(prefix)
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
	*/
}
