package nmea2mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"../Error"
	"../nmea"
)

func (run *Engine) createTimestamp(time int64, collection string) bool {
	if run.timestampExists(time, collection) {
		return true
	}

	coll := run.database.Collection(collection)
	_, err := coll.InsertOne(context.TODO(), bson.M{"_id": time})
	if err != nil {
		run.errorChan <- Error.Err(Error.Low, err, mongoFlag)
		return false
	}
	return true
}

func (run *Engine) write(data *nmea.Data, collection string) {

	// check if entry already exists
	if run.deviceEntryExists(data.Timestamp, data.DeviceID(), collection) {
		return
	}
	run.createTimestamp(data.Timestamp, collection)

	coll := run.database.Collection(collection)
	filter := bson.M{"_id": data.Timestamp}
	update := bson.M{"$push": bson.M{
		"data":    data.Data,
		"devices": data.DeviceID(),
	}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		run.errorChan <- Error.Err(Error.Low, err, mongoFlag)
		return
	}

}

func (run *Engine) writeAverage(nmeaType string, start int64, interval int64) {
	timestamp := start / interval
	if run.timestampExists(timestamp, nmeaType+interval2string(interval)) {
		return
	}

	filter := bson.M{"_id": bson.M{
		"$gte": start,
		"$lt":  start + interval},
	}

	collSeconds := run.database.Collection(nmeaType)
	cursor, err2 := collSeconds.Find(context.TODO(), filter, options.Find())
	if err2 != nil {
		run.errorChan <- Error.Err(Error.Low, err2, mongoFlag)
		return
	}
	defer cursor.Close(context.Background())

	var list []nmea.DataMap
	for cursor.Next(context.TODO()) {
		var current *Result
		err2 = cursor.Decode(&current)
		if err2 != nil {
			run.errorChan <- Error.Err(Error.Low, err2, mongoFlag)
		} else {
			for _, value := range current.Data {
				list = append(list, value)
			}
		}
	}
	if len(list) == 0 {
		return
	}

	datamap, err2 := nmea.Average(list)
	if err2 != nil {
		run.errorChan <- Error.Err(Error.Info, err2, mongoFlag)
	}
	average := &nmea.Data{
		Timestamp: timestamp,
		Type:      nmeaType,
		Data:      *datamap,
	}

	run.write(average, nmeaType+interval2string(interval))

}
