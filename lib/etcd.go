package lib

import (
	"errors"
	_ "fmt"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/client"
	context "golang.org/x/net/context"
)

type EtcdClient struct {
	c            etcd.Client
	kapi         etcd.KeysAPI
	initializing bool
	Prefix       string
}

func NewEtcdClient(endpoints []string, prefix string) (*EtcdClient, error) {
	var (
		ec     *EtcdClient
		client etcd.Client
		err    error
	)

	client, err = etcd.New(etcd.Config{
		Endpoints:               endpoints,
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})
	if err != nil {
		return nil, err
	}

	ec = &EtcdClient{
		c:      client,
		kapi:   etcd.NewKeysAPI(client),
		Prefix: prefix,
	}

	if err = ec.Initialize("/am/v1"); err != nil {
		return nil, err
	}

	return ec, nil
}

func (ec *EtcdClient) Initialize(prefix string) error {
	var (
		err      error
		item     string
		testpath string
	)

	ec.initializing = true

	testpath = ""
	for _, item = range strings.Split(prefix, "/")[1:] {
		testpath = testpath + "/" + item
		if err = ec.MkdirIfNotExists(testpath); err != nil {
			return err
		}
	}

	bgp_peers := ec.Prefix + "/peers"
	if err = ec.MkdirIfNotExists(bgp_peers); err != nil {
		return err
	}

	anycast_entries := ec.Prefix + "/anycast"
	if err = ec.MkdirIfNotExists(anycast_entries); err != nil {
		return err
	}

	ec.initializing = false

	return nil
}

func (ec *EtcdClient) GetWithResponse(key string) (*etcd.Response, error) {
	var (
		response *etcd.Response
		err      error
	)

	response, err = ec.kapi.Get(context.Background(), key, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (ec *EtcdClient) Get(key string) (string, error) {
	var (
		response *etcd.Response
		err      error
	)

	if response, err = ec.GetWithResponse(key); err != nil {
		return "", err
	}

	return response.Node.Value, nil
}

func (ec *EtcdClient) IsDir(key string) bool {
	var (
		response *etcd.Response
		err      error
	)

	if response, err = ec.GetWithResponse(key); err != nil {
		return false
	}

	return response.Node.Dir
}

func (ec *EtcdClient) Has(key string) bool {
	var (
		response string
		err      error
	)

	response, err = ec.Get(key)

	return (response != "") && (err == nil)

}

func (ec *EtcdClient) Set(key, value string) error {
	var (
		err error
	)

	_, err = ec.kapi.Set(context.Background(), key, value, nil)
	if err != nil {
		return err
	}

	return nil
}

func (ec *EtcdClient) MkDir(path string) error {
	var (
		err error
	)

	_, err = ec.kapi.Set(context.Background(), path, "", &etcd.SetOptions{
		Dir: true,
	})
	if err != nil {
		return err
	}

	return nil
}

func (ec *EtcdClient) MkdirIfNotExists(path string) error {
	var (
		err error
	)

	if !ec.IsDir(path) {
		if ec.Has(path) {
			if err = ec.Delete(path); err != nil {
				err = errors.New("ec.Delete: " + err.Error())
				return err
			}
		}
		if err = ec.MkDir(path); err != nil {
			err = errors.New("ec.MkDir: " + err.Error())
			return err
		}
	}

	return nil
}

func (ec *EtcdClient) Delete(key string) error {
	var (
		err error
	)

	_, err = ec.kapi.Delete(context.Background(), key, nil)
	if err != nil {
		return err
	}

	return nil
}

func (ec *EtcdClient) Ls(path string) ([]string, error) {
	var (
		response *etcd.Response
		items    []string
		err      error
	)

	if response, err = ec.GetWithResponse(path); err != nil {
		err = errors.New("Ls: ec.GetWithResponse() failed: " + err.Error())
		return nil, err
	}

	for _, item := range response.Node.Nodes {
		path_tokens := strings.Split(item.Key, "/")
		name := path_tokens[len(path_tokens)-1:][0]
		items = append(items, name)
	}

	return items, nil
}

func (ec *EtcdClient) Rm(path string) error {
	var (
		err error
	)

	_, err = ec.kapi.Delete(context.Background(), path, &etcd.DeleteOptions{
		Dir:       true,
		Recursive: true,
	})
	return err
}
