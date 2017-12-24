package consul

import (
	"errors"
	"fmt"
	"strconv"

	"reflect"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/r3boot/anycast-agent/lib/structs"
)

func (c *Consul) Connect() error {
	var err error

	c.client, err = api.NewClient(api.DefaultConfig())
	if err != nil {
		return fmt.Errorf("Consul.Connect api.NewClient: %v", err)
	}

	c.kv = c.client.KV()

	return nil
}

func (c *Consul) Set(key, value string) error {
	if strings.HasPrefix(key, "/") {
		key = key[1:]
	}

	fmt.Printf("Consul.Set: key=%s; value: %s\n", key, value)

	data := &api.KVPair{Key: key, Value: []byte(value)}
	_, err := c.kv.Put(data, nil)
	if err != nil {
		return fmt.Errorf("Consul.Set kv.Put: %v", err)
	}

	return nil
}

func (c *Consul) Get(key string) (string, error) {
	data, _, err := c.kv.Get(key, nil)
	if err != nil {
		return "", fmt.Errorf("Consul.Get: kv.Get: %v", err)
	}

	return string(data.Value), nil
}

func (c *Consul) Ls(path string) ([]string, error) {
	data, _, err := c.kv.List(path, &api.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("Consul.Ls: kv.List: %v", err)
	}

	allEntries := []string{}
	for _, entry := range data {
		tokens := strings.Split(string(entry.Value), "/")
		key := tokens[len(tokens)]
		allEntries = append(allEntries, key)
	}

	return allEntries, nil
}

func (c *Consul) ApplyObject(object interface{}) error {
	var (
		err error
	)

	switch reflect.TypeOf(object).Name() {
	case "BgpPeerObject":
		{
			bgpPeer := object.(structs.BgpPeerObject)
			path := c.Prefix + "/peers/" + bgpPeer.Meta.Name

			if err = c.Set(path+"/asnum", strconv.Itoa(bgpPeer.Spec.AsNumber)); err != nil {
				return err
			}

			if bgpPeer.Spec.IP != "" {
				if err = c.Set(path+"/ip", bgpPeer.Spec.IP); err != nil {
					return err
				}
			}

			if bgpPeer.Spec.IP6 != "" {
				if err = c.Set(path+"/ip6", bgpPeer.Spec.IP6); err != nil {
					return err
				}
			}
		}
	case "AnycastObject":
		{
			anycast := object.(structs.AnycastObject)
			path := c.Prefix + "/services/" + anycast.Meta.Name

			if err = c.Set(path+"/asnum", strconv.Itoa(anycast.Spec.AsNumber)); err != nil {
				return err
			}

			if err = c.Set(path+"/ip", anycast.Spec.IP); err != nil {
				return err
			}

			if err = c.Set(path+"/ip6", anycast.Spec.IP6); err != nil {
				return err
			}

			if err = c.Set(path+"/healthcheck", anycast.Spec.HealthCheck); err != nil {
				return err
			}

			peers := strings.Join(anycast.Spec.Peers, ",")
			if err = c.Set(path+"/peers", peers); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Consul) GetObject(objType, name string) (interface{}, error) {
	var (
		err error
	)

	switch objType {
	case structs.TypeBgpPeer:
		{
			path := c.Prefix + "/peers/" + name

			object := structs.BgpPeerObject{
				ApiVersion: 1,
				Type:       structs.TypeBgpPeer,
				Meta: structs.BgpPeerMetaObject{
					Name: name,
				},
				Spec: structs.BgpPeerSpecObject{},
			}

			response, err := c.Get(path + "/asnum")
			if err != nil {
				err = errors.New("GetObject: asnum not found")
				return nil, err
			}
			asnum, err := strconv.Atoi(response)
			if err != nil {
				err = errors.New("GetObject: Failed to convert asnum to int")
				return nil, err
			}
			object.Spec.AsNumber = asnum

			if response, err = c.Get(path + "/ip"); err == nil {
				object.Spec.IP = response
			}

			if response, err = c.Get(path + "/ip6"); err == nil {
				object.Spec.IP6 = response
			}

			return object, nil
		}
	case structs.TypeAnycast:
		{
			path := c.Prefix + "/services/" + name

			object := structs.AnycastObject{
				ApiVersion: 1,
				Type:       structs.TypeAnycast,
				Meta: structs.AnycastMetaObject{
					Name: name,
				},
				Spec: structs.AnycastSpecObject{},
			}

			if response, err := c.Get(path + "/asnum"); err == nil {
				if object.Spec.AsNumber, err = strconv.Atoi(response); err != nil {
					return nil, err
				}
			}

			if response, err := c.Get(path + "/ip"); err == nil {
				object.Spec.IP = response
			}

			if response, err := c.Get(path + "/ip6"); err == nil {
				object.Spec.IP6 = response
			}

			if response, err := c.Get(path + "/healthcheck"); err == nil {
				object.Spec.HealthCheck = response
			}

			if response, err := c.Get(path + "/peers"); err == nil {
				object.Spec.Peers = strings.Split(response, ",")
			}

			return object, nil
		}
	}

	err = errors.New("GetObject: Unknown error")
	return nil, err
}

func (c *Consul) GetAllObjects(objType, path string) ([]interface{}, error) {
	var (
		items    []interface{}
		name     string
		diritems []string
		err      error
	)

	if diritems, err = c.Ls(path); err != nil {
		err = errors.New("GetAllObjects: c.Ls() failed: " + err.Error())
		return nil, err
	}

	for _, name = range diritems {
		object, err := c.GetObject(objType, name)
		if err != nil {
			err = errors.New("GetAllObjects: c.GetObject() failed: " + err.Error())
			return nil, err
		}

		items = append(items, object)
	}

	return items, err
}
