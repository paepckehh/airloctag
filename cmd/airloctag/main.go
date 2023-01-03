// package main
package main

import (
	"os"
	"strconv"

	"paepcke.de/airloctag"
	"paepcke.de/airloctag/locenv"
)

// main ...
func main() {
	var (
		err       error
		key       string
		verbose   bool
		precision int
	)
	verbose = true // debug mode
	l := len(os.Args)
	switch {
	case l > 1:
		for i := 1; i < l; i++ {
			o := os.Args[i]
			switch o {
			case "--precision", "-p":
				precision, err = strconv.Atoi(o)
				if err != nil || precision > 64 || precision < 1 {
					out("invalid optional precision parameter [valid:1-64], example: airloctag 32")
					os.Exit(1)
				}
			case "--key", "-k":
				key = o
			case "--verbose", "-v":
				// currently default on via debug mode
				// verbose = true
			case "--help", "-h":
				out(_syntax)
				os.Exit(0)
			default:
				out("[error] [unknown option] [" + o + "]")
				os.Exit(1)
			}
		}
	}
	lat, long, elevation, err := locenv.Get()
	if err != nil {
		out("[error] [env variable provided location] [" + err.Error() + "]")
		os.Exit(1)
	}
	hash, hashbase, err := airloctag.Encode(lat, long, elevation, key, precision)
	if err != nil {
		out("[error] [" + err.Error())
		os.Exit(1)
	}
	switch verbose {
	case true:
		out("TAG:" + hash + "\nDEBUG:" + hashbase)
	default:
		out("TAG:" + hash)
	}
}

// const ...
const _syntax = "syntax: airloctag [options]\n\n-k --key\n\t\toptional custom key parameter\n\n-p --precision\n\t\toptional precision parameter, range 1-64\n\n-v --verbose\n\t\tverbose debug output\n\nEXAMPLE\n\nairloctag --verbose --precision 32 --key $MYSECRECTKEY\n\nENV VARIABLES\n\nGPS_LAT,GPS_LONG, GPS_ELEVATION\t\tspecify your current location via GPS coordinates\nIATA\t\t\t\t\tspecify your current location via nearest Airport IATA code\n"

//
// LITTLE GENERIC HELPER SECTION
//

// func out
func out(message string) {
	os.Stdout.Write([]byte(message + "\n"))
}
