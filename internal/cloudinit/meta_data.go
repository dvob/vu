package cloudinit

// MetaData is a struct to render the meta data of the cloud init configuration
type MetaData struct {
	Raw        map[string]interface{} `json:"-"`
	Hostname   string                 `json:"local-hostname,omitempty"`
	InstanceID string                 `json:"instnace-id,omitempty"`
}

func (md *MetaData) Marshal() ([]byte, error) {
	return mergeMarshal(md, md.Raw)
}

func (md *MetaData) Unmarshal(data []byte) error {
	return rawUnmarshal(data, md, &md.Raw)
}
