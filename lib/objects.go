package lib

const (
	TypeBgpPeer string = "bgpPeer"
	TypeAnycast string = "anycast"
)

type objectTypeExtractor struct {
	ApiVersion int    `yaml:"apiVersion"`
	Type       string `yaml:"type"`
}

type bgpPeerMetaObject struct {
	Name string `yaml:"name"`
}

type BgpPeerSpecObject struct {
	AsNumber int    `yaml:"asNumber"`
	IP       string `yaml:"IP"`
	IP6      string `yaml:"IP6"`
}

type BgpPeerObject struct {
	ApiVersion int               `yaml:"apiVersion"`
	Type       string            `yaml:"type"`
	Meta       bgpPeerMetaObject `yaml:"meta"`
	Spec       BgpPeerSpecObject `yaml:"spec"`
}

type anycastMetaObject struct {
	Name string `yaml:"name"`
}

type AnycastSpecObject struct {
	AsNumber    int      `yaml:"asnum"`
	IP          string   `yaml:"ip"`
	IP6         string   `yaml:"ip6"`
	HealthCheck string   `yaml:"healthCheck"`
	Peers       []string `yaml:"bgpPeers"`
}

type AnycastObject struct {
	ApiVersion int               `yaml:"apiVersion"`
	Type       string            `yaml:"type"`
	Meta       anycastMetaObject `yaml:"meta"`
	Spec       AnycastSpecObject `yaml:"spec"`
}
