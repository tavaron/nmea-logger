package sensors

import (
	"errors"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices/bmxx80"

	"strconv"
	"time"

	"../Error"
	"../nmea"
	"./config"
)

func readBmxx80(conn *I2CConnection) error {

	cfg, err := config.I2CFromInterface(conn.config)
	if err != nil {
		return err
	}

	// define filter settings and oversampling
	devOpts := bmxx80.Opts{
		Temperature: bmxx80.O16x,
		Pressure:    bmxx80.O16x,
		Humidity:    bmxx80.O16x,
		Filter:      bmxx80.F16,
	}

	dev, err := bmxx80.NewI2C(conn.bus, cfg.Address(), &devOpts)
	if err != nil {
		return err
	}
	conn.error(errors.New("devBmxx80 was opened"))
	conn.stop = dev.Halt

	envCh, err := dev.SenseContinuous(time.Second)
	if err != nil {
		return err
	}
	conn.error(errors.New("devBmxx80 is in continuous sense mode"))
	go func() {
		for !conn.IsStopped() {

			select {
			case env, chOpen := <-envCh:
				if chOpen != true {
					conn.error(errors.New("channel to devBmxx80 was closed"), Error.Low)
					return
				}

				temp := int(env.Temperature / (physic.MilliKelvin)) // milli kelvin
				humi := int(env.Humidity / physic.MilliRH)          // tenth percent rH
				pres := int(env.Pressure / (physic.Pascal))         // pascals
				nmeaSentence := "$--PAD,"
				nmeaSentence += strconv.Itoa(temp) + ","
				nmeaSentence += strconv.Itoa(humi) + ","
				nmeaSentence += strconv.Itoa(pres) + ",*PP" //TODO implement NMEA checksum

				nmeaData, err := nmea.NewData(nmeaSentence, cfg.DeviceID())
				if err != nil {
					conn.error(err, Error.Low)
				} else {
					conn.send(nmeaData)
				}

			}
		}
	}()
	return nil
}
