package sensors

import (
	"../Error"
	"../nmea"
	"./config"
	"periph.io/x/periph/conn/i2c"
)

const (
	// i2c device types
	devBmxx80 string = "bmxx80"

	ErrFlag    string = "[sensors]"
	ErrI2CFlag string = "[I2C]"
)

type Engine struct {
	errorChan          chan<- *Error.Error
	nmeaChan           chan<- *nmea.Data
	intervalInMs       uint
	connList           map[int64]Connection
	i2cHostInitialized bool
	i2cBuses           map[string]i2c.BusCloser
}

type Connection interface {
	DeviceID() int64
	Type() string
	Stop()
	connect() error
}

func NewEngine(nmeaChan chan<- *nmea.Data, errorChan chan<- *Error.Error) *Engine {
	cd := &Engine{
		errorChan:          errorChan,
		nmeaChan:           nmeaChan,
		connList:           map[int64]Connection{},
		i2cHostInitialized: false,
		i2cBuses:           map[string]i2c.BusCloser{},
	}

	return cd
}

func (e *Engine) Connect(cfg config.Config) Connection {
	var conn Connection
	switch cfg.Type() {
	case config.TypeSerial:
		sConn, err := e.newSerialConnection(cfg)
		if err != nil {
			e.error(err)
		}
		conn = sConn
	case config.TypeI2C:
		iConn, err := e.newI2CConnection(cfg)
		if err != nil {
			e.error(err)
		}
		conn = iConn
	}
	if conn == nil {
		return nil
	}
	err := conn.connect()
	if err != nil {
		e.error(err)
	}

	e.connList[conn.DeviceID()] = conn
	return conn

}

func (e *Engine) Stop() error {
	//TODO
	return nil
}

func (e *Engine) error(err error, lvl ...Error.Level) {
	level := Error.Debug
	if len(lvl) > 0 {
		level = lvl[0]
	}
	e.errorChan <- Error.Err(level, err, ErrFlag)
}

func (e *Engine) i2cBusIsOpen(path string) bool {
	for key, _ := range e.i2cBuses {
		if key == path {
			return true
		}
	}
	return false
}
