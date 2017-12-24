package consul

import (
	"fmt"
	"strings"
)

func NewConsul(uri string) (*Consul, error) {
	if strings.HasPrefix(uri, "http://") {
		uri = uri[7:]
	} else if strings.HasPrefix(uri, "https://") {
		uri = uri[8:]
	}

	c := &Consul{
		Prefix: "services",
		uri:    uri,
	}

	err := c.Connect()
	if err != nil {
		return nil, fmt.Errorf("NewConsul: %v", err)
	}

	return c, nil
}
