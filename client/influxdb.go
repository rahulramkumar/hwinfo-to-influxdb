package client

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"time"
)

var SharedSecret string

var TokenString string

type InfluxDBClient struct {
	remoteEndpoint string
	org            string
	bucketName     string
	username       string
	client         influxdb2.Client
	tokenExpiry    int64
}

type WithAuthHeader struct {
	rt http.RoundTripper
}

func (wah WithAuthHeader) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", "Bearer "+TokenString)
	return wah.rt.RoundTrip(r)
}

func NewInfluxDBClient(remoteEndpoint string, org string, bucketName string, username string, sharedSecretEnv string) (*InfluxDBClient, error) {
	SharedSecret = os.Getenv(sharedSecretEnv)
	client := &InfluxDBClient{
		remoteEndpoint: remoteEndpoint,
		bucketName:     bucketName,
		username:       username,
		org:            org,
	}

	// Create JWT
	err := client.rotateToken()
	if err != nil {
		return nil, errors.Wrap(err, "creating new jwt")
	}

	// Add HTTP Transport to add Authorization header to each request to InfluxDB endpoint
	httpClient := &http.Client{
		Transport: WithAuthHeader{rt: http.DefaultTransport},
	}

	// Empty Auth token since we're using JWT
	influxClient := influxdb2.NewClientWithOptions(remoteEndpoint, "",
		influxdb2.DefaultOptions().SetHTTPClient(httpClient))
	client.client = influxClient

	return client, nil
}

func (c *InfluxDBClient) PostToServer(sensorReadings *[]HWinfoSensorReading) error {
	if c.tokenExpiry < time.Now().Unix() {
		fmt.Println("Rotating JWT")
		err := c.rotateToken()
		if err != nil {
			return errors.Wrap(err, "rotating existing jwt")
		}
	}

	writeAPI := c.client.WriteAPI(c.org, c.bucketName)

	errorChannel := writeAPI.Errors()
	go func() {
		for err := range errorChannel {
			fmt.Printf("write error: %s\n", err.Error())
		}
	}()

	for _, reading := range *sensorReadings {
		point := createInfluxDBPoint(&reading)
		writeAPI.WritePoint(point)
	}
	writeAPI.Flush()

	return nil
}

func (c *InfluxDBClient) rotateToken() error {
	exp := time.Now().Unix() + 3600 // JWT rotates once per hour
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": c.username,
		"exp":      exp,
	})
	signedString, err := token.SignedString([]byte(SharedSecret))
	if err != nil {
		return err
	}
	TokenString = signedString
	c.tokenExpiry = exp

	return nil
}

func createInfluxDBPoint(reading *HWinfoSensorReading) (influxDBWritePoint *write.Point) {
	return write.NewPoint(
		reading.Class,
		map[string]string{
			"sensor_name": reading.Name,
		},
		map[string]interface{}{
			"value": reading.Value,
		},
		reading.GoTime,
	)
}
