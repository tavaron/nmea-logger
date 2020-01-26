package config

type I2CConfig struct {
	deviceID                uint32
	configMap               map[string]string
	bus                     string
	primaryAddress          uint16
	secondaryAddress        uint16
	gpioAddressSwitch       uint
	gpioPrimaryAddressLevel bool
	deviceType              string
}

// necessary I2C config:
// type = i2c
// bus = /dev/i2c-*
// address = uint8
// device = string

func NewI2C(configMap map[string]string) (*I2CConfig, error) {
	return DefaultI2C(), nil
}

func DefaultI2C() *I2CConfig {
	return &I2CConfig{
		deviceID:                0x76,
		configMap:               map[string]string{},
		bus:                     "/dev/i2c-1",
		primaryAddress:          0x76,
		secondaryAddress:        0,
		gpioAddressSwitch:       0,
		gpioPrimaryAddressLevel: false,
		deviceType:              "bmxx80",
	}
}

func I2CFromInterface(cfg Config) (*I2CConfig, error) {
	return NewI2C(cfg.Map())
}

// Config interface implementation
func (config I2CConfig) Map() map[string]string {
	return config.configMap
}

func (config I2CConfig) Type() string {
	return TypeI2C
}

func (config I2CConfig) DeviceID() int64 {
	return int64(config.deviceID)
}

func (config *I2CConfig) Address() uint16 {
	return config.primaryAddress
}

func (config *I2CConfig) Bus() string {
	return config.bus
}

func (config *I2CConfig) BusPath() string {
	return config.bus
}

func (config *I2CConfig) DeviceType() string {
	return config.deviceType
}
