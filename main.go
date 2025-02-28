package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"tinygo.org/x/bluetooth"
	"github.com/influxdata/influxdb-client-go/v2"
)

var adapter = bluetooth.DefaultAdapter
var knownDevices map[string]string
var ch chan bluetooth.ScanResult

var rDB *redis.Client
var influxDB 
var ctx context.Context

var influxToken := "loJwIREFDqGQcO6ummRk5lRkdbHA1oulukQCeT0oM1wF53AeI50JCYV10LTXgD4mlzIymoFAaASYNJM0Mkl6GA=="

func main() {
	ctx = context.Background()
	initDatabases()
	initBluetooth()

	knownDevices = make(map[string]string)
	ch = make(chan bluetooth.ScanResult, 1)
	for {

		// Start scanning.
		println("scanning...")
		err := adapter.Scan(processScannedDevice)
		must("start scan", err)
		var device bluetooth.Device
		select {
		case result := <-ch:
			println("Storing device address and name")

			deviceKey := fmt.Sprintf("gotooth:%s", result.Address.String())
			err := rDB.Set(ctx, deviceKey, result.LocalName(), 0).Err()
			if err != nil {
				panic(err)
			}
			//knownDevices[result.Address.String()] = result.LocalName()
			device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
			if err != nil {
				println(err.Error())
				continue
			}
			println("connected to ", result.Address.String())
			//discoverDevice(device)
			err = device.Disconnect()
			if err != nil {
				println(err)
			}
		}
	}

}

func initBluetooth() {
	must("enable BLE stack", adapter.Enable())
}

func initDatabases() {
	rDB = redis.NewClient(&redis.Options{
		Addr:     "redis.local:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Create a new client using an InfluxDB server base URL and an authentication token
	influxDB = influxdb2.NewClient("http://influx.local:8086", influxToken)
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

// DeviceAddress is the MAC address of the Bluetooth peripheral you want to connect to.
// Replace this by using -ldflags="-X main.DeviceAddress=[MAC ADDRESS]"
// where [MAC ADDRESS] is the actual MAC address of the peripheral.
// For example:
// tinygo flash -target circuitplay-bluefruit -ldflags="-X main.DeviceAddress=7B:36:98:8C:41:1C" ./examples/discover/
var DeviceAddress string

func connectAddress() string {
	return DeviceAddress
}

// wait on baremetal, proceed immediately on desktop OS.
func wait() {
	time.Sleep(3 * time.Second)
}

func processScannedDevice(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	var exists bool
	deviceKey := fmt.Sprintf("gotooth:%s", device.Address.String())
	_, err := rDB.Get(ctx, deviceKey).Result()

	if err == redis.Nil {
		exists = false
	} else if err != nil {
		panic(err)
	} else {
		exists = true
	}
	if !exists {
		println("found device:", device.Address.String(), device.RSSI, device.LocalName(), device.ManufacturerData())
		adapter.StopScan()
		ch <- device
	} else {
		println("known device:", device.Address.String(), device.RSSI, device.LocalName())
	}

}

func discoverDevice(device bluetooth.Device) {
	// get services
	println("discovering services/characteristics")
	srvcs, err := device.DiscoverServices(nil)
	must("discover services", err)

	// buffer to retrieve characteristic data
	buf := make([]byte, 255)

	for _, srvc := range srvcs {
		println("- service", srvc.UUID().String())

		chars, err := srvc.DiscoverCharacteristics(nil)
		if err != nil {
			println(err)
		}
		for _, char := range chars {
			println("-- characteristic", char.UUID().String())
			mtu, err := char.GetMTU()
			if err != nil {
				println("    mtu: error:", err.Error())
			} else {
				println("    mtu:", mtu)
			}
			n, err := char.Read(buf)
			if err != nil {
				println("    ", err.Error())
			} else {
				println("    data bytes", strconv.Itoa(n))
				println("    value =", string(buf[:n]))
			}
		}
	}

	err = device.Disconnect()
	if err != nil {
		println(err)
	}

}
