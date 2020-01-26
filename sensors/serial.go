package sensors

import (
	"../nmea"
	"./config"
	"bufio"
	"errors"
	"github.com/tarm/serial"
)

type SerialConnection struct {
	engine *Engine
	config *config.SerialConfig
	port   *serial.Port
	stop   bool
}

func (e *Engine) newSerialConnection(cfg config.Config) (*SerialConnection, error) {
	configuration, err := config.SerialFromInterface(cfg)
	if err != nil {
		return nil, err
	}

	return &SerialConnection{
		engine: e,
		config: configuration,
		port:   nil,
		stop:   false,
	}, nil
}

// Connection interface implementation
func (sc SerialConnection) DeviceID() int64 {
	return sc.config.DeviceID()
}

func (sc SerialConnection) Type() string {
	return sc.config.Type()
}

func (sc SerialConnection) Stop() {
	if !sc.stop {
		sc.engine.error(errors.New(
			"stopping serial sensors on " +
				sc.config.DeviceConfig().Name))

		sc.stop = true
		err := sc.port.Close()
		if err != nil {
			sc.engine.error(err)
		}
		sc.port = nil
	}
}

func (sc SerialConnection) connect() error {
	if sc.port != nil {
		sc.Stop()
	}

	port, err := serial.OpenPort(sc.config.DeviceConfig())
	if err != nil {
		return err
	}
	sc.port = port
	sc.stop = false
	go sc.readRoutine()

	return nil
}

func (sc *SerialConnection) readRoutine() {
	defer sc.Stop()

	var err error
	for !sc.stop && err == nil {
		var line string
		line, err = sc.readLine()
		if err != nil {
			sc.engine.error(err)
		} else if nmea.GetType(line) == "$GPRMC" || nmea.GetType(line) == "$--RMC" {
			data, err := nmea.NewData(line, sc.DeviceID())
			if err != nil {
				sc.engine.error(err)
			} else {
				sc.engine.nmeaChan <- data
			}
		}
	}
}

func (sc *SerialConnection) readLine() (string, error) {
	scanner := bufio.NewScanner(sc.port)
	for scanner.Scan() {
		if scanner.Err() != nil {
			return "", scanner.Err()
		}
		return scanner.Text(), nil
	}
	return "", errors.New("unexpectedly reached end of serialReadLine")
}
