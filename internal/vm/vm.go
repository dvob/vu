package vm

type Manager interface {
	Create(name string, config *Config) error
	Start(name string) error
	Shutdown(name string, force bool) error
	Remove(name string) error
	List() ([]State, error)
	Get(name string) (*State, error)
}

type Config struct {
	Image    string
	ISO      string
	Memory   uint64
	CPUCount uint
	Network  string
	DiskSize uint64
}

type State struct {
	Name      string
	State     string
	IPAddress string
	Images    []string
}
