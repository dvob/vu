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
