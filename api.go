// Package airloctag encodes gps coordinates into an airloctag
package airloctag

// Encode encodes gps coordinates data (lat,long,elevation, optional: private-key) into an airloctag
func Encode(lat, long, elevation float64, keyin string, prec int) (tag, lists string, err error) {
	return encode(lat, long, elevation, keyin, prec)
}
