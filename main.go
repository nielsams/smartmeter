package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/roaldnefs/go-dsmr"
	"github.com/tarm/serial"
)

var (
	config         *serial.Config
	currentMetrics smartMeterData
)

type smartMeterData struct {
	Timestamp               string  `json:"timestamp"`
	CurrentPowerConsumption float64 `json:"currentPowerConsumption"`
	InstVoltL1              float64 `json:"instVoltL1"`
	InstVoltL2              float64 `json:"instVoltL2"`
	InstVoltL3              float64 `json:"instVoltL3"`
	InstCurrentL1           float64 `json:"instCurrentL1"`
	InstCurrentL2           float64 `json:"instCurrentL2"`
	InstCurrentL3           float64 `json:"instCurrentL3"`
	GasDelivered            float64 `json:"gasDelivered"`
	PowerDeliveredTariff1   float64 `json:"powerDeliveredTariff1"`
	PowerDeliveredTariff2   float64 `json:"powerDeliveredTariff2"`
}

func main() {
	serialPort := os.Getenv("USBDEVICE")
	config = &serial.Config{
		Name: serialPort,
		Baud: 115200,
	}

	fmt.Printf("Starting with usb device %s\n", serialPort)

	readMetrics()
	http.HandleFunc("/data", handleRequest)
	http.ListenAndServe(":8080", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentMetrics)
}

func readMetrics() {

	go func() {
		stream, err := serial.OpenPort(config)
		if err != nil {
			fmt.Println(err)
		}

		reader := bufio.NewReader(stream)

		for {
			dt := time.Now()
			currentMetrics.Timestamp = dt.Format(time.RFC3339)

			// Peek at the next byte, and look for the start of the telegram
			if peek, err := reader.Peek(1); err == nil {
				// The telegram starts with a '/' character keep reading
				// bytes until the start of the telegram is found
				if string(peek) != "/" {
					reader.ReadByte()
					continue
				}
			} else {
				continue
			}

			// Keep reading until the '!' character which indicates the end of
			// the telegram and is followed by the CRC
			rawTelegram, err := reader.ReadBytes('!')
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Read the CRC which can be used to detect faulty telegram
			// TODO check CRC
			_, err = reader.ReadBytes('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}

			telegram, err := dsmr.ParseTelegram(string(rawTelegram))
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Test for different telegram types and update global metrics object

			if rawValue, ok := telegram.InstantaneousVoltageL1(); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.InstVoltL1 = value
			}

			if rawValue, ok := telegram.InstantaneousVoltageL2(); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.InstVoltL2 = value
			}

			if rawValue, ok := telegram.InstantaneousVoltageL3(); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.InstVoltL3 = value
			}

			if rawValue, ok := telegram.InstantaneousCurrentL1(); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.InstCurrentL1 = value
			}

			if rawValue, ok := telegram.InstantaneousCurrentL2(); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.InstCurrentL2 = value
			}

			if rawValue, ok := telegram.InstantaneousCurrentL3(); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.InstCurrentL3 = value
			}

			if rawValue, ok := telegram.MeterReadingElectricityDeliveredToClientTariff1(); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.PowerDeliveredTariff1 = value
			}

			if rawValue, ok := telegram.MeterReadingElectricityDeliveredToClientTariff2(); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.PowerDeliveredTariff2 = value
			}

			if rawValue, ok := telegram.MeterReadingGasDeliveredToClient(1); ok {
				value, err := strconv.ParseFloat(rawValue, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.GasDelivered = value
			}

			i, ok := telegram.DataObjects["1-0:1.7.0"]
			if ok {
				value, err := strconv.ParseFloat(i.Value, 64)
				if err != nil {
					fmt.Println(err)
					continue
				}
				currentMetrics.CurrentPowerConsumption = value
			}
		}
	}()

}
