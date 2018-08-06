// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package gbif

import (
	"bytes"
	"testing"
)

func TestDecodeTaxonList(t *testing.T) {
	b := bytes.NewBufferString(brontoBlob)
	sp, err := decodeTaxonList(b)
	if err != nil {
		t.Errorf("decoding error on brontoBlob: %v", err)
	}

	if len(sp.Results) != 4 {
		t.Errorf("number of species %d, want %d", len(sp.Results), 4)
	}

	b = bytes.NewBufferString(noTaxBlob)
	sp, _ = decodeTaxonList(b)
	if len(sp.Results) != 0 {
		t.Errorf("number of species %d, want %d", len(sp.Results), 0)
	}
}

func TestDecodeTaxon(t *testing.T) {
}

var brontoBlob = `
{"offset":0,"limit":300,"endOfRecords":true,"results":[
	{"key":7889012,"nameKey":16133223,"taxonID":"gbif:7889012","sourceTaxonKey":144289144,"kingdom":"Animalia","phylum":"Arthropoda","order":"Hemiptera","family":"Reduviidae","genus":"Brontostoma","species":"Brontostoma discus","kingdomKey":1,"phylumKey":54,"classKey":216,"orderKey":809,"familyKey":4324,"genusKey":4773811,"speciesKey":7889012,"datasetKey":"d7dddbf4-2cf0-4f39-9b2a-bb099caae36c","constituentKey":"9ca92552-f23a-41a8-a140-01abaa31c931","parentKey":4773811,"parent":"Brontostoma","scientificName":"Brontostoma discus (Burmeister, 1835)","canonicalName":"Brontostoma discus","authorship":" (Burmeister, 1835)","nameType":"SCIENTIFIC","rank":"SPECIES","origin":"SOURCE","taxonomicStatus":"ACCEPTED","nomenclaturalStatus":[],"remarks":"","numDescendants":2,"lastCrawled":"2018-06-20T14:41:51.801+0000","lastInterpreted":"2018-06-20T13:58:15.692+0000","issues":[],"synonym":false,"class":"Insecta"},
	{"key":119338319,"nubKey":7889012,"nameKey":1665729,"taxonID":"03B087B69E55747DFA612F8AD492D8C2.taxon","kingdom":"Animalia","phylum":"Arthropoda","order":"Hemiptera","family":"Reduviidae","genus":"Brontostoma","species":"Brontostoma discus","kingdomKey":143840506,"phylumKey":143840508,"classKey":143840509,"orderKey":143840510,"familyKey":143840513,"genusKey":143840524,"speciesKey":119338319,"datasetKey":"06ab81eb-f8e8-4b77-ad90-758c5ff37337","parentKey":143840524,"parent":"Brontostoma","scientificName":"Brontostoma discus Burmeister 1835","canonicalName":"Brontostoma discus","authorship":"Burmeister, 1835","nameType":"SCIENTIFIC","rank":"SPECIES","origin":"SOURCE","taxonomicStatus":"ACCEPTED","nomenclaturalStatus":[],"numDescendants":0,"references":"http://treatment.plazi.org/id/03B087B69E55747DFA612F8AD492D8C2","lastCrawled":"2018-06-25T16:50:14.651+0000","lastInterpreted":"2018-05-08T16:58:08.458+0000","issues":[],"synonym":false,"class":"Insecta"},
	{"key":134686291,"nubKey":7889012,"nameKey":16133223,"taxonID":"748806","kingdom":"Animalia","phylum":"Arthropoda","order":"Hemiptera","family":"Reduviidae","genus":"Brontostoma","species":"Brontostoma discus","kingdomKey":134521164,"phylumKey":134603258,"classKey":134630959,"orderKey":134682251,"familyKey":134685070,"genusKey":134686286,"speciesKey":134686291,"datasetKey":"0e61f8fe-7d25-4f81-ada7-d970bbb2c6d6","parentKey":134686286,"parent":"Brontostoma","scientificName":"Brontostoma discus (Burmeister, 1835)","canonicalName":"Brontostoma discus","authorship":" (Burmeister, 1835)","nameType":"SCIENTIFIC","rank":"SPECIES","origin":"SOURCE","taxonomicStatus":"ACCEPTED","nomenclaturalStatus":[],"numDescendants":0,"references":"https://inpn.mnhn.fr/espece/cd_nom/748806","lastCrawled":"2018-06-25T19:40:40.264+0000","lastInterpreted":"2017-09-20T15:15:00.154+0000","issues":["SCIENTIFIC_NAME_ASSEMBLED","CLASSIFICATION_NOT_APPLIED"],"synonym":false,"class":"Hexapoda"},
	{"key":144289144,"nubKey":7889012,"nameKey":16133223,"taxonID":"1069430","kingdom":"Animalia","phylum":"Arthropoda","order":"Hemiptera","family":"Reduviidae","genus":"Brontostoma","species":"Brontostoma discus","kingdomKey":101683523,"phylumKey":101701105,"classKey":101760843,"orderKey":101855098,"familyKey":101859209,"genusKey":144289121,"speciesKey":144289144,"datasetKey":"9ca92552-f23a-41a8-a140-01abaa31c931","parentKey":144289121,"parent":"Brontostoma","scientificName":"Brontostoma discus (Burmeister, 1835)","canonicalName":"Brontostoma discus","authorship":" (Burmeister, 1835)","nameType":"SCIENTIFIC","rank":"SPECIES","origin":"SOURCE","taxonomicStatus":"ACCEPTED","nomenclaturalStatus":[],"numDescendants":2,"lastCrawled":"2018-07-28T09:01:30.419+0000","lastInterpreted":"2018-07-28T10:02:49.795+0000","issues":["SCIENTIFIC_NAME_ASSEMBLED"],"synonym":false,"class":"Insecta"}
]}
`

var noTaxBlob = `{"offset":0,"limit":300,"endOfRecords":true,"results":[]}`
