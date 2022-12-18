// package locenv
package locenv

// import
import (
	"errors"
	"os"
	"strconv"

	"paepcke.de/airloctag/airports"
)

// const
const (
	// env variable name definitions
	_ENV_LAT       = "GPS_LAT"
	_ENV_LONG      = "GPS_LONG"
	_ENV_ELEVATION = "GPS_ELEVATION"
	_ENV_IATA      = "IATA"
	// error messages
	_ERR_MISSING = "enviroment variable missing: "
	_ERR_INVALID = "enviroment variable with invalid data: "
)

//
// EXTERNAL INTERFACES
//

// Get ...
func Get() (lat, long, elevation float64, err error) {
	var empty bool
	lat, long, elevation, empty, err = getCoord()
	if empty {
		lat, long, elevation, empty, err = getIata()
		if empty {
			err = errors.New(_ERR_MISSING + " {any}")
			return lat, long, elevation, err
		}
	}
	return lat, long, elevation, err
}

//
// INTERNAL BACKEND
//

func getCoord() (lat, long, elevation float64, empty bool, err error) {
	x := ""
	empty = true
	x = os.Getenv(_ENV_LAT)
	if x == "" {
		err = errors.New(_ERR_MISSING + _ENV_LAT)
		return lat, long, elevation, empty, err
	}
	empty = false
	lat, err = strconv.ParseFloat(x, 64)
	if err != nil {
		err = errors.New(_ERR_INVALID + _ENV_LAT + " [" + err.Error() + "]")
		return lat, long, elevation, empty, err
	}
	x = os.Getenv(_ENV_LONG)
	if x == "" {
		err = errors.New(_ERR_MISSING + _ENV_LONG)
		return lat, long, elevation, empty, err
	}
	long, err = strconv.ParseFloat(x, 64)
	if err != nil {
		err = errors.New(_ERR_INVALID + _ENV_LONG + " [" + err.Error() + "]")
		return lat, long, elevation, empty, err
	}
	x = os.Getenv(_ENV_ELEVATION)
	if x != "" {
		elevation, err = strconv.ParseFloat(x, 64)
		if err != nil {
			err = errors.New(_ERR_INVALID + _ENV_ELEVATION + " [" + err.Error() + "]")
			return lat, long, elevation, empty, err
		}
	}
	return lat, long, elevation, false, nil
}

func getIata() (lat, long, elevation float64, empty bool, err error) {
	x := ""
	x = os.Getenv(_ENV_IATA)
	empty = true
	if x == "" {
		err = errors.New(_ERR_MISSING + _ENV_IATA)
		return lat, long, elevation, empty, err
	}
	empty = false
	if len(x) != 3 {
		err = errors.New(_ERR_INVALID + _ENV_IATA)
		return lat, long, elevation, empty, err
	}
	loc, ok := airports.Airports[x]
	if !ok {
		err = errors.New(_ERR_INVALID + _ENV_IATA)
		return lat, long, elevation, empty, err
	}
	return loc.A, loc.O, loc.L, false, nil
}
