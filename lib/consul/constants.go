package consul

import (
	"github.com/hashicorp/consul/api"
)

type Consul struct {
	uri    string
	Prefix string
	client *api.Client
	kv     *api.KV
}
