package cloudinit

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type MetaData struct {
	Hostname   string `yaml:"local-hostname"`
	InstanceId string `yaml:"instnace-id,omitempty"`
}

func (md *MetaData) String() (string, error) {
	data, err := yaml.Marshal(md)
	return string(data), err
}

type User struct {
	Name string `yaml:"name"`
	SSHAuthorizedKeys []string `yaml:"ssh-authorized-keys,omitempty"`
	Sudo string `yaml:"sudo,omitempty"`
}

type UserData struct {
	Hostname string
	//Password string
	Users []User `yaml:"users,omitempty"`
}

func (ud *UserData) String() (string, error) {
	data, err := yaml.Marshal(ud)
	return fmt.Sprintf("#cloud-config\n%s", string(data)), err
}

func userFromLocal() (*User, error) {
	localUser, err := user.Current()
	if err != nil {
		return &User{}, err
	}

	path := filepath.Join(localUser.HomeDir, ".ssh", "id_rsa.pub")
	ssh_authorized_key, err := ioutil.ReadFile(path)
	if err != nil {
		return &User{}, err
	}

	u := &User{
		Name: localUser.Username,
		Sudo: "ALL=(ALL) NOPASSWD:ALL",
		SSHAuthorizedKeys: []string{
			string(ssh_authorized_key),
		},
	}

	return u, nil
}

func GetMetaData(hostname string) (string, error) {
	md := &MetaData{
		Hostname: hostname,
		InstanceId: hostname,
	}

	return md.String()
}

func GetUserData(hostname string) (string, error) {
	user, err := userFromLocal()
	if err != nil {
		return "", err
	}
	ud := &UserData{
		Hostname: hostname,
		Users: []User{
			*user,
		},
	}

	return ud.String()
}
