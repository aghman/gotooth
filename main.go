package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/redis/go-redis/v9"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter
var ch chan bluetooth.ScanResult

var rDB *redis.Client
var influxDB influxdb2.Client
var radioAPI api.WriteAPIBlocking
var ctx context.Context
var hostname string
var err error

func main() {
	ctx = context.Background()
	initDatabases()
	initBluetooth()

	hostname, err = os.Hostname()

	ch = make(chan bluetooth.ScanResult, 1)
	for {

		// Start scanning.
		println("scanning...")
		err = adapter.Scan(processScannedDevice)
		must("start scan", err)
		var device bluetooth.Device
		select {
		case result := <-ch:
			println("Storing device address and name")

			deviceKey := fmt.Sprintf("gotooth:%s", result.Address.String())
			err = rDB.Set(ctx, deviceKey, result.LocalName(), 0).Err()
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
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Create a new client using an InfluxDB server base URL and an authentication token
	influxDB = influxdb2.NewClient(influxURL, influxToken)
	radioAPI = influxDB.WriteAPIBlocking(influxOrg, influxBucket)
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

var DeviceAddress string

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

	p := influxdb2.NewPoint("device",
		map[string]string{"strength": "dBm", "address": device.Address.String(), "host": hostname},
		map[string]interface{}{"last": device.RSSI},
		time.Now())
	// write point immediately
	radioAPI.WritePoint(context.Background(), p)
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
