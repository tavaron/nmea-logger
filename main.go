package main

import (
	"./Error"
	"./database"
	"./nmea"
	"./sensors"
	sensorCfg "./sensors/config"
)

//TODO config files
type mainConfig struct {
	I2cBus           string ""
	SensorConfigPath string ""
	toMongo          bool
	Db               nmea2mongo.DbConfig
	toSerial         bool
	serialPath       string
	toConsole        bool
}

type ChannelList struct {
	Error      chan *Error.Error
	In         chan *nmea.Data
	MongoDb    chan *nmea.Data
	SerialPort chan *nmea.Data

	StopI2c       chan bool
	StopSerialIn  chan bool
	StopSerialOut chan bool
	StopMongoDb   chan bool
	StopConsole   chan bool
}

func main() {
	channels := &ChannelList{
		Error:         make(chan *Error.Error, 128),
		In:            make(chan *nmea.Data),
		MongoDb:       make(chan *nmea.Data, 1024),
		SerialPort:    nil, //make(chan nmea.Raw, 1024),
		StopI2c:       make(chan bool, 1),
		StopSerialIn:  make(chan bool, 1),
		StopSerialOut: nil, //make(chan bool, 1),
		StopMongoDb:   make(chan bool, 1),
		StopConsole:   make(chan bool, 1),
	}

	go nmeaDispatcher(channels)

	mongoDb := nmea2mongo.New(channels.MongoDb, channels.Error)
	mongoDb.RecalculateAverage()
	mongoDb.Run(false)

	sensorEng := sensors.NewEngine(channels.In, channels.Error)
	configGPS := sensorCfg.DefaultSerial()
	configBmxx80 := sensorCfg.DefaultI2C()
	sensorEng.Connect(configGPS)
	sensorEng.Connect(configBmxx80)

	for err := range channels.Error {
		switch err.Lvl {
		case Error.Debug:
			println("[DEBUG] " + err.Text)
		case Error.Info:
			println("[INFO]  " + err.Text)
		case Error.Warning:
			println("[WARN]  " + err.Text)
		case Error.Low:
			println("[LOW]   " + err.Text)
		case Error.High:
			println("[HIGH]  " + err.Text)
		case Error.Fatal:
			println("[FATAL] " + err.Text)
		default:
			println("[UNKWN] " + err.Text)
		}
	}
}

func nmeaDispatcher(channels *ChannelList) {
	for data := range channels.In {
		channels.MongoDb <- data
	}
}
