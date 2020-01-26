package config

import "github.com/tarm/serial"

type SerialConfig struct {
	deviceID     uint32
	configMap    map[string]string
	deviceConfig *serial.Config
}

// necessary serial config:
// type = serial
// path = /dev/tty*
// baud = int
// size = int
// parity = int
// stop = int

func NewSerial(configMap map[string]string) (*SerialConfig, error) {
	return DefaultSerial(), nil
}
func DefaultSerial() *SerialConfig {
	return &SerialConfig{
		deviceConfig: &serial.Config{
			Name:        "/dev/ttyAMA0",
			Baud:        9600,
			ReadTimeout: 0,
			Size:        8,
			Parity:      serial.ParityNone,
			StopBits:    serial.Stop1,
		},
		configMap: map[string]string{},
	}
}

func SerialFromInterface(cfg Config) (*SerialConfig, error) {
	return NewSerial(cfg.Map())
}

func (config *SerialConfig) DeviceConfig() *serial.Config {
	return config.deviceConfig
}

// Config interface implementation
func (config SerialConfig) Map() map[string]string {
	return config.configMap
}

func (config SerialConfig) Type() string {
	return TypeSerial
}

func (config SerialConfig) DeviceID() int64 {
	return int64(config.deviceID)
}
