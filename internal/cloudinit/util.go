package cloudinit

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
)

type Marshaler interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func structToMap(s interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func mergeMarshal(overrides interface{}, raw map[string]interface{}) ([]byte, error) {
	overridesMap, err := structToMap(overrides)
	if err != nil {
		return nil, err
	}

	err = mergo.Merge(&raw, &overridesMap, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(&raw)
}

func rawUnmarshal(data []byte, o interface{}, raw map[string]interface{}) error {
	err := yaml.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, o)
}

// unmarshalFromFile unmarshals the content of a file if the file exists
func unmarshalFromFile(file string, m Marshaler) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return m.Unmarshal(data)
}

// marshalToFile marshals to a file if the marshaler is not nil
func marshalToFile(file string, m Marshaler) error {
	if m == nil {
		return nil
	}
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, data, 0640)
}
