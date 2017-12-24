package bgp

import (
	"errors"
	"github.com/r3boot/anycast-agent/lib"
	"github.com/r3boot/anycast-agent/lib/bgp/bgp2go"
	"strings"
	"time"
)

type BGP struct {
	context     bgp2go.BGPContext
	cmdToPeer   chan bgp2go.BGPProcessMsg
	cmdFromPeer chan bgp2go.BGPProcessMsg
	peers       []string
}

type BGPConfig struct {
	Asnum      int
	RouterId   string
	NextHopIP  string
	NextHopIP6 string
	LocalPref  int
	BgpPeers   []string
}

var Logger lib.Logger

func NewBGP(logger lib.Logger) (*BGP, error) {
	var (
		bgp *BGP
	)

	Logger = logger

	bgp = &BGP{
		context:     bgp2go.BGPContext{},
		cmdToPeer:   make(chan bgp2go.BGPProcessMsg),
		cmdFromPeer: make(chan bgp2go.BGPProcessMsg),
	}

	return bgp, nil
}

// Configure this side of the BGP routine
func (bgp *BGP) Initialize(cfg *BGPConfig) (err error) {

	bgp.context.ASN = uint32(cfg.Asnum)
	bgp.context.ListenLocal = true
	bgp.peers = cfg.BgpPeers

	bgp.context.RouterID, err = bgp2go.IPv4ToUint32(cfg.RouterId)
	if err != nil {
		err = errors.New("bgp.Initialize(): Failed to parse RouterID: " + err.Error())
		return
	}

	bgp.context.NextHop, err = bgp2go.IPv4ToUint32(cfg.NextHopIP)
	if err != nil {
		err = errors.New("bgp.Initialize(): Failed to parse IPv4 NextHop: " + err.Error())
		return
	}

	bgp.context.NextHopV6, err = bgp2go.IPv6StringToAddr(cfg.NextHopIP6)
	if err != nil {
		err = errors.New("bgp.Initialize(): Failed to parse IPv6 NextHop: " + err.Error())
		return
	}

	bgp.context.LocalPref = uint32(cfg.LocalPref)

	return
}

func (bgp *BGP) ServerRoutine() {
	var (
		bgpPeer string
	)

	bgp.cmdToPeer = make(chan bgp2go.BGPProcessMsg)
	bgp.cmdFromPeer = make(chan bgp2go.BGPProcessMsg)

	Logger.Debug("bgp: Starting ServerRoutine")
	go bgp2go.StartBGPProcess(bgp.cmdToPeer, bgp.cmdFromPeer, bgp.context)

	time.Sleep(1 * time.Second)
	for _, bgpPeer = range bgp.peers {
		Logger.Debug("bgp: Adding bgp peer %s", bgpPeer)
		bgp.AddNeighbor(bgpPeer)
	}
}

func (bgp *BGP) addv4Neighbor(ipaddr string) {
	Logger.Debug("bgp: Adding IPv4 neighbor " + ipaddr)
	bgp.cmdToPeer <- bgp2go.BGPProcessMsg{
		Cmnd: "AddNeighbour",
		Data: ipaddr + " inet",
	}
}

func (bgp *BGP) addv6Neighbor(ipaddr string) {
	Logger.Debug("bgp: Adding IPv6 neighbor " + ipaddr)
	bgp.cmdToPeer <- bgp2go.BGPProcessMsg{
		Cmnd: "AddNeighbour",
		Data: "[" + ipaddr + "] inet6",
	}
}

func (bgp *BGP) AddNeighbor(ipaddr string) {
	if strings.Contains(ipaddr, ":") {
		bgp.addv6Neighbor(ipaddr)
	} else {
		bgp.addv4Neighbor(ipaddr)
	}
}

func (bgp *BGP) removev4Neighbor(ipaddr string) {
	Logger.Debug("bgp: Removing IPv4 neighbor " + ipaddr)
	bgp.cmdToPeer <- bgp2go.BGPProcessMsg{
		Cmnd: "RemoveNeighbour",
		Data: ipaddr + " inet",
	}
}

func (bgp *BGP) removev6Neighbor(ipaddr string) {
	Logger.Debug("bgp: Removing IPv6 neighbor " + ipaddr)
	bgp.cmdToPeer <- bgp2go.BGPProcessMsg{
		Cmnd: "RemoveNeighbour",
		Data: "[" + ipaddr + "] inet6",
	}
}

func (bgp *BGP) RemoveBGPNeighbor(ipaddr string) {
	if strings.Contains(ipaddr, ":") {
		bgp.removev6Neighbor(ipaddr)
	} else {
		bgp.removev4Neighbor(ipaddr)
	}
}

func (bgp *BGP) addv4Route(prefix string) {
	Logger.Debug("bgp: Adding IPv4 prefix " + prefix)
	bgp.cmdToPeer <- bgp2go.BGPProcessMsg{
		Cmnd: "AddV4Route",
		Data: prefix,
	}
}

func (bgp *BGP) addv6Route(prefix string) {
	Logger.Debug("bgp: Adding IPv6 prefix " + prefix)
	bgp.cmdToPeer <- bgp2go.BGPProcessMsg{
		Cmnd: "AddV6Route",
		Data: prefix,
	}
}

func (bgp *BGP) AddRoute(prefix string) {
	prefix = add_cidr_mask(prefix)
	if strings.Contains(prefix, ":") {
		bgp.addv6Route(prefix)
	} else {
		bgp.addv4Route(prefix)
	}
}

func (bgp *BGP) removev4Route(prefix string) {
	Logger.Debug("bgp: Removing IPv4 prefix " + prefix)
	bgp.cmdToPeer <- bgp2go.BGPProcessMsg{
		Cmnd: "WithdrawV4Route",
		Data: prefix,
	}
}

func (bgp *BGP) removev6Route(prefix string) {
	Logger.Debug("bgp: Removing IPv6 prefix" + prefix)
	bgp.cmdToPeer <- bgp2go.BGPProcessMsg{
		Cmnd: "WithdrawV6Route",
		Data: prefix,
	}
}

func (bgp *BGP) RemoveRoute(prefix string) {
	prefix = add_cidr_mask(prefix)
	if strings.Contains(prefix, ":") {
		bgp.removev6Route(prefix)
	} else {
		bgp.removev4Route(prefix)
	}
}
