# hwinfo-to-influxdb

## Background
This is inspired by [nickbabcock/OhmGraphite](https://github.com/nickbabcock/OhmGraphite)

I use OhmGraphite to pipe host metrics from my PC to InfluxDB to use within Grafana visualizations. 
A recent Nvidia driver update caused one of the sensors in one of my graphs to not update data so I wrote this to grab the same data from hwinfo instead.

This could probably be extended to support OpenHardwareMonitor, GPU-Z, and AIDA64 as Remote Sensor Monitor supports pulling data from those programs as well.

## Prerequisites
* Windows 10 (HWiNFO is Windows only, InfluxDB can be hosted anywhere you please)
* HWiNFO installed: https://www.hwinfo.com/
* Remote Sensor Monitor add-on for HWiNFO: https://www.hwinfo.com/add-ons/
* InfluxDB database setup with JWT authorization: https://docs.influxdata.com/

## Install
1. Clone this repo onto the Windows machine you want to read sensor data from. 
   Alternatively you can set up Remote Sensor Monitor on the Windows machine to make the server available external to the machine and run this script from another machine that can connect to Remote Sensor Monitor.
    
2. Copy `sample-config.yml` to `config.yml` and populate with the information for your setup.

3. Make sure you create an environment variable for the JWT shared secret

4. Run the script `go run main.go`