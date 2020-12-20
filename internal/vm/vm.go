package vm

const descriptionPrefix = "created by vu"

type Config struct {
	Name      string
	BaseImage string
	ConfigISO string
	Memory    uint
	VCPU      uint
	Network   string
	DiskSize  uint64
}

type VM struct {
	Name      string
	State     string
	IPAddress string
}
