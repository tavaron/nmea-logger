package nmea2mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"../Error"
	"../nmea"
)

func mongoPAD(data nmea.Data, conn *connectionData) {
	pad, err := nmea.NewPAD(data)
	if err != nil {
		conn.errChan<-Error.Err(Error.Low, err)
		return
	}

	err = conn.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		conn.errChan<-Error.Err(Error.Low, err)
		return
	}

	coll := conn.database.Collection("PAD")
	b, err := createTimestamp(data.Timestamp, coll)
	if err != nil {
		conn.errChan<-Error.Err(Error.Low, err)
		return
	}
	if !b {
		conn.errChan<-Error.New(Error.Low, "could not create PAD timestamp")
		return
	}

	/*if checkForDeviceEntry(data.Timestamp, data.DeviceID, coll) {
		conn.errChan <- Error.New(Error.Debug, "dismissed PAD sentence")
		return
	}*/

	filter := bson.D{{"_id", data.Timestamp}}
	update := bson.D{{ "$push", bson.D{{"data", pad}} }}
	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		conn.errChan<-Error.New(Error.Low,"error while writing PAD data to mongodb")
		return
	}
}