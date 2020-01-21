package nmea2mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"../Error"
	"../nmea"
)

const (
	minute int64 = 60
	hour   int64 = 3600
	day    int64 = 86400
)

type DbConfig struct {
	username string
	password string
	uri      string
	database string
}

type Engine struct {
	config     DbConfig
	clientOpts *options.ClientOptions
	client     *mongo.Client
	database   *mongo.Database
	errorChan  chan<- *Error.Error
	dataChan   <-chan *nmea.Data
}

type Result struct {
	Id      int64          `bson:"_id"`
	Devices []uint32       `bson:"devices"`
	Data    []nmea.DataMap `bson:"data"`
}

func Run(ch <-chan *nmea.Data, errCh chan<- *Error.Error) {
	config := DbConfig{
		username: "",
		password: "",
		uri:      "mongodb://boatpi:27017",
		database: "NMEA0183",
	}

	conn := &Engine{
		config:     config,
		clientOpts: options.Client().ApplyURI(config.uri),
		client:     nil,
		database:   nil,
		errorChan:  errCh,
		dataChan:   ch,
	}

	client, err := mongo.NewClient(conn.clientOpts)
	if err != nil {
		conn.errorChan <- Error.Err(Error.High, err)
		return
	}
	conn.client = client
	err = conn.client.Connect(context.Background())
	if err != nil {
		conn.errorChan <- Error.Err(Error.High, err)
		return
	}
	//defer conn.client.Disconnect(context.Background())
	if !conn.Ping() {
		return
	}

	conn.database = conn.client.Database(config.database)
	conn.errorChan <- Error.New(Error.Debug, "mongodb connected")
	go conn.routine()

}

func (run *Engine) Ping() bool {
	err := run.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		run.errorChan <- Error.Err(Error.High, err)
		return false
	}
	return true
}

func (run *Engine) routine() {
	defer run.client.Disconnect(context.Background())
	for data := range run.dataChan {
		go run.write(data, data.Type)
	}
}

//TODO fix checks
func (run *Engine) checkForTimestamp(time int64, collection string) (bool, *Error.Error) {
	return run.checkForDeviceEntry(time, 0, collection)
}
func (run *Engine) checkForDeviceEntry(time int64, deviceID uint32, collection string) (bool, *Error.Error) {

	// check connectivity
	Err := run.checkAsError()
	if Err != nil {
		return false, Err
	}

	// assign and check collection
	coll := run.database.Collection(collection)
	if coll == nil {
		return false, Error.New(Error.Low, "nmea2mongo.checkForDeviceEntry failed due to nil pointer collection")
	}

	// differ between timestamp-only and device-entry
	var filter bson.M
	if deviceID <= 0 {
		filter = bson.M{
			"_id":     time,
			"devices": deviceID,
		}
	} else {
		filter = bson.M{
			"_id": time,
		}
	}

	// search for timestamp
	var result *Result = nil
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if result == nil && err != nil && err.Error() == "mongo: no documents in result" {
		return false, nil
	} else if err != nil {
		return false, Error.Err(Error.Debug, err)
	}

	return true, nil
}

func (run *Engine) checkAsError() *Error.Error {

	err := run.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return Error.Err(Error.Low, err)
	}
	return nil
}

func (run *Engine) createTimestamp(time int64, collection string) (bool, *Error.Error) {
	coll := run.database.Collection(collection)
	_, err := coll.InsertOne(context.TODO(), bson.M{"_id": time})
	if err != nil {
		return false, Error.Err(Error.Low, err)
	}
	return true, nil
}

func (run *Engine) getLastSecond(collection string, interval int64) (int64, *Error.Error) {
	coll := run.database.Collection(collection)
	opts := options.FindOne()
	opts.SetSort(bson.M{"_id": -1})

	var result *Result
	err := coll.FindOne(context.TODO(), bson.D{}, opts).Decode(&result)
	if err != nil {
		return 0, Error.Err(Error.Low, err)
	} else if result == nil {
		return 0, Error.New(Error.Debug, "getLastSecond quits without result")
	}

	if mod := result.Id & interval; mod != 0 {
		return result.Id - mod - interval, nil
	} else {
		return result.Id + 1 - interval, nil
	}
}

func (run *Engine) write(data *nmea.Data, collection string) {

	// check if entry already exists
	b, Err := run.checkForTimestamp(data.Timestamp, collection)
	if b {
		b, Err = run.checkForDeviceEntry(data.Timestamp, data.DeviceID(), collection)
		if b {
			return
		}
	} else {
		run.createTimestamp(data.Timestamp, collection)
	}
	if Err != nil {
		run.errorChan <- Err
		return
	}

	coll := run.database.Collection(collection)
	filter := bson.M{"_id": data.Timestamp}
	update := bson.M{"$push": bson.M{
		"data":    data.Data,
		"devices": data.DeviceID(),
	}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		run.errorChan <- Error.Err(Error.Low, err)
		return
	}

}

func (run *Engine) average(nmeaType string, interval int64) *Error.Error {
	collSeconds := run.database.Collection(nmeaType)
	var intervalStr string
	switch interval {
	case minute:
		intervalStr = "minutes"
	case hour:
		intervalStr = "hours"
	case day:
		intervalStr = "days"
	default:
		intervalStr = "customInterval"
	}

	start, err := run.getLastSecond(nmeaType, interval)
	if err != nil {
		return err
	}
	intervalID := start / interval

	run.checkForTimestamp(intervalID, nmeaType+intervalStr)

	filter := bson.M{"_id": bson.M{
		"$ge": start,
		"$lt": start + interval},
	}
	cursor, err2 := collSeconds.Find(context.TODO(), filter, options.Find())
	if err2 != nil {
		return Error.Err(Error.Low, err2)
	}
	defer cursor.Close(context.TODO())

	var list []nmea.DataMap
	for cursor.Next(context.TODO()) {
		var current *Result
		err2 = cursor.Decode(&current)
		if err2 != nil {
			run.errorChan <- Error.Err(Error.Low, err2)
		} else {
			for _, value := range current.Data {
				list = append(list, value)
			}
		}
	}
	if len(list) == 0 {
		return Error.New(Error.Low, "no data to calculate average")
	}
	average, err2 := nmea.Average(list)
	if err2 != nil {
		return Error.Err(Error.Low, err2)
	}

	run.write(average, nmeaType+intervalStr)

	return nil
}
