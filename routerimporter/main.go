package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type pathHandler struct {
	filepath       string
	measurement    string
	influxdbClient client.Client
	config         client.BatchPointsConfig
}

func newPathHandler(filepath, measurement, database, precision, addr, username, password string) (*pathHandler, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	return &pathHandler{
		filepath:       filepath,
		measurement:    measurement,
		influxdbClient: c,
		config: client.BatchPointsConfig{
			Database:  database,
			Precision: precision,
		},
	}, nil
}

type jsonPath struct {
	Paths []path `json:"paths"`
}

type path struct {
	Points points `json:"points"`
}

type position [2]float64

type points struct {
	Coordinates []position `json:"coordinates"`
}

func float2String(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}

func (p pathHandler) parseAndIndex(s *string) (err error) {
	var r jsonPath
	err = json.Unmarshal([]byte(*s), &r)
	if err != nil {
		return err
	}
	paths := r.Paths

	if len(paths) == 0 {
		return fmt.Errorf("routerimporter: invalid path %s", *s)
	}

	batch, err := client.NewBatchPoints(p.config)
	if err != nil {
		return err
	}

	coords := paths[0].Points.Coordinates
	// the start position
	src := coords[0]
	// the destination position
	dest := coords[len(coords)-1]

	tags := map[string]string{
		"src.lat":  float2String(src[0]),
		"src.lon":  float2String(src[1]),
		"dest.lat": float2String(dest[0]),
		"dest.lon": float2String(dest[1]),
	}

	since := time.Date(2017, 1, 1, 1, 34, 58, 651387237, time.UTC)

	for _, c := range coords {
		// Each coordinate as a value
		fields := map[string]interface{}{
			"lat": c[0],
			"lon": c[1],
		}

		since = since.Add(time.Second)
		point, err := client.NewPoint(p.measurement, tags, fields, since)
		if err != nil {
			return err
		}
		batch.AddPoint(point)
	}

	if err := p.influxdbClient.Write(batch); err != nil {
		return err
	}
	return nil
}

func (p pathHandler) run() error {
	file, err := os.Open(p.filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var s = scanner.Text()
		err = p.parseAndIndex(&s)
		if err != nil {
			fmt.Printf("error %v\n", err)
		}
	}
	if err != nil {
		return err
	}
	return scanner.Err()
}

func main() {
	var filepath = flag.String("file", "./output/paths.json",
		"The name of file used to read path file")
	var addr = flag.String("addr", "http://localhost:8086", "The influxdb host address")

	var measurement = flag.String("measurement", "my_test", "The measurement used to store data")

	var db = flag.String("db", "mydb", "The name of database used to store data")
	flag.Parse()

	// no authentication for now
	handler, err := newPathHandler(*filepath, *measurement, *db, "s", *addr, "", "")

	if err != nil {
		fmt.Printf("cannot connect to %s\n", addr)
		return
	}

	if err = handler.run(); err != nil {
		fmt.Printf("error %v\n", err)
	}
}
