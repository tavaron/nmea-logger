package sensors

import (
	"errors"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"

	"../Error"
	"../nmea"

	"./config"
)

type i2cReadFunc func(conn *I2CConnection) error
type i2cStopFunc func() error

type I2CConnection struct {
	engine    *Engine
	read      i2cReadFunc
	stop      i2cStopFunc
	bus       i2c.Bus
	config    *config.I2CConfig
	isStopped bool
}

func (ic I2CConnection) DeviceID() int64 {
	return ic.config.DeviceID()
}

func (ic I2CConnection) Type() string {
	return ic.config.Type()
}

func (ic I2CConnection) Stop() {
	if !ic.isStopped {
		ic.error(errors.New(
			"stopping i2c sensor " +
				ic.config.DeviceType()))

		ic.isStopped = true
		err := ic.stop()
		if err != nil {
			ic.error(err, Error.Low)
		}
	}
}

func (ic I2CConnection) connect() error {
	return ic.read(&ic)
}

func (ic *I2CConnection) IsStopped() bool {
	return ic.isStopped
}

func (e *Engine) newI2CConnection(cfg config.Config) (*I2CConnection, error) {
	configuration, err := config.I2CFromInterface(cfg)
	if err != nil {
		return nil, err
	}

	conn := &I2CConnection{
		engine: e,
		read:   nil,
		stop:   nil,
		bus:    nil,
		config: configuration,
	}

	switch configuration.DeviceType() {
	case devBmxx80:
		conn.read = readBmxx80
	default:
		return nil, errors.New(
			ErrI2CFlag + ": device type " +
				configuration.DeviceType() + " not supported")
	}

	err = e.i2cBusInit(configuration.BusPath())
	if err != nil {
		e.error(err)
	}
	conn.bus = e.i2cBuses[configuration.BusPath()]

	return conn, nil
}

func (e *Engine) i2cBusInit(path string) error {

	if !e.i2cHostInitialized {
		if _, err := host.Init(); err != nil {
			return err
		}
		e.i2cHostInitialized = true
		e.errorChan <- Error.New(Error.Debug, "host initialized")
	}

	if !e.i2cBusIsOpen(path) {
		bus, err := i2creg.Open(path)
		if err != nil {
			return err
		}
		e.i2cBuses[path] = bus
		e.errorChan <- Error.New(Error.Debug, "opened "+path)
	}
	return nil
}

func (ic I2CConnection) error(err error, lvl ...Error.Level) {
	errLvl := Error.Debug
	if len(lvl) > 0 {
		errLvl = lvl[0]
	}
	newErr := errors.New(ErrI2CFlag + err.Error())
	ic.engine.error(newErr, errLvl)
}

func (ic *I2CConnection) send(data *nmea.Data) {
	ic.engine.nmeaChan <- data
}
