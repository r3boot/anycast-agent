package consul

import "fmt"

func NewConsul(uri string) (*Consul, error) {
	c := &Consul{
		Prefix: "/services/anycast-agent",
		uri:    uri,
	}

	err := c.Connect()
	if err != nil {
		return nil, fmt.Errorf("NewConsul: %v", err)
	}

	return c, nil
}
