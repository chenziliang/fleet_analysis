package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type GraphHopperClient struct {
	host          string
	elevation     string
	locale        string
	pointsEncoded string
	client        *http.Client
}

func newGraphHopperClient(host, elevation, locale, pointsEncoded string) *GraphHopperClient {
	return &GraphHopperClient{
		host:          host,
		pointsEncoded: pointsEncoded,
		locale:        locale,
		elevation:     elevation,
		client:        &http.Client{},
	}
}

// assemblePoint create a point as order "latitude,longitude"
func (g GraphHopperClient) assemblePoint(lat, lon float32) string {
	return fmt.Sprintf("%f,%f", lat, lon)
}

func (g GraphHopperClient) newRequest(from, to *coordinate) (*http.Request, error) {
	req, err := http.NewRequest("GET", g.host, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("points_encoded", g.pointsEncoded)
	q.Add("elevation", g.elevation)
	q.Add("point", g.assemblePoint(from.Lat, from.Lon))
	q.Add("point", g.assemblePoint(to.Lat, to.Lon))
	// type must be json
	q.Add("type", "json")
	q.Add("locale", g.locale)
	//q.Add("instructions", "false")
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func (g GraphHopperClient) request(r *http.Request) ([]byte, error) {
	//fmt.Printf("Requesting %s\n", r.URL.RawQuery)

	resp, err := g.client.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("graphhopper: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (g GraphHopperClient) GetPath(from, to *coordinate) ([]byte, error) {
	r, err := g.newRequest(from, to)
	if err != nil {
		return nil, err
	}
	p, err := g.request(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}
