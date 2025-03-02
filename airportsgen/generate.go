// package main airportsgen
package main

// import
import (
	"encoding/csv"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (

	// files
	_filename       = "airport-codes.csv"
	_file_out_nativ = "../airports/airports.go"

	// components
	_header_nativ    = "//package airports\n// Code generated : IATA Airport gps coords as embedded golang map - DO NOT EDIT: "
	_package_nativ   = "package airports\n\n"
	_map_start_nativ = "// Airports ... \nvar Airports = map[string]Coord{\n"
	_map_end         = "}\n"
	_type_nativ      = "// Coord ...\ntype Coord struct {\n\tA float64 // Latitude\n\tO float64 // Longitude\n\tL float64 // Elevation\n}\n\n"
)

func main() {
	// global
	t0 := time.Now()
	feed_chan := make(chan []string, 2000)
	co_nativ_chan := make(chan []byte, 2000)

	// collect
	counter, size_nativ := 0, 0
	co := sync.WaitGroup{}
	co.Add(1)
	go func() {
		content := make([]byte, 0, 300*1024)
		content = append(content, []byte(_header_nativ+time.Now().Format(time.RFC3339)+"\n")...)
		content = append(content, []byte(_package_nativ)...)
		content = append(content, []byte(_type_nativ)...)
		content = append(content, []byte(_map_start_nativ)...)
		for line := range co_nativ_chan {
			content = append(content, line...)
			counter++
		}
		content = append(content, []byte(_map_end)...)
		size_nativ = len(content) / 1024
		err := os.WriteFile(_file_out_nativ, content, 0o644)
		if err != nil {
			panic("[airportgen] unable to write generated output file")
		}
		co.Done()
	}()

	// worker
	worker := (runtime.NumCPU() * 2) - 1
	bg := sync.WaitGroup{}
	bg.Add(worker)
	go func() {
		for i := 0; i < worker; i++ {
			go func() {
				for s := range feed_chan {
					if len(s) < 11 {
						outSkip("[too few number of elements]: ", strings.Join(s, ","))
						continue
					}
					if len(s[9]) != 3 {
						/// fmt.Printf("not 3 letter iata code => %v\n", s[10])
						// skip all non-iata-3-letter-code airports
						// DEBUG ONLY: outSkip("[three letter site code]:     ", strings.Joint(s,","))
						continue
					}
					longlat := strings.Split(s[12], ",")
					lat, err := strconv.ParseFloat(strings.ReplaceAll(longlat[0], " ", ""), 64)
					if err != nil {
						outSkip("[coordinates format, lat]:       ", strings.Join(s, ","))
						// fmt.Printf("\n%v => %v", strings.ReplaceAll(longlat[0], " ", ""), err.Error())
						continue
					}
					long, err := strconv.ParseFloat(strings.ReplaceAll(longlat[1], " ", ""), 64)
					if err != nil {
						outSkip("[coordinates format, long]:      ", strings.Join(s, ","))
						// fmt.Printf("\n%v => %v", strings.ReplaceAll(longlat[1], " ", ""), err.Error())
						continue
					}
					elev, err := strconv.ParseFloat(s[3], 64)
					if err != nil {
						// fmt.Printf("\n Elevation VALUE=%v ERROR:%v \n", s[3], err.Error())
						// outSkip("[coordinates format, elevation]: ", strings.Join(s, ","))
						// continue
						elev = 0
					}
					co_nativ_chan <- []byte("\t\"" + s[9] + "\":{A:" + fl(lat) + ",O:" + fl(long) + ",L:" + fl(elev) + "},\n")
				}
				bg.Done()
			}()
		}
		bg.Wait()
		close(co_nativ_chan)
	}()

	// feeder
	total := 0
	csv := readcsv(_filename)
	for _, line := range csv {
		feed_chan <- line
		total++
	}
	close(feed_chan)
	co.Wait()
	out("Generated golang nativ code: [" + strconv.Itoa(size_nativ) + "k] [" + _file_out_nativ + "]")
	out("Input lines processed:       " + strconv.Itoa(total))
	out("IATA registered Airports:    " + strconv.Itoa(counter))
	out("Time needed:                 " + time.Since(t0).String())
}

//
// LITTLE HELPER
//

func out(message string) {
	os.Stdout.Write([]byte(message + "\n"))
}

func outSkip(reason, line string) {
	if len(line) > 80 {
		line = line[:79] + "..."
	}
	out("skipping line " + reason + line)
}

func fl(in float64) string {
	return strconv.FormatFloat(in, 'f', -1, 64)
}

func readcsv(filename string) [][]string {
	f, err := os.Open(filename)
	if err != nil {
		panic("unable to read db file [" + filename + "] [" + err.Error() + "]")
	}
	csv, err := csv.NewReader(f).ReadAll()
	if err != nil {
		panic("unable to parse db file [" + filename + "] [" + err.Error() + "]")
	}
	return csv
}
