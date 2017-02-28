package lib

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v2"
	"net"
	"os"
	"os/exec"
	"strings"
)

const (
	AF_INET  int = 4
	AF_INET6 int = 6
)

func ReadFile(fname string) ([]byte, error) {
	var (
		fs    os.FileInfo
		fd    *os.File
		data  []byte
		nread int
		err   error
	)

	if fs, err = os.Stat(fname); err != nil {
		err = errors.New("ReadFile: os.Stat() failed: " + err.Error())
		return nil, err
	}

	if fs.IsDir() {
		err = errors.New("ReadFile: os.Stat() failed: is a directory")
		return nil, err
	}

	if fd, err = os.Open(fname); err != nil {
		err = errors.New("ReadFile: os.Open() failed: " + err.Error())
		return nil, err
	}

	data = make([]byte, fs.Size())
	if nread, err = fd.Read(data); err != nil {
		err = errors.New("ReadFile: fd.Read() failed: " + err.Error())
		return nil, err
	}

	if int64(nread) != fs.Size() {
		err = errors.New("ReadFile: fd.Read() failed: short read")
		return nil, err
	}

	return data, nil
}

func DumpYaml(data interface{}) ([]byte, error) {
	var (
		output []byte
		err    error
	)

	if output, err = yaml.Marshal(&data); err != nil {
		err = errors.New("DumpYaml: yaml.Marshal() failed: " + err.Error())
		return nil, err
	}

	return output, nil
}

func GetDefaultInterface(af int) (string, error) {
	var (
		stdout_buf bytes.Buffer
		stderr_buf bytes.Buffer
		command    string
		cmd        *exec.Cmd
		err        error
	)

	switch af {
	case AF_INET:
		{
			command = "-4 route show default"
		}
	case AF_INET6:
		{
			command = "-6 route show default"
		}
	}
	cmd = exec.Command("/sbin/ip", strings.Split(command, " ")...)
	cmd.Stdout = &stdout_buf
	cmd.Stderr = &stderr_buf

	if err = cmd.Run(); err != nil {
		err = errors.New("GetDefaultInterface: " + err.Error())
		return "", err
	}

	if stderr_buf.String() != "" {
		err = errors.New("GetDefaultInterface: " + stderr_buf.String())
		return "", err
	}

	tokens := strings.Split(stdout_buf.String(), " ")
	intf := tokens[4]

	return intf, nil
}

func GetNextHopAddress(af int) (string, error) {
	var (
		nexthop_intf string
		interfaces   []net.Interface
		address      string
		err          error
	)

	if nexthop_intf, err = GetDefaultInterface(af); err != nil {
		err = errors.New("GetNextHopAddress: " + err.Error())
		return "", err
	}

	if interfaces, err = net.Interfaces(); err != nil {
		err = errors.New("GetNextHopAddress: " + err.Error())
		return "", err
	}

	for _, intf := range interfaces {
		if intf.Name != nexthop_intf {
			continue
		}

		addrs, err := intf.Addrs()
		if err != nil {
			err = errors.New("GetNextHopAddress: " + err.Error())
			return "", err
		}

		for _, addr := range addrs {
			if af == AF_INET {
				if strings.Contains(addr.String(), ".") {
					address = strings.Split(addr.String(), "/")[0]
					break
				}
			} else {
				if strings.Contains(addr.String(), ":") {
					address = strings.Split(addr.String(), "/")[0]
					break
				}
			}
		}
	}

	return address, nil
}

func RunsOK(checkCommand string) bool {
	var (
		cmdTokens []string
		cmd       *exec.Cmd
	)

	cmdTokens = strings.Split(checkCommand, " ")
	if len(cmdTokens) == 1 {
		cmd = exec.Command(checkCommand)
	} else {
		cmd = exec.Command(cmdTokens[0], cmdTokens[1:]...)
	}

	return cmd.Run() == nil
}

func AddAnycastAddress(ip string) error {
	var (
		params []string
	)
	if strings.Contains(ip, ":") {
		params = strings.Split("-6 addr add "+ip+"/128 dev lo", " ")
	} else {
		params = strings.Split("-4 addr add "+ip+"/32 dev lo", " ")
	}

	cmd := exec.Command("/sbin/ip", params...)
	return cmd.Run()
}

func RemoveAnycastAddress(ip string) error {
	var (
		params []string
	)
	if strings.Contains(ip, ":") {
		params = strings.Split("-6 addr add "+ip+"/128 dev lo", " ")
	} else {
		params = strings.Split("-4 addr add "+ip+"/32 dev lo", " ")
	}

	cmd := exec.Command("/sbin/ip", params...)
	return cmd.Run()
}
