package sensors

import (
	"../Error"
	"../nmea"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

type i2cInitFunc func(data *i2cConnectionData) error

type i2cConnectionData struct {
	connectionData
	Init i2cInitFunc
	bus i2c.Bus
	config I2cConfig
}


type I2cConfig struct {
	bus            			string
	primaryAddress 			uint16
	secondaryAddress		uint16
	gpioAddressSwitch		uint
	gpioPrimaryAddressLevel bool
	deviceType				string
}



func i2cReadConfig(path string, devKey string) (*i2cConnectionData, error) {
	//TODO read from JSON to data.config
	var connData i2cConnectionData = i2cConnectionData{
		connectionData: connectionData{
			DeviceID:     0,
			nmeaChan:     nil,
			intervalInMs: 0,
			Stop:         nil,
		},
		Init:			i2cInitBmxx80,
		bus:            nil,
			config:         I2cConfig{
				bus:                     "",
				primaryAddress:          0x76,
				secondaryAddress:        0,
				gpioAddressSwitch:       0,
				gpioPrimaryAddressLevel: false,
				deviceType:              "bmxx80",
			},
		}
	return &connData, nil
}




func I2cInit(nmeaChan chan<- nmea.Data, configFile string, errCh chan<- Error.Error) {

	// start default i2c bus
	if _, err := host.Init(); err != nil {
		errCh <- Error.Err(Error.Low, err)
	}
	errCh <- Error.New(Error.Debug,"host init")

	bus, err := i2creg.Open("/dev/i2c-1")
	if err != nil {
		errCh <- Error.Err(Error.Fatal, err)
	}
	defer bus.Close()
	errCh <- Error.New(Error.Debug,"i2c bus open")

	// device map
	connectionMap := make(map[uint32]*i2cConnectionData)

	// TODO for-each entry in config file
	connData, err := i2cReadConfig("", "")
	connData.errChan = errCh
	if err != nil {
		connData.errChan <- Error.Err(Error.Low, err)
	} else {
		connData.bus = bus
		connData.nmeaChan = nmeaChan
		err = connData.Init(connData) // init bmxx80
		if err != nil {
			connData.errChan <- Error.Err(Error.High, err)
			err = nil
		} else {
			defer connData.Stop()
			connectionMap[connData.DeviceID] = connData
		}
	}
	select {}
}