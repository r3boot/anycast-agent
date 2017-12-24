package structs

import (
	"errors"
	"io/ioutil"
	"net"
	"strconv"

	"gopkg.in/yaml.v2"
)

func ValidateBgpPeerYaml(bgpPeer BgpPeerObject) error {
	var (
		err error
	)
	if bgpPeer.Meta.Name == "" {
		err = errors.New("ValidateBgpPeerYaml: bgpPeer.Meta.Name not set")
		return err
	}

	if bgpPeer.Spec.AsNumber == 0 {
		err = errors.New("ValidateBgpPeerYaml: bgpPeer.Spec.AsNumber not set")
		return err
	}

	if bgpPeer.Spec.IP == "" && bgpPeer.Spec.IP6 == "" {
		err = errors.New("ValidateBgpPeerYaml: neither bgpPeer.Spec.IP or bgpPeer.Spec.IP6 set")
		return err
	}

	if bgpPeer.Spec.IP != "" {
		if ip := net.ParseIP(bgpPeer.Spec.IP); ip == nil {
			err = errors.New("ValidateBgpPeerYaml: bgpPeer.Spec.IP: Not an ip address: " + bgpPeer.Spec.IP)
			return err
		}
	}

	if bgpPeer.Spec.IP6 != "" {
		if ip := net.ParseIP(bgpPeer.Spec.IP6); ip == nil {
			err = errors.New("ValidateBgpPeerYaml: bgpPeer.Spec.IP6: Not an ip address: " + bgpPeer.Spec.IP)
			return err
		}
	}

	return nil
}

func LoadFromYaml(fname string) (interface{}, error) {
	var (
		te   objectTypeExtractor
		data []byte
		err  error
	)

	if data, err = ioutil.ReadFile(fname); err != nil {
		err = errors.New("LoadFromYaml: " + err.Error())
		return nil, err
	}

	if err = yaml.Unmarshal(data, &te); err != nil {
		err = errors.New("LoadFromYaml: yaml.Unmarshal() failed: " + err.Error())
		return nil, err
	}

	if te.ApiVersion != 1 {
		err = errors.New("LoadFromYaml: unknown apiVersion: " + strconv.Itoa(te.ApiVersion))
		return nil, err
	}

	switch te.Type {
	case TypeBgpPeer:
		{
			var bgpPeer BgpPeerObject

			if err = yaml.Unmarshal(data, &bgpPeer); err != nil {
				err = errors.New("LoadFromYaml: Failed to unmarshal to bgpPeer: " + err.Error())
				return nil, err
			}

			if err = ValidateBgpPeerYaml(bgpPeer); err != nil {
				err = errors.New("LoadFromYaml: " + err.Error())
				return nil, err
			}

			return bgpPeer, nil
		}
	case TypeAnycast:
		{
			var anycast AnycastObject

			if err = yaml.Unmarshal(data, &anycast); err != nil {
				err = errors.New("LoadFromYaml: Failed to unmarshal to anycast: " + err.Error())
				return nil, err
			}

			return anycast, nil
		}
	}

	err = errors.New("LoadFromYaml: Unknown error")
	return nil, err
}
