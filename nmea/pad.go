package nmea

import (
	"errors"
	"strconv"
	"strings"
)

const ZeroCelsiusInKelvin = float64(273.15)

// proprietary air data
type PAD struct {
	// device id
	DeviceID uint32

	// degree Celsius
	Temperature float64

	// percent RH
	Humidity float64

	// hecto Pascals
	Pressure float64
}

func NewPAD(nmea Data) (*PAD, error) {
	buffer := strings.Split(nmea.Sentence, ",")

	// check for proper amount of substrings
	if len(buffer) != 5 {
		return nil, errors.New("malformed sentence received! length: " + strconv.Itoa(len(buffer)))
	}

	// check if really a PAD sentence
	if buffer[0] != "$--PAD" {
		return nil, errors.New("invalid PAD sentence received: " + nmea.Sentence)
	}

	pad := PAD{
		DeviceID:    nmea.DeviceID,
		Temperature: 0,
		Humidity:    0,
		Pressure:    0,
	}

	temp, err := strconv.Atoi(buffer[1])
	if err != nil {
		return nil, errors.New("could not parse temperature from pad sentence")
	}
	pad.Temperature = (float64(temp) / 1000.0) - ZeroCelsiusInKelvin

	humi, err := strconv.Atoi(buffer[2])
	if err != nil {
		return nil, errors.New("could not parse temperature from pad sentence")
	}
	pad.Humidity = float64(humi) / 10.0

	pres, err := strconv.Atoi(buffer[3])
	if err != nil {
		return nil, errors.New("could not parse temperature from pad sentence")
	}
	pad.Pressure = float64(pres) / 100.0

	return &pad, nil
}