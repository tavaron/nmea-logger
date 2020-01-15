package nmea2mongo

import (
	"context"
	//"errors"
	//"go.mongodb.org/mongo-driver
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"../Error"
	"../nmea"
)


type DbConfig struct {
	username string
	password string
	uri string
	database string
}

type connectionData struct {
	config DbConfig
	clientOpts *options.ClientOptions
	client *mongo.Client
	database *mongo.Database
	errChan chan<- Error.Error
}

func InitMongo(ch <-chan nmea.Data, errCh chan<- Error.Error) {
	config := DbConfig{
		username:       "",
		password:       "",
		uri:            "mongodb://boatpi:27017",
		database:		"NMEA0183",
	}

	conn := &connectionData{
		config:		config,
		clientOpts:	options.Client().ApplyURI(config.uri),
		client:		nil,
		database:	nil,
		errChan:	errCh,
	}


	client, err := mongo.NewClient(conn.clientOpts)
	if err != nil {
		conn.errChan<-Error.Err(Error.High, err)
		return
	}
	conn.client = client
	err = conn.client.Connect(context.Background())
	if err != nil {
		conn.errChan<-Error.Err(Error.High, err)
		return
	}
	//defer conn.client.Disconnect(context.Background())
	err = conn.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		conn.errChan<-Error.Err(Error.High, err)
		return
	}

	conn.database = conn.client.Database(config.database)
	conn.errChan<-Error.New(Error.Info, "mongodb connected")
	go writeMongo(ch, conn)

}

func writeMongo(nmeaCh <-chan nmea.Data, conn *connectionData) {
	defer conn.client.Disconnect(context.Background())
	for data := range nmeaCh {
		switch data.Type {
		case "GPRMC":
			go mongoRMC(data, conn)
		case "--PAD":
			go mongoPAD(data, conn)
		default:
			print("unrecognized sentence: " + data.Sentence + "\n")
		}
	}
}


func checkForTimestamp(unixTime int64, coll *mongo.Collection) (bool, error) {
	filter := bson.D{{ "_id", unixTime}}
	coll.FindOne(context.TODO(), filter)
	return false, nil
}


// TODO check if device already has an entry on this timestamp
func checkForDeviceEntry(unixTime int64, devID uint32, coll *mongo.Collection) bool {
	filter := bson.M{ "_id": unixTime+10,
		"data": bson.D{{ "DeviceID", devID}},
	}
	result := coll.FindOne(context.TODO(),filter)
	if result != nil {
		return true
	}
	return false
}

func createTimestamp(unixTime int64, coll *mongo.Collection) (bool, error) {
	b, err := checkForTimestamp(unixTime, coll)
	if err != nil {
		return false, err
	}
	if (b) {
		return true, nil
	}
	_, err = coll.InsertOne(context.TODO(), bson.D{{"_id", unixTime},})
	if err != nil {
		return false, err
	}
	return true, nil
}