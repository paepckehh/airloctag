// package main airportsgen 
package main

// import
import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

const (

	// files
	_file_source    = "airports.csv.zst"
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
	feed_chan := make(chan string, 2000)
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
		os.WriteFile(_file_out_nativ, content, 0o644)
		co.Done()
	}()

	// worker
	worker := (runtime.NumCPU() * 2) - 1
	bg := sync.WaitGroup{}
	bg.Add(worker)
	go func() {
		for i := 0; i < worker; i++ {
			go func() {
				for line := range feed_chan {
					s := strings.Split(line, ",")
					if len(s) < 8 {
						outSkip("[too few number of elements]: ", line)
						continue
					}
					if len(s[4]) != 5 {
						// skip all non-iata-3-letter-code airports
						// DEBUG ONLY: outSkip("[three letter site code]:     ", line)
						continue
					}
					lat, err := strconv.ParseFloat(s[6], 64)
					if err != nil {
						outSkip("[coordinates format, lat]:       ", line)
						continue
					}
					long, err := strconv.ParseFloat(s[7], 64)
					if err != nil {
						outSkip("[coordinates format, long]:      ", line)
						continue
					}
					elev, err := strconv.ParseFloat(s[8], 64)
					if err != nil {
						outSkip("[coordinates format, elevation]: ", line)
						continue
					}
					co_nativ_chan <- []byte("\t\"" + s[4][1:4] + "\":{A:" + fl(lat) + ",O:" + fl(long) + ",L:" + fl(elev) + "},\n")
				}
				bg.Done()
			}()
		}
		bg.Wait()
		close(co_nativ_chan)
	}()

	// feeder
	scanner, total := getFileScanner(_file_source), 0
	for scanner.Scan() {
		feed_chan <- scanner.Text()
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

func getFileScanner(filename string) (s *bufio.Scanner) {
	f, err := os.Open(filename)
	if err != nil {
		panic("unable to read db file [" + filename + "] [" + err.Error() + "]")
	}
	r, err := zstd.NewReader(f)
	if err != nil {
		panic("unable to read db file [" + filename + "] [" + err.Error() + "]")
	}
	return bufio.NewScanner(r)
}
