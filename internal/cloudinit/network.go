package cloudinit

import (
	"net"
)

type NetworkConfig struct {
	Raw       map[string]interface{} `json:"-"`
	Version   int                    `json:"version"`
	Ethernets map[string]Ethernet    `json:"ethernets"`
}

type Ethernet struct {
	Match     *Match   `json:"match,omitempty"`
	Addresses []string `json:"addresses,omitempty"`
	DHCP      *bool    `json:"dhcp4,omitempty"`
	Gateway   *string  `json:"gateway4,omitempty"`
	DNS       *DNS     `json:"nameservers,omitempty"`
}

type Match struct {
	Name *string `json:"name,omitempty"`
	MAC  *string `json:"macaddress,omitempty"`
}

type DNS struct {
	Servers []string `json:"addresses"`
	Search  []string `json:"search,omitempty"`
}

func (nc *NetworkConfig) Marshal() ([]byte, error) {
	return mergeMarshal(nc, nc.Raw)
}

func (nc *NetworkConfig) Unmarshal(data []byte) error {
	return rawUnmarshal(data, nc, &nc.Raw)
}

func (nc *NetworkConfig) Merge(nc2 *NetworkConfig) error {
	return merge(nc, nc2)
}

type NetworkConfigOptions struct {
	Address    string
	Gateway    string
	Nameserver []string
}

func NewNetworkConfig(nco NetworkConfigOptions) (*NetworkConfig, error) {
	var (
		// will only work for one interface
		matchName  = "en*"
		gateway    string
		nameserver = []string{}
	)

	if nco.Address == "" {
		return nil, nil
	}

	_, ipNet, err := net.ParseCIDR(nco.Address)
	if err != nil {
		return nil, err
	}

	if nco.Gateway == "" {
		gateway = getGatewayIP(ipNet).String()
	} else {
		gateway = nco.Gateway
	}

	if nco.Nameserver == nil || len(nco.Nameserver) == 0 {
		nameserver = append(nameserver, gateway)
	} else {
		nameserver = nco.Nameserver
	}

	c := &NetworkConfig{
		Version: 2,
		Ethernets: map[string]Ethernet{
			"default": {
				Match: &Match{
					Name: &matchName,
				},
				Addresses: []string{nco.Address},
				Gateway:   &gateway,
				DNS: &DNS{
					Servers: nameserver,
				},
			},
		},
	}
	return c, nil
}

func getGatewayIP(ipNet *net.IPNet) net.IP {
	return incrementIP(ipNet.IP, 1)
}

// incrementIP increments an IP https://stackoverflow.com/a/49057611
func incrementIP(ip net.IP, inc uint) net.IP {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v += inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)
	return net.IPv4(v0, v1, v2, v3)
}
