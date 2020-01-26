package nmea2mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strconv"

	"../Error"
)

func (run *Engine) pingAsError() *Error.Error {

	err := run.client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		return Error.Err(Error.Low, err, mongoFlag)
	}
	return nil
}

func (run *Engine) timestampExists(time int64, collection string) bool {
	return run.deviceEntryExists(time, -1, collection)
}

func (run *Engine) deviceEntryExists(time int64, deviceID int64, collection string) bool {

	// assign and check collection
	if !run.collectionExists(collection) {
		return false
	}
	coll := run.database.Collection(collection)
	if coll == nil {
		run.errorChan <- Error.New(Error.Debug,
			"deviceEntryExists() failed due to nil pointer collection",
			mongoFlag)
		return false
	}

	filter := bson.D{{"_id", time}}

	// search for timestamp
	var result *Result = nil
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return false
	} else if err != nil {
		run.errorChan <- Error.Err(Error.Debug, err, mongoFlag)
		return true
	}

	// return if only timestamp should be checked
	if deviceID < 0 {
		return true
	}

	run.errorChan <- Error.New(Error.Debug,
		"checking for device: "+strconv.FormatInt(deviceID, 10)+"\n",
		mongoFlag)

	for _, value := range result.Devices {
		if value == deviceID {
			return true
		}
	}

	return false

}

func (run *Engine) readLastSecond(collection string, interval int64, readFirst ...bool) int64 {
	coll := run.database.Collection(collection)
	opts := options.FindOne()
	if len(readFirst) == 0 || !readFirst[0] {
		opts.SetSort(bson.M{"_id": -1})
	}

	var result *Result
	err := coll.FindOne(context.TODO(), bson.M{}, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		run.errorChan <- Error.Err(Error.Info, err, mongoFlag)
		return 0
	}
	if err != nil {
		run.errorChan <- Error.Err(Error.Low, err, mongoFlag)
		return -1
	}

	if mod := result.Id % interval; mod != 0 {
		return result.Id - mod - interval
	} else {
		return result.Id
	}
}

func (run *Engine) collectionExists(collection string) bool {
	exists := false
	collectionList, _ := run.database.ListCollectionNames(context.TODO(), bson.M{})
	for _, subStr := range collectionList {
		if subStr == collection {
			exists = true
			break
		}
	}
	return exists
}
