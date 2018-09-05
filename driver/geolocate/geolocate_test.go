// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package geolocate

import (
	"bytes"
	"testing"
)

var lasPavasBlob = `
{ "type": "FeatureCollection",
"features": [
{ "type": "Feature",
"geometry": {"type": "Point", "coordinates": [-64.583333, -22.466667]},
"properties": {
"parsePattern" : "LAS PAVAS",
"precision" : "High",
"score" : 84,
"uncertaintyRadiusMeters" : 301,
"uncertaintyPolygon" : "Unavailable",
"displacedDistanceMiles" : 0,
"displacedHeadingDegrees" : 0,
"debug" : ":GazPartMatch=False|:inAdm=True|:Adm=SALTA|:NPExtent=500|:NP=LAS PAVAS|:KFID=|LAS PAVAS"
}
},
{ "type": "Feature",
"geometry": {"type": "Point", "coordinates": [-65.873989, -27.253746]},
"properties": {
"parsePattern" : "LAS PAVAS",
"precision" : "High",
"score" : 84,
"uncertaintyRadiusMeters" : 301,
"uncertaintyPolygon" : "Unavailable",
"displacedDistanceMiles" : 0,
"displacedHeadingDegrees" : 0,
"debug" : ":GazPartMatch=False|:inAdm=True|:Adm=TUCUMÁN|:NPExtent=500|:NP=LAS PAVAS|:KFID=|LAS PAVAS"
}
},
{ "type": "Feature",
"geometry": {"type": "Point", "coordinates": [-64.583333, -24.8]},
"properties": {
"parsePattern" : "LAS PAVAS",
"precision" : "High",
"score" : 84,
"uncertaintyRadiusMeters" : 301,
"uncertaintyPolygon" : "Unavailable",
"displacedDistanceMiles" : 0,
"displacedHeadingDegrees" : 0,
"debug" : ":GazPartMatch=False|:inAdm=True|:Adm=SALTA|:NPExtent=500|:NP=LAS PAVAS|:KFID=|LAS PAVAS"
}
}
 ],
"crs": { "type" : "EPSG", "properties" : { "code" : 4326 }}
}
`

var errorBlob = "No Locality Provided"

func TestDecodePointList(t *testing.T) {
	b := bytes.NewBufferString(lasPavasBlob)
	ls, err := decodePointList(b)
	if err != nil {
		t.Errorf("decoding error on lasPavasBlob: %v", err)
	}
	if len(ls.Features) != 3 {
		t.Errorf("number of points %d, want %d", len(ls.Features), 3)
	}

	b = bytes.NewBufferString(errorBlob)
	ls, err = decodePointList(b)
	if err == nil {
		t.Errorf("decoding error wanted")
	}
}
