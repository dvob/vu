package cloudinit

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/matryer/is"
)

func Test_UserData_Marshaling(t *testing.T) {
	is := is.New(t)

	userdata := `#cloud-config
unsupported_field: blabla
users:
- name: sepp
  lock_passwd: false
`
	ud := &UserData{}
	err := ud.Unmarshal([]byte(userdata))
	is.NoErr(err)

	is.Equal(ud.Raw["unsupported_field"], "blabla") // unsupported_field was read
	is.Equal(len(ud.Users), 1)                      // one user in users
	is.Equal(ud.Users[0].Name, "sepp")              // first users is sepp

	ud.Users[0].Name = "vreni"

	output, err := ud.Marshal()
	is.NoErr(err)

	raw := map[string]interface{}{}
	err = yaml.Unmarshal(output, &raw)
	is.NoErr(err)

	is.Equal(raw["unsupported_field"], "blabla") // unsupported filed is still here

	var ud1 UserData
	err = ud1.Unmarshal([]byte(output))
	is.NoErr(err)

	is.Equal(len(ud1.Users), 1)          // one user in users
	is.Equal(ud1.Users[0].Name, "vreni") // vreni overrides sepp
}

func Test_MergeConfig(t *testing.T) {
	is := is.New(t)

	gw1 := "192.168.1.1"
	c1 := &Config{
		UserData: &UserData{
			Raw: map[string]interface{}{
				"final_message": "Bla bli blo",
			},
			Users: []User{
				{
					Name: "john",
				},
			},
		},
		NetworkConfig: &NetworkConfig{
			Ethernets: map[string]Ethernet{
				"default": {
					Gateway: &gw1,
					DNS: &DNS{
						Servers: []string{"8.8.8.8"},
					},
				},
			},
		},
	}

	gw2 := "10.0.0.1"
	c2 := &Config{
		MetaData: &MetaData{
			Hostname: "myserver.example.com",
		},
		NetworkConfig: &NetworkConfig{
			Ethernets: map[string]Ethernet{
				"default": {
					Gateway: &gw2,
				},
			},
		},
	}

	expectedConfig := &Config{
		MetaData: &MetaData{
			Hostname: "myserver.example.com",
		},
		UserData: &UserData{
			Raw: map[string]interface{}{
				"final_message": "Bla bli blo",
			},
			Users: []User{
				{
					Name: "john",
				},
			},
		},
		NetworkConfig: &NetworkConfig{
			Ethernets: map[string]Ethernet{
				"default": {
					Gateway: &gw2,
					DNS: &DNS{
						Servers: []string{"8.8.8.8"},
					},
				},
			},
		},
	}
	expectedOutput, err := expectedConfig.String()
	is.NoErr(err) // failed to render expectedOutput

	err = c1.Merge(c2)
	is.NoErr(err) // failed to merge c2 into c1

	c1Output, err := c1.String()
	is.NoErr(err) // failed to render c1Output
	is.Equal(expectedOutput, c1Output)
}
