package main

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"hwinfo-to-influxdb/client"
	"os"
	"strconv"
	"time"
)

var log = logrus.New()

type Config struct {
	HWinfoConfig struct {
		ConnectionString string `yaml:"connection_string"`
		RefreshInterval  int    `yaml:"refresh_interval"`
	} `yaml:"hwinfo_config"`
	InfluxDBConfig struct {
		ConnectionString string `yaml:"connection_string"`
		Org              string `yaml:"org"`
		Bucket           string `yaml:"bucket"`
		Username         string `yaml:"username"`
		SharedSecretEnv  string `yaml:"shared_secret_env"`
	} `yaml:"influxdb_config"`
}

func main() {
	log.Out = os.Stdout
	if val, exists := os.LookupEnv("LOG_LEVEL"); exists {
		logLevel, err := logrus.ParseLevel(val)
		if err != nil {
			log.Fatal(err)
		}
		log.Level = logLevel
	} else {
		log.Level = logrus.DebugLevel
	}

	config, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	hwinfoClient, err := client.NewHWinfoClient(config.HWinfoConfig.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	influxdbClient, err := client.NewInfluxDBClient(
		config.InfluxDBConfig.ConnectionString,
		config.InfluxDBConfig.Org,
		config.InfluxDBConfig.Bucket,
		config.InfluxDBConfig.Username,
		config.InfluxDBConfig.SharedSecretEnv)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting hwinfo-to-influxdb - %ds refresh interval", config.HWinfoConfig.RefreshInterval)
	for range time.Tick(time.Duration(config.HWinfoConfig.RefreshInterval) * time.Second) {
		start := time.Now().UnixNano()
		sensorReadings, err := hwinfoClient.GetCurrentSensorReadings()
		if err != nil {
			log.Fatal(err)
		}
		err = influxdbClient.PostToServer(sensorReadings)
		if err != nil {
			log.Fatal(err)
		}
		end := time.Now().UnixNano()

		log.WithFields(logrus.Fields{
			"elapsed": strconv.FormatInt((end-start)/int64(time.Millisecond), 10) + " ms",
		}).Debug("finished scanning sensor data and wrote to influxdb")
	}

}

func parseConfig() (*Config, error) {
	f, err := os.Open("config.yml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
