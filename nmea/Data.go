package nmea

type Data struct {
	// 0 <= devID < uint16max 			i2c devices
	// uint16max < devID < uint16max*2 	serial / nmea devices
	// uint16max*2 < devID				others
	DeviceID uint32
	Timestamp int64
	Type string
	Sentence string
}



