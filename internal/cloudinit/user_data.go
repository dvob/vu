package cloudinit

import "bytes"

// UserData is a struct to render the user data of the cloud init configuration
type UserData struct {
	Raw map[string]interface{} `json:"-"`
	//Hostname string
	//Password        string `yaml:"password,omitempty"`
	//SSHPasswordAuth bool   `yaml:"ssh_pwauth,omitempty"`
	Users []User `json:"users,omitempty"`
}

// User definition of cloud init configuration
type User struct {
	Name              string   `json:"name"`
	SSHAuthorizedKeys []string `json:"ssh-authorized-keys,omitempty"`
	Sudo              string   `json:"sudo,omitempty"`
	LockPasswd        *bool    `json:"lock_passwd,omitempty"`
	Passwd            string   `json:"passwd,omitempty"`
}

func (ud *UserData) Marshal() ([]byte, error) {
	data, err := mergeMarshal(ud, ud.Raw)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	buf.WriteString("#cloud-config\n")
	_, err = buf.Write(data)
	return buf.Bytes(), err
}

func (ud *UserData) Unmarshal(data []byte) error {
	return rawUnmarshal(data, ud, &ud.Raw)
}

func (ud *UserData) Merge(ud2 *UserData) error {
	return merge(ud, ud2)
}
