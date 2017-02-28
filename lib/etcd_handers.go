package lib

import (
	"errors"
	_ "fmt"
	"reflect"
	"strconv"
	"strings"
)

func (ec *EtcdClient) ApplyObject(object interface{}) error {
	var (
		err error
	)

	switch reflect.TypeOf(object).Name() {
	case "BgpPeerObject":
		{
			bgpPeer := object.(BgpPeerObject)
			path := ec.Prefix + "/peers/" + bgpPeer.Meta.Name

			if err = ec.MkdirIfNotExists(path); err != nil {
				return err
			}

			if err = ec.Set(path+"/asnum", strconv.Itoa(bgpPeer.Spec.AsNumber)); err != nil {
				return err
			}

			if bgpPeer.Spec.IP != "" {
				if err = ec.Set(path+"/ip", bgpPeer.Spec.IP); err != nil {
					return err
				}
			}

			if bgpPeer.Spec.IP6 != "" {
				if err = ec.Set(path+"/ip6", bgpPeer.Spec.IP6); err != nil {
					return err
				}
			}
		}
	case "AnycastObject":
		{
			anycast := object.(AnycastObject)
			path := ec.Prefix + "/anycast/" + anycast.Meta.Name

			if err = ec.MkdirIfNotExists(path); err != nil {
				return err
			}

			if err = ec.Set(path+"/asnum", strconv.Itoa(anycast.Spec.AsNumber)); err != nil {
				return err
			}

			if err = ec.Set(path+"/ip", anycast.Spec.IP); err != nil {
				return err
			}

			if err = ec.Set(path+"/ip6", anycast.Spec.IP6); err != nil {
				return err
			}

			if err = ec.Set(path+"/healthcheck", anycast.Spec.HealthCheck); err != nil {
				return err
			}

			peers := strings.Join(anycast.Spec.Peers, ",")
			if err = ec.Set(path+"/peers", peers); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ec *EtcdClient) GetObject(objType, name string) (interface{}, error) {
	var (
		err error
	)

	switch objType {
	case TypeBgpPeer:
		{
			path := ec.Prefix + "/peers/" + name

			if !ec.IsDir(path) {
				err = errors.New("GetObject: Object does not exist")
				return nil, err
			}

			object := BgpPeerObject{
				ApiVersion: 1,
				Type:       TypeBgpPeer,
				Meta: bgpPeerMetaObject{
					Name: name,
				},
				Spec: BgpPeerSpecObject{},
			}

			response, err := ec.Get(path + "/asnum")
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

			if response, err = ec.Get(path + "/ip"); err == nil {
				object.Spec.IP = response
			}

			if response, err = ec.Get(path + "/ip6"); err == nil {
				object.Spec.IP6 = response
			}

			return object, nil
		}
	case TypeAnycast:
		{
			path := ec.Prefix + "/anycast/" + name

			if !ec.IsDir(path) {
				err = errors.New("GetObject: Object does not exist")
				return nil, err
			}

			object := AnycastObject{
				ApiVersion: 1,
				Type:       TypeAnycast,
				Meta: anycastMetaObject{
					Name: name,
				},
				Spec: AnycastSpecObject{},
			}

			if response, err := ec.Get(path + "/asnum"); err == nil {
				if object.Spec.AsNumber, err = strconv.Atoi(response); err != nil {
					return nil, err
				}
			}

			if response, err := ec.Get(path + "/ip"); err == nil {
				object.Spec.IP = response
			}

			if response, err := ec.Get(path + "/ip6"); err == nil {
				object.Spec.IP6 = response
			}

			if response, err := ec.Get(path + "/healthcheck"); err == nil {
				object.Spec.HealthCheck = response
			}

			if response, err := ec.Get(path + "/peers"); err == nil {
				object.Spec.Peers = strings.Split(response, ",")
			}

			return object, nil
		}
	}

	err = errors.New("GetObject: Unknown error")
	return nil, err
}

func (ec *EtcdClient) GetAllObjects(objType, path string) ([]interface{}, error) {
	var (
		items    []interface{}
		name     string
		diritems []string
		err      error
	)

	if diritems, err = ec.Ls(path); err != nil {
		err = errors.New("GetAllObjects: ec.Ls() failed: " + err.Error())
		return nil, err
	}

	for _, name = range diritems {
		object, err := ec.GetObject(objType, name)
		if err != nil {
			err = errors.New("GetAllObjects: ec.GetObject() failed: " + err.Error())
			return nil, err
		}

		items = append(items, object)
	}

	return items, nil
}
