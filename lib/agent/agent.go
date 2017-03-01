package agent

import (
	"errors"
	"github.com/r3boot/anycast-agent/lib"
	"github.com/r3boot/anycast-agent/lib/bgp"
	"github.com/r3boot/anycast-agent/lib/healthcheck"
	"time"
)

type AnycastAgent struct {
	Name        string
	Logger      lib.Logger
	LocalAs     int
	NextHopIP   string
	NextHopIP6  string
	IP          string
	IP6         string
	BgpPeers    []string
	bgpService  *bgp.BGP
	healthCheck *healthcheck.HealthCheck
}

func NewAnycastAgent(endpoints []string, profile string) (*AnycastAgent, error) {
	var (
		agent *AnycastAgent
		err   error
	)

	agent = &AnycastAgent{
		Name:   profile,
		Logger: lib.NewLogger(true),
	}

	if err = agent.Initialize(endpoints); err != nil {
		err = errors.New("NewAnycastAgent: " + err.Error())
		return nil, err
	}

	return agent, nil
}

func (aa *AnycastAgent) Initialize(endpoints []string) error {
	var (
		etcd         *lib.EtcdClient
		hcResultChan chan bool
		err          error
	)

	if etcd, err = lib.NewEtcdClient(endpoints, "/am/v1"); err != nil {
		err = errors.New("Initialize: " + err.Error())
		return err
	}

	object, err := etcd.GetObject(lib.TypeAnycast, aa.Name)
	if err != nil {
		err = errors.New("Initialize: " + err.Error())
		return err
	}

	aa.LocalAs = object.(lib.AnycastObject).Spec.AsNumber
	aa.IP = object.(lib.AnycastObject).Spec.IP
	aa.IP6 = object.(lib.AnycastObject).Spec.IP6

	if aa.NextHopIP, err = lib.GetNextHopAddress(lib.AF_INET); err != nil {
		aa.Logger.Warn("AnycastAgent: No ipv4 next-hop address found: " + err.Error())
	}

	if aa.NextHopIP6, err = lib.GetNextHopAddress(lib.AF_INET6); err != nil {
		aa.Logger.Warn("AnycastAgent: No ipv6 next-hop address found: " + err.Error())
	}

	all_objects, err := etcd.GetAllObjects(lib.TypeBgpPeer, etcd.Prefix+"/peers")
	if err != nil {
		aa.Logger.Error("AnycastAgent: Failed to retrieve bgp peers: " + err.Error())
	}

	for _, peer := range all_objects {
		spec := peer.(lib.BgpPeerObject).Spec
		if spec.IP != "" {
			aa.BgpPeers = append(aa.BgpPeers, spec.IP)
		}
		if spec.IP6 != "" {
			aa.BgpPeers = append(aa.BgpPeers, spec.IP6)
		}
	}

	hcResultChan = make(chan bool, 10)
	aa.healthCheck = healthcheck.NewHealthCheck(aa.Logger, healthcheck.HealthCheckConfig{
		Command:     object.(lib.AnycastObject).Spec.HealthCheck,
		Interval:    3 * time.Second,
		InitDamping: 2,
		MaxRetries:  5,
		ResultChan:  hcResultChan,
	})

	if aa.bgpService, err = bgp.NewBGP(aa.Logger); err != nil {
		aa.Logger.Error("AnycastAgent: Failed to get BGP service: " + err.Error())
	}

	err = aa.bgpService.Initialize(&bgp.BGPConfig{
		Asnum:      aa.LocalAs,
		RouterId:   aa.NextHopIP,
		NextHopIP:  aa.NextHopIP,
		NextHopIP6: aa.NextHopIP6,
		LocalPref:  100,
		BgpPeers:   aa.BgpPeers,
	})
	if err != nil {
		aa.Logger.Error("AnycastAgent: Failed to initialize BGP service: " + err.Error())
	}

	return nil
}

func (aa *AnycastAgent) isHealthy(results []bool) bool {
	if aa.healthCheck.Health {
		for i := 0; i < aa.healthCheck.Config.MaxRetries; i++ {
			if results[i] {
				return true
			}
		}
		return false
	} else {
		for i := 0; i < aa.healthCheck.Config.InitDamping; i++ {
			if !results[i] {
				return false
			}
		}
		return true
	}

	return false
}

func (aa *AnycastAgent) RunAnycastService() {
	var (
		lastResults []bool
		health      bool
		stateChan   chan bool
		err         error
	)

	numItems := aa.healthCheck.Config.MaxRetries
	lastResults = make([]bool, numItems, numItems)
	aa.healthCheck.Health = false
	health = false

	stateChan = make(chan bool, 1)

	go aa.healthCheck.CheckRoutine()
	go aa.bgpService.ServerRoutine()

	for {
		select {
		case result := <-aa.healthCheck.Config.ResultChan:
			{
				numItems := aa.healthCheck.Config.MaxRetries

				for i := numItems - 1; i > 0; i-- {
					lastResults[i] = lastResults[i-1]
				}
				lastResults[0] = result

				if aa.healthCheck.Health {
					health = false
					for i := 0; i < aa.healthCheck.Config.MaxRetries; i++ {
						if lastResults[i] {
							health = true
						}
					}
				} else {
					health = true
					for i := 0; i < aa.healthCheck.Config.InitDamping; i++ {
						if !lastResults[i] {
							health = false
						}
					}
				}

				if health != aa.healthCheck.Health {
					aa.healthCheck.Health = health
					stateChan <- health
				}
			}
		case curState := <-stateChan:
			{
				if curState {
					aa.Logger.Debug("AnycastAgent: State changed to UP")
					if aa.IP != "" {
						aa.bgpService.AddRoute(aa.IP)
					}
					if aa.IP6 != "" {
						aa.bgpService.AddRoute(aa.IP6)
					}
				} else {
					aa.Logger.Debug("AnycastAgent: State changed to DOWN")
					if aa.IP != "" {
						aa.bgpService.RemoveRoute(aa.IP)
					}
					if aa.IP6 != "" {
						aa.bgpService.RemoveRoute(aa.IP6)
					}
				}
			}
		}
	}
}
