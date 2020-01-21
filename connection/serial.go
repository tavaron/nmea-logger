package sensors

import (
	"bufio"
	"errors"
	"github.com/tarm/serial"
	"math"
	"strings"

	"../Error"
	"../nmea"
)

type serialInitFunc func(data *serialConnectionData) error

type serialConnectionData struct {
	connectionData
	config *serial.Config
	port   *serial.Port
}

func SerialInit(nmeaChan chan<- *nmea.Data, configFile string, errCh chan<- *Error.Error) {
	conn := &serialConnectionData{
		connectionData: connectionData{
			DeviceID:     0 + math.MaxInt16,
			nmeaChan:     nmeaChan,
			intervalInMs: 0,
			Stop:         nil,
			errorChan:    errCh,
		},
		config: &serial.Config{
			Name:        "/dev/ttyAMA0",
			Baud:        9600,
			ReadTimeout: 0,
			Size:        8,
			Parity:      serial.ParityNone,
			StopBits:    serial.Stop1,
		},
		port: nil,
	}

	port, err := serial.OpenPort(conn.config)
	conn.port = port
	if err != nil {
		conn.errorChan <- Error.Err(Error.Low, err)
		err = nil
	}
	conn.Stop = conn.port.Close

	go serialGpsRead(conn)
}

func serialGpsRead(conn *serialConnectionData) {
	defer conn.Stop()

	for {
		line, err := serialReadLine(conn.port)
		if err != nil {
			conn.errorChan <- Error.Err(Error.Low, err)
		} else if getNmeaType(line) == "$GPRMC" || getNmeaType(line) == "$--RMC" {
			data, err := nmea.NewData(line, conn.DeviceID)
			if err != nil {
				conn.errorChan <- Error.Err(Error.Low, err)
			} else {
				conn.nmeaChan <- data
			}
		}
	}
}

func getNmeaType(s string) string {
	sub := strings.Split(s, ",")
	if len(sub) > 0 {
		return sub[0]
	}
	return ""
}

func serialReadLine(port *serial.Port) (string, error) {
	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		if scanner.Err() != nil {
			return "", scanner.Err()
		}
		return scanner.Text(), nil
	}
	return "", errors.New("unexpectedly reached end of serialReadLine")
}
