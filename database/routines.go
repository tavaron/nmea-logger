package nmea2mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"math"
	"strconv"
	"strings"
	"time"

	"../Error"
	"../nmea"
)

func New(ch <-chan *nmea.Data, errCh chan<- *Error.Error) *Engine {
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

	if conn.errorChan == nil {
		return nil
	}
	if conn.dataChan == nil {
		conn.errorChan <- Error.New(Error.Fatal,
			"nil pointer: nmea data channel",
			mongoFlag)
	}

	client, err := mongo.NewClient(conn.clientOpts)
	if err != nil {
		conn.errorChan <- Error.Err(Error.High, err, mongoFlag)
		return nil
	}
	conn.client = client
	err = conn.client.Connect(context.Background())
	if err != nil {
		conn.errorChan <- Error.Err(Error.High, err, mongoFlag)
		return nil
	}

	if !conn.Ping() {
		return nil
	}

	conn.database = conn.client.Database(config.database)
	conn.errorChan <- Error.New(Error.Debug,
		"engine connected to "+config.uri,
		mongoFlag)

	return conn
}

func (run *Engine) Run(calculateAverages ...bool) bool {
	if !run.Ping() {
		return false
	}
	go run.dataRoutine()
	if len(calculateAverages) > 0 && calculateAverages[0] {
		go run.averageRoutine()
	}
	return false
}

//TODO include complete state validation
func (run *Engine) Ping() bool {
	err := run.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		run.errorChan <- Error.Err(Error.High, err, mongoFlag)
		return false
	}
	return true
}

//TODO
func (run *Engine) Stop() {
	err := run.client.Disconnect(context.Background())
	if err != nil {
		run.errorChan <- Error.Err(Error.Debug, err, mongoFlag)
	}

}

func (run *Engine) RecalculateAverage() {
	run.errorChan <- Error.New(Error.Debug,
		"recalculating averages: please wait",
		mongoFlag)

	if err := run.pingAsError(); err != nil {
		run.errorChan <- err
		return
	}

	collList, err := run.database.ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		run.errorChan <- Error.Err(Error.Debug, err, mongoFlag)
		return
	}

	for _, collName := range collList {
		if strings.Contains(collName, "minutes") ||
			strings.Contains(collName, "hours") ||
			strings.Contains(collName, "days") {
			err = run.database.Collection(collName).Drop(context.TODO())
			if err != nil {
				run.errorChan <- Error.Err(Error.Debug, err, mongoFlag)
			}
		}
	}

	go run.averageWorkDispatcher()
}

func (run *Engine) dataRoutine() {
	run.errorChan <- Error.New(Error.Debug,
		"data routine started",
		mongoFlag)

	for data := range run.dataChan {
		go run.write(data, data.Type)
	}
}

func (run *Engine) averageRoutine() {
	minuteTicker := time.NewTicker(time.Minute)
	defer minuteTicker.Stop()
	run.averageWorkDispatcher()

	for !run.averageStopIndicator {
		select {
		case <-minuteTicker.C:
			go run.averageWorkDispatcher()
		}
	}
}

func (run *Engine) averageWorkDispatcher() {
	if err := run.pingAsError(); err != nil {
		run.errorChan <- err
		return
	}

	nmeaTypes := make([]string, 0)
	collList, err := run.database.ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		run.errorChan <- Error.Err(Error.Debug, err, mongoFlag)
		return
	}

	for _, collName := range collList {
		if !strings.Contains(collName, "minutes") &&
			!strings.Contains(collName, "hours") &&
			!strings.Contains(collName, "days") {
			nmeaTypes = append(nmeaTypes, collName)
		}
	}

	for _, nmeaType := range nmeaTypes {
		run.averageWorker(nmeaType)
	}

}

func (run *Engine) averageWorker(nmeaType string) {

	startTime := time.Now()
	todo := make([]int64, 0)
	intervalList := []int64{60, 3600, 86400}
	for _, interval := range intervalList {
		// just do last entry if collection exists
		if run.collectionExists(nmeaType + interval2string(minute)) {
			run.writeAverage(
				nmeaType,
				run.readLastSecond(nmeaType, interval),
				interval)
		} else {
			todo = append(todo, interval)
		}
	}

	if len(todo) > 0 {
		start := int64(math.MaxInt64)
		end := int64(math.MinInt64)
		for _, interval := range intervalList {
			startTemp := run.readLastSecond(nmeaType, interval, true)
			endTemp := run.readLastSecond(nmeaType, interval)
			if startTemp < start {
				start = startTemp
			}
			if endTemp > end {
				end = endTemp
			}
		}

		run.errorChan <- Error.New(Error.Debug,
			"recalculating "+nmeaType+"\n"+
				"start: "+time.Unix(start, 0).String()+"\n"+
				"end:   "+time.Unix(end, 0).String(),
			mongoFlag)

		//TODO improve search
		for i := start; i <= end; i += minute {
			for _, interval := range todo {
				if i%interval == 0 {
					run.writeAverage(nmeaType, i, interval)
				}
			}
		}
	}

	run.errorChan <- Error.New(Error.Debug,
		"finished calculating "+nmeaType+
			" in "+strconv.FormatFloat(time.Since(startTime).Seconds(), 'f', 3, 64)+"s",
		mongoFlag)

}
