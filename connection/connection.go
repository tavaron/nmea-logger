package sensors

import (
	"../Error"
	"../nmea"
	"time"
)

type stopFunc func() error

type connectionData struct {
	//
	DeviceID     uint32
	nmeaChan     chan<- *nmea.Data
	intervalInMs uint

	Stop stopFunc

	errorChan chan<- *Error.Error
}

const (
	SENSORINTERVALLINMSEC    = time.Millisecond * 100
	DATAPOINTTIMESPAWNINMSEC = time.Millisecond * 1000
)
