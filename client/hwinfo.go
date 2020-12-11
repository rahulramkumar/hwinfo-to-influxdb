package client

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

type HWinfoSensorReading struct {
	App        string  `json:"SensorApp"`
	Class      string  `json:"SensorClass"`
	Name       string  `json:"SensorName"`
	Value      float64 `json:"SensorValue,string"`
	Unit       string  `json:"SensorUnit"`
	UpdateTime int64   `json:"SensorUpdateTime"`
	GoTime     time.Time
}

type HWinfoClient struct {
	RemoteSensorMonitorEndpoint string
}

func NewHWinfoClient(remoteSensorMonitorURL string) (*HWinfoClient, error) {
	resp, err := http.Get(remoteSensorMonitorURL)
	if err != nil {
		return nil, errors.Wrap(err, "requesting remote sensor monitor endpoint")
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("remote sensor endpoint response with non 200 HTTP status: %s\n", resp.Status)
	}
	return &HWinfoClient{RemoteSensorMonitorEndpoint: remoteSensorMonitorURL}, nil
}

func (c *HWinfoClient) GetCurrentSensorReadings() (*[]HWinfoSensorReading, error) {
	resp, err := http.Get(c.RemoteSensorMonitorEndpoint)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sensorReadings, err := c.convertRawJsonToSensorReadings(body)
	if err != nil {
		return nil, err
	}
	c.parseGoTime(sensorReadings)
	return sensorReadings, nil
}

func (c *HWinfoClient) convertRawJsonToSensorReadings(jsonString []byte) (*[]HWinfoSensorReading, error) {
	var output []HWinfoSensorReading
	err := json.Unmarshal(jsonString, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (c *HWinfoClient) parseGoTime(sensorReadings *[]HWinfoSensorReading) {
	var goTime time.Time
	for i, reading := range *sensorReadings {
		goTime = time.Unix(reading.UpdateTime, 0)
		(*sensorReadings)[i].GoTime = goTime
	}
}
