package sensors

import (
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices/bmxx80"

	"strconv"
	"time"

	"../Error"
	"../nmea"
)

func i2cInitBmxx80(data *i2cConnectionData) error {

	// define filter settings and oversampling
	devOpts := bmxx80.DefaultOpts
	devOpts.Filter = bmxx80.F16
	devOpts.Temperature = bmxx80.O16x
	devOpts.Pressure = bmxx80.O16x
	devOpts.Humidity = bmxx80.O16x

	dev, err := bmxx80.NewI2C(data.bus, data.config.primaryAddress, &devOpts)
	if err != nil {
		return err
	}
	data.errorChan <- Error.New(Error.Debug, "bmxx80 i2c device is open")
	data.Stop = dev.Halt

	envCh, err := dev.SenseContinuous(time.Second)
	if err != nil {
		return err
	}
	data.errorChan <- Error.New(Error.Debug, "Bmxx80 is in continuous sense mode")
	go i2cToNmeaBmxx80(data, envCh, 0x76)
	return nil
}

func i2cToNmeaBmxx80(data *i2cConnectionData, envCh <-chan physic.Env, id uint32) {
	//println("channel to Bmxx80 is been watched")
	for {

		select {
		case env, chOpen := <-envCh:
			if chOpen != true {
				data.errorChan <- Error.New(Error.Low, "channel to Bmxx80 was closed")
				return
			}
			//println("received data on channel to Bmxx80")
			temp := int(env.Temperature / (physic.MilliKelvin)) // milli kelvin
			humi := int(env.Humidity / physic.MilliRH)          // tenth percent rH
			pres := int(env.Pressure / (physic.Pascal))         // pascals
			nmeaSentence := "$--PAD,"
			nmeaSentence += strconv.Itoa(temp) + ","
			nmeaSentence += strconv.Itoa(humi) + ","
			nmeaSentence += strconv.Itoa(pres) + ",*PP" // TODO implement NMEA checksum

			nmeaData, err := nmea.NewData(nmeaSentence, data.DeviceID)
			if err != nil {
				data.errorChan <- Error.Err(Error.Low, err)
			} else {
				data.nmeaChan <- nmeaData
			}

		}
	}
}
