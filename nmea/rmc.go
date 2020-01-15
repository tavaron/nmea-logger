package nmea

import (
	"errors"
	"strconv"
	"strings"
)

type RMC struct {
	// device id
	DeviceID uint32


	// North = positive
	// South = negative
	Latitude float64

	// East = positive
	// West = negative
	Longitude float64

	// speed in knots, negative if n/a
	Speed float64

	// true course in degrees, negative if n/a
	TrueCourse float64

	// East = positive
	// West = negative
	MagneticVariation float64
}



func NewRMC(nmea Data) (*RMC, error) {
	buffer := strings.Split(nmea.Sentence, ",")

	// check for proper amount of substrings
	if len(buffer) != 13 {
		return nil, errors.New("malformed RMC sentence received! length: " + strconv.Itoa(len(buffer)))
	}


	// check if really a RMC sentence
	if buffer[0] != "$GPRMC" {
		return nil, errors.New("invalid RMC sentence received: " + nmea.Sentence)
	}

	// check for gps fix
	if buffer[2] != "A" {
		return nil, errors.New("gps fix not established")
	}




	// read latitude
	lati , err:= strconv.ParseFloat(buffer[3], 64)
	if err != nil {
		return nil, errors.New("could not parse latitude from rmc sentence")
	}
	if buffer[4] == "S" {
		lati *= -1
	} else if buffer[4] != "N" {
		return nil, errors.New("invalid latitude heading in rmc sentence")
	}


	// read longitude
	long , err:= strconv.ParseFloat(buffer[5], 64)
	if err != nil {
		return nil, errors.New("could not parse latitude from rmc sentence")
	}
	if buffer[6] == "W" {
		long *= -1
	} else if buffer[6] != "E" {
		return nil, errors.New("invalid longitude heading in rmc sentence")
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

	rmc := RMC {
		DeviceID:		   nmea.DeviceID,
		Latitude:          lati,
		Longitude:         long,
		Speed:             speed,	// negative if not given
		TrueCourse:        tc,		// negative if not given
		MagneticVariation: mv,		// 1024 if not given
	}


	return &rmc, nil
}



func (rmc *RMC) ToString() string {
	// TODO implement Stringer
	return ""
}