package nmea2mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"../Error"
	"../nmea"
)

const (
	minute    int64  = 60
	hour      int64  = 3600
	day       int64  = 86400
	mongoFlag string = "[mongodb]"
)

type DbConfig struct {
	username string
	password string
	uri      string
	database string
}

type Engine struct {
	config               DbConfig
	clientOpts           *options.ClientOptions
	client               *mongo.Client
	database             *mongo.Database
	errorChan            chan<- *Error.Error
	dataChan             <-chan *nmea.Data
	averageStopIndicator bool

	collections map[string]*mongo.Collection
}

type Result struct {
	Id      int64          `bson:"_id"`
	Devices []int64        `bson:"devices"`
	Data    []nmea.DataMap `bson:"data"`
}

func interval2string(interval int64) string {
	switch interval {
	case minute:
		return "minutes"
	case hour:
		return "hours"
	case day:
		return "days"
	default:
		return "customInterval"
	}
}
