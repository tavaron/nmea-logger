package nmea2mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"../Error"
	"../nmea"
)

type resultRMC struct {
	_id int64
	data []nmea.RMC
}

func mongoRMC(data nmea.Data, conn *connectionData) {
	rmc, err := nmea.NewRMC(data)
	if err != nil {
		conn.errChan<-Error.Err(Error.Low, err)
		return
	}

	err = conn.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		conn.errChan<-Error.Err(Error.Low, err)
		return
	}

	coll := conn.database.Collection("RMC")
	b, err := createTimestamp(data.Timestamp, coll)
	if err != nil {
		conn.errChan<-Error.Err(Error.Low, err)
		return
	}
	if !b {
		conn.errChan<-Error.New(Error.Low,"could not create RMC timestamp")
		return
	}

	/*if checkForDeviceEntry(data.Timestamp, data.DeviceID, coll) {
		conn.errChan <- Error.New(Error.Debug, "dismissed RMC sentence")
		return
	}*/

	filter := bson.D{{"_id", data.Timestamp}}
	update := bson.D{{ "$push", bson.D{{"data", rmc}} }}
	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		conn.errChan<-Error.Err(Error.Low, err)
		return
	}
}