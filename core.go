package airloctag

import (
	"crypto/sha512"
	"encoding/base32"
	"errors"
	"math"
	"sort"
	"strconv"

	"golang.org/x/crypto/sha3"
	"paepcke.de/airloctag/airports"
)

// provides an privacy-preseving [none-reverseable|hard-brute-force-able] uniq [key|salt]-able 3D location identifier
func encode(lat, long, elevation float64, keyin string, prec int) (tag, lists string, err error) {
	sep := "@" // indicator for default pre-computed static key, alias [unkeyed|public-hash]
	key := []byte("9485d3900f1c91e528a96c9290af88886408a9d5120db87be54d9bd60763244dc4167b2fca745d716e34767b3ed2d29d")
	if keyin != "" {
		sep = "#" // indicator for individual [key|salt]
		hash := sha512.Sum512([]byte(keyin))
		key = hash[:]
	}
	switch {
	case prec == 0: // precision zero defaults to 64 meter
		prec = 64
	case prec < 0:
		return "", "", errors.New("precision parameter to small, min: 1 meter")
	case prec > 64:
		return "", "", errors.New("precision parameter to high, max: 64 meter")
	}
	size := len(airports.Airports)
	dlist := make([]float64, 0, size)
	myDistance := make(map[float64]string, size)
	for place, co := range airports.Airports {
		d := dist(lat, long, elevation, co.A, co.O, co.L)
		if d > 1000000 { // skip all iata airports < 1000 km radius as landmarks
			myDistance[d] = place
			dlist = append(dlist, d)
		}
	}
	sort.Float64s(dlist)
	p := (512 - (prec*2)*4) + prec*2
	rounds := int(p/100) * 5
	skip := int(len(dlist) / p)
	var list, hash []byte
	for i := skip; i < (skip*p)-skip; i += skip {
		list = append(list, []byte(sep+myDistance[dlist[i]]+":"+strconv.Itoa(int(dlist[i])/1000))...)
		hash = hashWrap(append(hash, list...), rounds)
		switch { // list change via hash-state, to avoid pre-computed lists (to separate hash & fpu intensive task)
		case hash[0] < 10:
			i--
			i--
		case hash[0] < 128:
			i--
		}
	}
	hash = hashWrap(append(hash, key...), rounds)
	return genTag(hash, sep, rounds, prec), string(list), nil
}

// 3D distance between 2 points
func dist(xa, xo, xl, ya, yo, yl float64) float64 {
	xa, xo, ya, yo = xa*(math.Pi/180), xo*(math.Pi/180), ya*(math.Pi/180), yo*(math.Pi/180)
	h := math.Pow(math.Sin((ya-xa)/2), 2) + math.Cos(xa)*math.Cos(ya)*math.Pow(math.Sin((yo-xo)/2), 2)
	return 2*6378100*math.Asin(math.Sqrt(h)) + math.Abs(xl-yl)
}

// generate hash
func genTag(hash []byte, sep string, rounds, prec int) string {
	b := []byte(strconv.Itoa(prec) + sep)
	l := len(b)
	s := make([]byte, 64*2)
	base32.StdEncoding.Encode(s, hashWrap(hash, rounds*512))
	out := make([]byte, 22+l)
	copy(out[:l], b[:l])
	copy(out[l:4+l], s[:4])
	out[4+l] = '-'
	copy(out[5+l:7+l], s[4:6])
	out[7+l] = '-'
	copy(out[8+l:14+l], s[6:12])
	out[14+l] = '-'
	copy(out[15+l:17+l], s[12:14])
	out[17+l] = '-'
	copy(out[18+l:22+l], s[14:18])
	return string(out)
}

// sha512/sha3-256 sandwich wrap/hashchain
func hashWrap(in []byte, rounds int) []byte {
	if rounds < 1 {
		panic("invalid hashWrap rounds")
	}
	var hash, hsum [64]byte
	for i := 0; i < rounds; i++ {
		hash = sha512.Sum512(in)
		hsum = sha3.Sum512(hash[:])
		in = hsum[:]
	}
	return hsum[:]
}
