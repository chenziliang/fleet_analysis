package main

import (
	"flag"
	"fmt"
	"github.com/fleet_analysis/routergenerator/router"
	"os"
)

const (
	RouteApiHost = "http://localhost:8989/route"
)

func main() {
	var elevation = flag.String("elevation", "false",
		"If true a third dimension - the elevation - is included in "+
			"the polyline or in the GeoJson. ")
	var locale = flag.String("locale", "en-US",
		"The locale of the resulting turn instructions. E.g. pt_PT for"+
			" Portuguese or de for German.")
	var pointsEncoded = flag.String("points_encoded", "false",
		"If false the coordinates in point and snapped_waypoints are"+
			" returned as array using the order [lon,lat,elevation]"+
			" for every point. If true the coordinates will be encoded"+
			" as string leading to less bandwith usage. ")
	var outputDir = flag.String("outputDir", "./output",
		"The directory used to store route results")
	var filename = flag.String("file", "paths.json",
		"The name of file used to store route results")

	// Will generate n * (n - 1) / 2 paths
	var n = flag.Int("num", 10, "The number of path to generate")

	// Will generate all paths between any two coordinates in inputFile
	var inputFile = flag.String("inputfile", "./routergenerator/location/locations.json", "The input file of coordinates")

	flag.Parse()

	g := router.NewGenerator(RouteApiHost, *elevation, *inputFile, *locale, *pointsEncoded, *outputDir, *filename)

	if err := g.Generate(*n); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
