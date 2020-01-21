package nmea

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	ZeroCelsiusInKelvin float64 = 273.15
)

type DataMap map[string]float64
type Data struct {
	// 0 <= devID < uint16max 			i2c devices
	// uint16max < devID < uint16max*2 	serial / nmea devices
	// uint16max*2 < devID				others
	Timestamp int64
	Type      string
	Data      DataMap `bson:"data"`
}

func NewData(sentence string, deviceID uint32) (*Data, error) {
	var d Data
	d.Timestamp = time.Now().Unix()
	d.Data = make(DataMap)
	d.Data["deviceid"] = float64(deviceID)
	buffer := strings.Split(sentence, ",")

	var err error
	switch buffer[0] {
	case "$--PAD":
		err = d.FromPADString(buffer)
	case "$GPRMC", "$--RMC":
		err = d.FromRMCString(buffer)
	default:
		err = d.FromRAWString(buffer)
	}

	return &d, err
}

func Average(data []DataMap) (*Data, error) {
	var result Data
	amounts := make(map[string]int)
	for _, currentMap := range data {
		for key, value := range currentMap {
			result.Data[key] += value
			amounts[key] += 1
		}
	}
	for key := range result.Data {
		result.Data[key] /= float64(amounts[key])
	}
	return &result, nil
}

func (d *Data) DeviceID() uint32 {
	return uint32(d.Data["deviceid"])
}

func (d *Data) FromRMCString(buffer []string) error {
	d.Type = "MALFORMED"

	// check for proper amount of substrings
	if len(buffer) != 13 {
		return errors.New("malformed RMC sentence received! length: " + strconv.Itoa(len(buffer)))
	}

	// check if really a RMC sentence
	if buffer[0] != "$GPRMC" && buffer[0] != "$--RMC" {
		return errors.New("invalid RMC sentence received")
	}

	// check for gps fix
	if buffer[2] != "A" {
		return errors.New("gps fix not established")
	}

	// read latitude
	lati, err := strconv.ParseFloat(buffer[3], 64)
	if err != nil {
		return errors.New("could not parse latitude from rmc sentence")
	}
	if buffer[4] == "S" {
		lati *= -1
	} else if buffer[4] != "N" {
		return errors.New("invalid latitude heading in rmc sentence")
	}

	// read longitude
	long, err := strconv.ParseFloat(buffer[5], 64)
	if err != nil {
		return errors.New("could not parse latitude from rmc sentence")
	}
	if buffer[6] == "W" {
		long *= -1
	} else if buffer[6] != "E" {
		return errors.New("invalid longitude heading in rmc sentence")
	}

	// check if speed was given and read it
	var speed float64
	if buffer[7] != "" {
		speed, err = strconv.ParseFloat(buffer[7], 64)
	}
	if err != nil || buffer[7] == "" {
		speed = -1.0
	}
	err = nil

	// check if true course was given and read it
	var tc float64
	if buffer[8] != "" {
		tc, err = strconv.ParseFloat(buffer[8], 64)
	}
	if err != nil || buffer[8] == "" {
		tc = -1.0
	}
	err = nil

	// check if magnetic variation was given and read it
	var mv float64
	if buffer[8] != "" {
		mv, err = strconv.ParseFloat(buffer[8], 64)
	}
	if err != nil || buffer[8] == "" {
		mv = 1024.0
	}
	err = nil

	d.Type = "RMC"
	d.Data["latitude"] = lati
	d.Data["longitude"] = long
	d.Data["speed"] = speed
	d.Data["truecourse"] = tc
	d.Data["magneticvariation"] = mv
	return nil
}

func (d *Data) FromPADString(buffer []string) error {
	d.Type = "MALFORMED"

	// check for proper amount of substrings
	if len(buffer) != 5 {
		return errors.New("malformed sentence received! length: " + strconv.Itoa(len(buffer)))
	}

	// check if really a PAD sentence
	if buffer[0] != "$--PAD" {
		return errors.New("invalid PAD sentence received")
	}

	temp, err := strconv.Atoi(buffer[1])
	if err != nil {
		return errors.New("could not parse temperature from pad sentence")
	}

	humi, err := strconv.Atoi(buffer[2])
	if err != nil {
		return errors.New("could not parse temperature from pad sentence")
	}

	pres, err := strconv.Atoi(buffer[3])
	if err != nil {
		return errors.New("could not parse temperature from pad sentence")
	}

	d.Type = "PAD"
	d.Data["temperature"] = (float64(temp) / 1000.0) - ZeroCelsiusInKelvin
	d.Data["humidity"] = float64(humi) / 10.0
	d.Data["pressure"] = float64(pres) / 100.0
	return nil
}

func (d *Data) FromRAWString(buffer []string) error {
	d.Type = "MALFORMED"
	if len(buffer) > 0 {
		if len(buffer[0]) == 6 {
			d.Type = "RAW" + buffer[0][1:5]
		} else {
			d.Type = "RAWUNKNOWN"
		}
	} else {
		return errors.New("received empty NMEA string")
	}
	for i, value := range buffer {
		if len(value) > 0 {
			temporary, err := strconv.ParseFloat(value, 64)
			if err == nil {
				d.Data[strconv.Itoa(i)] = temporary
			}
		}
	}
	return nil
}
