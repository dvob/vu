package cloudinit

import (
	"net"

	"gopkg.in/yaml.v2"
)

type NetworkParameter struct {
	Address    string
	Gateway    string
	Nameserver []string
	DHCP       bool
}

type NetworkConfig struct {
	Version   int                 `yaml:"version"`
	Ethernets map[string]Ethernet `yaml:"ethernets"`
}

type Ethernet struct {
	Match     *Match   `yaml:"match,omitempty"`
	Addresses []string `yaml:"addresses,omitempty"`
	DHCP      *bool    `yaml:"dhcp4,omitempty"`
	Gateway   *string  `yaml:"gateway4,omitempty"`
	DNS       *DNS     `yaml:"nameservers,omitempty"`
}

type Match struct {
	Name *string `yaml:"name,omitempty"`
	MAC  *string `yaml:"macaddress,omitempty"`
}

type DNS struct {
	Servers []string `yaml:"addresses"`
	Search  []string `yaml:"search,omitempty"`
}

func (c *NetworkConfig) String() (string, error) {
	data, err := yaml.Marshal(c)
	return string(data), err
}

func NewNetworkConfig(np *NetworkParameter) (*NetworkConfig, error) {
	var (
		matchName  = "en*"
		gateway    string
		nameserver = []string{}
	)

	if np.Address == "" {
		return nil, nil
	}

	_, ipNet, err := net.ParseCIDR(np.Address)
	if err != nil {
		return nil, err
	}

	if np.Gateway == "" {
		gateway = getGatewayIP(ipNet).String()
	} else {
		gateway = np.Gateway
	}

	if np.Nameserver == nil || len(np.Nameserver) == 0 {
		nameserver = append(nameserver, gateway)
	} else {
		nameserver = np.Nameserver
	}

	c := &NetworkConfig{
		Version: 2,
		Ethernets: map[string]Ethernet{
			"default": Ethernet{
				Match: &Match{
					Name: &matchName,
				},
				Addresses: []string{np.Address},
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

// increments IPhttps://stackoverflow.com/a/49057611
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
