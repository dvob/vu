package cloudinit

import (
	"encoding/json"
	"io/ioutil"
	"reflect"

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

func merge(m1, m2 Marshaler) error {
	// nothing to merge
	if m2 == nil || reflect.ValueOf(m2).IsNil() {
		return nil
	}
	m1Data, err := m1.Marshal()
	if err != nil {
		return err
	}
	m1Raw := map[string]interface{}{}
	err = yaml.Unmarshal(m1Data, &m1Raw)
	if err != nil {
		return err
	}

	m2Data, err := m2.Marshal()
	if err != nil {
		return err
	}
	m2Raw := map[string]interface{}{}
	err = yaml.Unmarshal(m2Data, &m2Raw)
	if err != nil {
		return err
	}

	err = mergo.Merge(&m1Raw, &m2Raw, mergo.WithOverride)
	if err != nil {
		return err
	}

	data, err := json.Marshal(&m1Raw)
	if err != nil {
		return err
	}
	return m1.Unmarshal(data)
}

func rawUnmarshal(data []byte, o interface{}, raw *map[string]interface{}) error {
	err := yaml.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, o)
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
	return ioutil.WriteFile(file, data, 0o640)
}
