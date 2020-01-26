package config

import "errors"

const (
	ErrFlag    string = "[DevConfig]"
	TypeSerial string = "serial"
	TypeI2C    string = "i2c"
	TypeEmpty  string = "empty"
	ParamType  string = "type"
)

var (
	ErrInvalidConfigMap error = errors.New(ErrFlag + ": invalid config map received")
)

type Config interface {
	Map() map[string]string
	Type() string
	DeviceID() int64
}

func NewConfig(configMap map[string]string) (*Config, error) {

	if len(configMap) <= 0 {
		return nil, ErrInvalidConfigMap
	}

	var result Config
	var err error = nil

	for key, value := range configMap {
		if key == ParamType {
			switch value {
			case TypeSerial:
				result, err = NewSerial(configMap)
			case TypeI2C:
				//TODO I2C config
			case TypeEmpty:
				result, err = NewEmpty(configMap)
			}
			break
		}
	}

	return &result, err

}
