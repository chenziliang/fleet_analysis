package router

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
)

type coordinate struct {
	Lat float32 `json:"lat"`
	Lon float32 `json:"lon"`
}

func square(f float32) float32 {
	return f * f
}

// farOrClose determine if coordinate far or close to another coordinate
func (c coordinate) farOrClose(other *coordinate) bool {
	distance := square(c.Lat-other.Lat) + square(c.Lon-other.Lon)
	return distance < 0.005 || distance > 10
}

func (c coordinate) String() string {
	return fmt.Sprintf("[%f,%f]", c.Lat, c.Lon)
}

// data struct to load coordinate array
type coordinates []coordinate

type Generator struct {
	host          string
	elevation     string
	inputFile     string
	locale        string
	pointsEncoded string
	outputDir     string
	filepath      string
}

func NewGenerator(host, elevation, inputFile, locale, pointsEncoded, outputDir, filename string) *Generator {
	return &Generator{
		host:          host,
		pointsEncoded: pointsEncoded,
		locale:        locale,
		inputFile:     inputFile,
		elevation:     elevation,
		outputDir:     outputDir,
		filepath:      path.Join(outputDir, filename),
	}
}

// loadCoordinatesFromFile read all coordinates from input file
func (g Generator) loadCoordinatesFromFile() (*coordinates, error) {
	file, err := ioutil.ReadFile(g.inputFile)
	if err != nil {
		return nil, err
	}
	var coords coordinates
	if err = json.Unmarshal(file, &coords); err != nil {
		return nil, err
	}
	return &coords, nil
}

func notExist(fileOrDir string) bool {
	_, err := os.Stat(fileOrDir)
	return err != nil && os.IsNotExist(err)
}

func (g Generator) createFile() (*os.File, error) {
	if notExist(g.outputDir) {
		if err := os.Mkdir(g.outputDir, 0700); err != nil {
			return nil, err
		}
	}
	return os.Create(g.filepath)
}

func writeBatch(w *bufio.Writer, data []string) int {
	n := 0
	for _, d := range data {
		if _, err := fmt.Fprintln(w, d); err != nil {
			fmt.Printf("error %v\n", err)
			continue
		}
		n++
	}
	w.Flush()
	return n
}

func (g Generator) Generate(limit int) (err error) {
	if limit <= 0 {
		return nil
	}

	coords, err := g.loadCoordinatesFromFile()
	if err != nil {
		return err
	}
	locations := *coords
	numLoc := len(locations)

	fmt.Printf("Got %d locations from response\n", numLoc)

	if numLoc == 0 {
		return errors.New("generator: no locations available")
	}
	ghc := newGraphHopperClient(g.host, g.elevation, g.locale, g.pointsEncoded)
	numPath := 0
	numTotal := 0

	f, err := g.createFile()
	defer f.Close()

	writer := bufio.NewWriter(f)
	batch := []string{}

	var offset int // A random offset

	for i := 0; i < numLoc && numTotal < limit; i++ {
		offset = rand.Intn(16) + 1
		for j := i + offset; j < numLoc; j++ {
			from, to := &locations[i], &locations[j]
			if from.farOrClose(to) {
				// Skip if too far or too close
				continue
			}

			fmt.Printf("Generating path from %s to %s\n", from, to)

			r, err := ghc.GetPath(from, to)
			if err != nil {
				fmt.Printf("error %s\n", err)
				continue
			}
			numPath += 1
			batch = append(batch, string(r))

			if numPath%1000 == 0 || numPath >= limit {
				n := writeBatch(writer, batch)
				numTotal += n
				batch = batch[:0]
				fmt.Printf("Generated %d paths\n", numTotal)

				if numTotal >= limit {
					break
				}
			}
		}
	}

	fmt.Printf("Generated %d paths in total\n", numTotal)

	return nil
}
