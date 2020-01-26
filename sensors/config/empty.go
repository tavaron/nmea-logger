package config

type EmptyConfig struct{}

func NewEmpty(configMap map[string]string) (*EmptyConfig, error) {
	return &EmptyConfig{}, nil
}

func (config EmptyConfig) DeviceID() int64 {
	return 0
}

func (config EmptyConfig) Map() map[string]string {
	return map[string]string{}
}

func (config EmptyConfig) Type() string {
	return TypeEmpty
}
