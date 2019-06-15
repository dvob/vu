package cloudinit

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Config represents all configurations for cloud init
type Config struct {
	MetaData      MetaData
	UserData      UserData
	NetworkConfig *NetworkConfig
}

func NewDefaultConfig(name, user, sshAuthKey string) *Config {
	return &Config{
		MetaData: MetaData{
			Hostname:   name,
			InstanceId: name,
		},
		UserData: UserData{
			Hostname: name,
			Users: []User{
				User{
					Name: user,
					Sudo: "ALL=(ALL) NOPASSWD:ALL",
					SSHAuthorizedKeys: []string{
						sshAuthKey,
					},
				},
			},
		},
	}
}

func (c *Config) String() (string, error) {
	var (
		meta    string
		user    string
		network string
		err     error
	)
	meta, err = c.MetaData.String()
	if err != nil {
		fmt.Errorf("could not render meta data: %s", err)
	}

	user, err = c.UserData.String()
	if err != nil {
		fmt.Errorf("could not render user data: %s", err)
	}

	if c.NetworkConfig != nil {
		network, err = c.NetworkConfig.String()
		if err != nil {
			fmt.Errorf("could not render network config: %s", err)
		}
	}
	return fmt.Sprintf("### meta-data ###\n%s\n### user-data ###\n%s\n### network-config ###\n%s\n", meta, user, network), nil
}

// MetaData

type MetaData struct {
	Hostname   string `yaml:"local-hostname"`
	InstanceId string `yaml:"instnace-id,omitempty"`
}

func (md *MetaData) String() (string, error) {
	data, err := yaml.Marshal(md)
	return string(data), err
}

// UserData

type UserData struct {
	Hostname string
	//Password string
	Users []User `yaml:"users,omitempty"`
}

func (ud *UserData) String() (string, error) {
	data, err := yaml.Marshal(ud)
	return fmt.Sprintf("#cloud-config\n%s", string(data)), err
}

type User struct {
	Name              string   `yaml:"name"`
	SSHAuthorizedKeys []string `yaml:"ssh-authorized-keys,omitempty"`
	Sudo              string   `yaml:"sudo,omitempty"`
}

// NetworkConfig

type NetworkConfig struct{}

func (nc *NetworkConfig) String() (string, error) {
	return "", nil
}
