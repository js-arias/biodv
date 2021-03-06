// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package gbif

import (
	"bytes"
	"testing"

	"github.com/js-arias/biodv"
)

func TestDecodeRecordList(t *testing.T) {
	b := bytes.NewBufferString(larusBlob)
	rc, err := decodeRecordList(b)
	if err != nil {
		t.Errorf("decoding error on larusBlob: %v", err)
	}
	if len(rc.Results) != 5 {
		t.Errorf("number of records %d, want %d", len(rc.Results), 5)
	}

	b = bytes.NewBufferString(noRecsBlob)
	rc, _ = decodeRecordList(b)
	if len(rc.Results) != 0 {
		t.Errorf("number of records %d, wan %d", len(rc.Results), 0)
	}
}

var larusBlob = `
{"offset":0,"limit":300,"endOfRecords":true,"count":8474604,"results":[
	{"key":1494057472,"datasetKey":"83e20573-f7dd-4852-9159-21566e1e691e","publishingOrgKey":"1cd669d0-80ea-11de-a9d0-f1765f95f18b","networkKeys":[],"installationKey":"9f25fd85-85dc-4dcd-a1b4-b31165442e2b","publishingCountry":"BE","protocol":"DWC_ARCHIVE","lastCrawled":"2017-10-08T08:55:41.659+0000","lastParsed":"2017-04-27T20:20:28.400+0000","crawlId":49,"extensions":{},"basisOfRecord":"MACHINE_OBSERVATION","sex":"FEMALE","lifeStage":"ADULT","taxonKey":2481139,"kingdomKey":1,"phylumKey":44,"classKey":212,"orderKey":7192402,"familyKey":9316,"genusKey":2481126,"speciesKey":2481139,"scientificName":"Larus argentatus Pontoppidan, 1763","kingdom":"Animalia","phylum":"Chordata","order":"Charadriiformes","family":"Laridae","genus":"Larus","species":"Larus argentatus","genericName":"Larus","specificEpithet":"argentatus","taxonRank":"SPECIES","decimalLongitude":1.596802,"decimalLatitude":50.223982,"coordinateUncertaintyInMeters":7.0,"elevation":0.0,"year":2016,"month":1,"day":1,"eventDate":"2016-01-01T11:50:15.000+0000","issues":["COORDINATE_ROUNDED","COUNTRY_DERIVED_FROM_COORDINATES"],"lastInterpreted":"2018-07-08T00:45:25.746+0000","license":"http://creativecommons.org/publicdomain/zero/1.0/legalcode","identifiers":[],"facts":[],"relations":[],"geodeticDatum":"WGS84","class":"Aves","countryCode":"FR","country":"France","rightsHolder":"INBO","minimumDistanceAboveSurfaceInMeters":"2","identifier":"6036:20160101115015","informationWithheld":"see metadata","dynamicProperties":"{\"device_info_serial\":6036, \"catch_location\":\"Vismijn, Oostende\", \"tracking_started_at\":\"2014-06-13\"}","samplingEffort":"{\"seconds_since_last_occurrence\":176}","georeferenceVerificationStatus":"unverified","datasetName":"Bird tracking - GPS tracking of Lesser Black-backed Gulls and Herring Gulls breeding at the southern North Sea coast","language":"en","gbifID":"1494057472","occurrenceID":"6036:20160101115015","type":"Event","vernacularName":"Herring Gull","organismID":"H903607","institutionCode":"INBO","organismName":"Stefanie","ownerInstitutionCode":"INBO/VLIZ/UGent/UA","samplingProtocol":"https://doi.org/10.1007/s10336-012-0908-1","datasetID":"https://doi.org/10.15468/02omly","accessRights":"http://www.inbo.be/en/norms-for-data-use","georeferenceSources":"GPS"},
	{"key":1494057832,"datasetKey":"83e20573-f7dd-4852-9159-21566e1e691e","publishingOrgKey":"1cd669d0-80ea-11de-a9d0-f1765f95f18b","networkKeys":[],"installationKey":"9f25fd85-85dc-4dcd-a1b4-b31165442e2b","publishingCountry":"BE","protocol":"DWC_ARCHIVE","lastCrawled":"2017-10-08T08:55:41.769+0000","lastParsed":"2017-04-27T20:20:28.550+0000","crawlId":49,"extensions":{},"basisOfRecord":"MACHINE_OBSERVATION","sex":"FEMALE","lifeStage":"ADULT","taxonKey":2481174,"kingdomKey":1,"phylumKey":44,"classKey":212,"orderKey":7192402,"familyKey":9316,"genusKey":2481126,"speciesKey":2481174,"scientificName":"Larus fuscus Linnaeus, 1758","kingdom":"Animalia","phylum":"Chordata","order":"Charadriiformes","family":"Laridae","genus":"Larus","species":"Larus fuscus","genericName":"Larus","specificEpithet":"fuscus","taxonRank":"SPECIES","decimalLongitude":-0.010204,"decimalLatitude":48.866221,"coordinateUncertaintyInMeters":5.0,"elevation":0.0,"year":2016,"month":1,"day":1,"eventDate":"2016-01-01T14:00:06.000+0000","issues":["COORDINATE_ROUNDED","COUNTRY_DERIVED_FROM_COORDINATES"],"lastInterpreted":"2018-07-08T00:50:41.146+0000","license":"http://creativecommons.org/publicdomain/zero/1.0/legalcode","identifiers":[],"facts":[],"relations":[],"geodeticDatum":"WGS84","class":"Aves","countryCode":"FR","country":"France","rightsHolder":"INBO","minimumDistanceAboveSurfaceInMeters":"87","identifier":"5199:20160101140006","informationWithheld":"see metadata","dynamicProperties":"{\"device_info_serial\":5199, \"catch_location\":\"Kleine Vlakte, Zeebrugge\", \"tracking_started_at\":\"2015-06-03\"}","samplingEffort":"{\"seconds_since_last_occurrence\":1200}","georeferenceVerificationStatus":"unverified","datasetName":"Bird tracking - GPS tracking of Lesser Black-backed Gulls and Herring Gulls breeding at the southern North Sea coast","language":"en","gbifID":"1494057832","occurrenceID":"5199:20160101140006","type":"Event","vernacularName":"Lesser Black-backed Gull","organismID":"L916309","institutionCode":"INBO","organismName":"Imme","ownerInstitutionCode":"INBO/VLIZ/UGent/UA","samplingProtocol":"https://doi.org/10.1007/s10336-012-0908-1","datasetID":"https://doi.org/10.15468/02omly","accessRights":"http://www.inbo.be/en/norms-for-data-use","georeferenceSources":"GPS"},
	{"key":1494058902,"datasetKey":"83e20573-f7dd-4852-9159-21566e1e691e","publishingOrgKey":"1cd669d0-80ea-11de-a9d0-f1765f95f18b","networkKeys":[],"installationKey":"9f25fd85-85dc-4dcd-a1b4-b31165442e2b","publishingCountry":"BE","protocol":"DWC_ARCHIVE","lastCrawled":"2017-10-08T08:55:42.079+0000","lastParsed":"2017-04-27T20:20:28.994+0000","crawlId":49,"extensions":{},"basisOfRecord":"MACHINE_OBSERVATION","sex":"FEMALE","lifeStage":"ADULT","taxonKey":2481174,"kingdomKey":1,"phylumKey":44,"classKey":212,"orderKey":7192402,"familyKey":9316,"genusKey":2481126,"speciesKey":2481174,"scientificName":"Larus fuscus Linnaeus, 1758","kingdom":"Animalia","phylum":"Chordata","order":"Charadriiformes","family":"Laridae","genus":"Larus","species":"Larus fuscus","genericName":"Larus","specificEpithet":"fuscus","taxonRank":"SPECIES","decimalLongitude":-0.12935,"decimalLatitude":48.965328,"coordinateUncertaintyInMeters":8.0,"elevation":0.0,"year":2016,"month":1,"day":1,"eventDate":"2016-01-01T20:02:42.000+0000","issues":["COORDINATE_ROUNDED","COUNTRY_DERIVED_FROM_COORDINATES"],"lastInterpreted":"2018-07-08T00:50:40.989+0000","license":"http://creativecommons.org/publicdomain/zero/1.0/legalcode","identifiers":[],"facts":[],"relations":[],"geodeticDatum":"WGS84","class":"Aves","countryCode":"FR","country":"France","rightsHolder":"INBO","minimumDistanceAboveSurfaceInMeters":"27","identifier":"5199:20160101200242","informationWithheld":"see metadata","dynamicProperties":"{\"device_info_serial\":5199, \"catch_location\":\"Kleine Vlakte, Zeebrugge\", \"tracking_started_at\":\"2015-06-03\"}","samplingEffort":"{\"seconds_since_last_occurrence\":1211}","georeferenceVerificationStatus":"unverified","datasetName":"Bird tracking - GPS tracking of Lesser Black-backed Gulls and Herring Gulls breeding at the southern North Sea coast","language":"en","gbifID":"1494058902","occurrenceID":"5199:20160101200242","type":"Event","vernacularName":"Lesser Black-backed Gull","organismID":"L916309","institutionCode":"INBO","organismName":"Imme","ownerInstitutionCode":"INBO/VLIZ/UGent/UA","samplingProtocol":"https://doi.org/10.1007/s10336-012-0908-1","datasetID":"https://doi.org/10.15468/02omly","accessRights":"http://www.inbo.be/en/norms-for-data-use","georeferenceSources":"GPS"},
	{"key":1494059772,"datasetKey":"83e20573-f7dd-4852-9159-21566e1e691e","publishingOrgKey":"1cd669d0-80ea-11de-a9d0-f1765f95f18b","networkKeys":[],"installationKey":"9f25fd85-85dc-4dcd-a1b4-b31165442e2b","publishingCountry":"BE","protocol":"DWC_ARCHIVE","lastCrawled":"2017-10-08T08:55:42.307+0000","lastParsed":"2017-04-27T20:20:29.376+0000","crawlId":49,"extensions":{},"basisOfRecord":"MACHINE_OBSERVATION","sex":"MALE","lifeStage":"ADULT","taxonKey":2481174,"kingdomKey":1,"phylumKey":44,"classKey":212,"orderKey":7192402,"familyKey":9316,"genusKey":2481126,"speciesKey":2481174,"scientificName":"Larus fuscus Linnaeus, 1758","kingdom":"Animalia","phylum":"Chordata","order":"Charadriiformes","family":"Laridae","genus":"Larus","species":"Larus fuscus","genericName":"Larus","specificEpithet":"fuscus","taxonRank":"SPECIES","decimalLongitude":-9.643028,"decimalLatitude":30.431327,"coordinateUncertaintyInMeters":8.0,"elevation":0.0,"year":2016,"month":1,"day":2,"eventDate":"2016-01-02T06:30:08.000+0000","issues":["COORDINATE_ROUNDED","COUNTRY_DERIVED_FROM_COORDINATES"],"lastInterpreted":"2018-07-08T00:50:40.817+0000","license":"http://creativecommons.org/publicdomain/zero/1.0/legalcode","identifiers":[],"facts":[],"relations":[],"geodeticDatum":"WGS84","class":"Aves","countryCode":"MA","country":"Morocco","rightsHolder":"INBO","minimumDistanceAboveSurfaceInMeters":"146","identifier":"849:20160102063008","informationWithheld":"see metadata","dynamicProperties":"{\"device_info_serial\":849, \"catch_location\":\"APM Terminals, Zeebrugge\", \"tracking_started_at\":\"2013-06-03\"}","samplingEffort":"{\"seconds_since_last_occurrence\":891}","georeferenceVerificationStatus":"unverified","datasetName":"Bird tracking - GPS tracking of Lesser Black-backed Gulls and Herring Gulls breeding at the southern North Sea coast","language":"en","gbifID":"1494059772","occurrenceID":"849:20160102063008","type":"Event","vernacularName":"Lesser Black-backed Gull","organismID":"L901266","institutionCode":"INBO","organismName":"Hilbran","ownerInstitutionCode":"INBO/VLIZ/UGent/UA","samplingProtocol":"https://doi.org/10.1007/s10336-012-0908-1","datasetID":"https://doi.org/10.15468/02omly","accessRights":"http://www.inbo.be/en/norms-for-data-use","georeferenceSources":"GPS"},
	{"key":1494059832,"datasetKey":"83e20573-f7dd-4852-9159-21566e1e691e","publishingOrgKey":"1cd669d0-80ea-11de-a9d0-f1765f95f18b","networkKeys":[],"installationKey":"9f25fd85-85dc-4dcd-a1b4-b31165442e2b","publishingCountry":"BE","protocol":"DWC_ARCHIVE","lastCrawled":"2017-10-08T08:55:42.283+0000","lastParsed":"2017-04-27T20:20:29.405+0000","crawlId":49,"extensions":{},"basisOfRecord":"MACHINE_OBSERVATION","sex":"FEMALE","lifeStage":"ADULT","taxonKey":2481174,"kingdomKey":1,"phylumKey":44,"classKey":212,"orderKey":7192402,"familyKey":9316,"genusKey":2481126,"speciesKey":2481174,"scientificName":"Larus fuscus Linnaeus, 1758","kingdom":"Animalia","phylum":"Chordata","order":"Charadriiformes","family":"Laridae","genus":"Larus","species":"Larus fuscus","genericName":"Larus","specificEpithet":"fuscus","taxonRank":"SPECIES","decimalLongitude":2.558335,"decimalLatitude":36.606869,"coordinateUncertaintyInMeters":4.0,"elevation":0.0,"year":2016,"month":1,"day":2,"eventDate":"2016-01-02T06:11:32.000+0000","issues":["COORDINATE_ROUNDED","COUNTRY_DERIVED_FROM_COORDINATES"],"lastInterpreted":"2018-07-08T00:50:41.488+0000","license":"http://creativecommons.org/publicdomain/zero/1.0/legalcode","identifiers":[],"facts":[],"relations":[],"geodeticDatum":"WGS84","class":"Aves","countryCode":"DZ","country":"Algeria","rightsHolder":"INBO","minimumDistanceAboveSurfaceInMeters":"0","identifier":"5065:20160102061132","informationWithheld":"see metadata","dynamicProperties":"{\"device_info_serial\":5065, \"catch_location\":\"Kleine Vlakte, Zeebrugge\", \"tracking_started_at\":\"2014-06-12\"}","samplingEffort":"{\"seconds_since_last_occurrence\":1234}","georeferenceVerificationStatus":"unverified","datasetName":"Bird tracking - GPS tracking of Lesser Black-backed Gulls and Herring Gulls breeding at the southern North Sea coast","language":"en","gbifID":"1494059832","occurrenceID":"5065:20160102061132","type":"Event","vernacularName":"Lesser Black-backed Gull","organismID":"L907424","institutionCode":"INBO","organismName":"Roxanne","ownerInstitutionCode":"INBO/VLIZ/UGent/UA","samplingProtocol":"https://doi.org/10.1007/s10336-012-0908-1","datasetID":"https://doi.org/10.15468/02omly","accessRights":"http://www.inbo.be/en/norms-for-data-use","georeferenceSources":"GPS"}
],"facets":[]}
`

var noRecsBlob = `{"offset":0,"limit":300,"endOfRecords":true,"count":0,"results":[],"facets":[]}`

func TestCatalog(t *testing.T) {
	testData := []struct {
		data *occurrence
		want string
	}{
		{&occurrence{InstitutionCode: "AM", CollectionCode: "Mammalogy", CatalogNumber: "M.1327"}, "AM:Mammalogy:M.1327"},
		{&occurrence{CollectionCode: "Mammalia", CatalogNumber: "RMNH.MAM.51709"}, "RMNH.MAM.51709"},
		{&occurrence{InstitutionCode: "AMNH", CollectionCode: "AMNH", CatalogNumber: "143466"}, "AMNH:143466"},
		{&occurrence{InstitutionCode: "AMNH", CollectionCode: "Mammals", CatalogNumber: "M-36242"}, "AMNH:Mammals:M-36242"},
		{&occurrence{InstitutionCode: "CIIDIR-IPN", CollectionCode: "CRD", CatalogNumber: "CRD5130"}, "CIIDIR-IPN:CRD5130"},
		{&occurrence{InstitutionCode: "CMN", CollectionCode: "CMNMA", CatalogNumber: "CMNMA 11269"}, "CMNMA 11269"},
		{&occurrence{InstitutionCode: "ICN", CollectionCode: "ICN-MHN-Ma", CatalogNumber: "1373"}, "ICN-MHN-Ma:1373"},
		{&occurrence{InstitutionCode: "INIREB", CollectionCode: "INIREB", CatalogNumber: "185"}, "INIREB:185"},
		{&occurrence{InstitutionCode: "ISM", CollectionCode: "ISM-Mammals", CatalogNumber: "686867"}, "ISM-Mammals:686867"},
		{&occurrence{InstitutionCode: "KPM", CollectionCode: "NF1", CatalogNumber: "KPM-NF1001895"}, "KPM-NF1001895"},
		{&occurrence{InstitutionCode: "MHNG", CollectionCode: "Mammals housed at MHNG, Geneva", CatalogNumber: "MHNG-MAM-1112.098"}, "MHNG-MAM-1112.098"},
		{&occurrence{InstitutionCode: "fcnym.unlp.edu.ar", CollectionCode: "mastozoologia", CatalogNumber: "292"}, "MLP:mastozoologia:292"},
		{&occurrence{InstitutionCode: "Instituto de Investigación de Recursos Biológicos Alexander von Humboldt (IAvH)", CollectionCode: "IAvH-M", CatalogNumber: "IAvH-M-1114"}, "IAvH-M-1114"},
		{&occurrence{InstitutionCode: "Pontificia Universidad Javeriana (PUJ)", CollectionCode: "MPUJ_MAMM", CatalogNumber: "1217"}, "PUJ:MPUJ_MAMM:1217"},
		{&occurrence{InstitutionCode: "Royal Ontario Museum: ROM", CollectionCode: "Vertebrate Paleontology", CatalogNumber: "0390"}, "ROM:Vertebrate Paleontology:0390"},
		{&occurrence{InstitutionCode: "Universidad del Valle", CollectionCode: "UV", CatalogNumber: "ColeccMamif:UV:7644"}, "UV:ColeccMamif:UV:7644"},
		{&occurrence{InstitutionCode: "Universidad Industrial de Santander (UIS)", CollectionCode: "MHN-UIS", CatalogNumber: "UIS-MHN-M-0132"}, "UIS-MHN-M-0132"},
		{&occurrence{InstitutionCode: "University of Alberta Museums", CollectionCode: "UAMZ", CatalogNumber: "M10837"}, "UAMZ:M10837"},
		{&occurrence{InstitutionCode: "University of Guelph, Centre for Biodiversity Genomics", CollectionCode: "BIOUG", CatalogNumber: "ABMC260-05"}, "BIOUG:ABMC260-05"},
	}

	for _, d := range testData {
		if v := d.data.Value(biodv.RecCatalog); v != d.want {
			t.Errorf("catalog %q, want %q", v, d.want)
		}
	}
}
