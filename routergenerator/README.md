# How to run

1. Download pbf file of New York city [new-york-latest.osm.pbf](http://download.geofabrik.de/north-america/us/new-york.html)
2. Download graphhopper [graphhopper](https://github.com/graphhopper/graphhopper)
3. Start local Graphhopper with the downloaded PBF file: ```./graphhopper.sh web new-york-latest.osm.pbf```
4. Run ```cd <project folder>/routergenerator && go run main.go``` in the project root folder. User can customize some arguments for example the number of paths to generate (10 by default): ```go run main.go -num 100000```. For other supported arguments, please take a look at ```main.go```.
5. The path file will be stored in ```<project folder>/routergenerator/output/paths.json```.


# How does it work

1. The program parses the PBF file of New York city and got a random coordinates set and stored in file ```<project folder>/routergenerator/location/locations.json```.
2. For each two coordinates loaded from file locations.json, check if they are near to each other and skip them if they're far away to each other. If too far, the path will contains a lot a points.
3. Send request to API of Graphhoper on localhost: ```http://localhost:8989/route```. The API endpoint is a little different with the public service of Graphhopper.
4. Store the response to file for each two coordinates until reach the number user want to get.
