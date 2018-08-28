// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package gbif

// MusAcronyms contains some museum acronyms
// that are read as full text in GBIF.
// This list is incomplete,
// and updated anytime an anomalous strings is discovered.
var musAcronyms = map[string]string{
	"fcnym.unlp.edu.ar": "MLP",
	"Instituto de Investigación de Recursos Biológicos Alexander von Humboldt (IAvH)": "IAvH",
	"Pontificia Universidad Javeriana (PUJ)":                                          "PUJ",
	"Royal Ontario Museum: ROM":                                                       "ROM",
	"Universidad del Valle":                                                           "UV",
	"Universidad Industrial de Santander (UIS)":                                       "UIS",
	"University of Alberta Museums":                                                   "UA",
	"University of Guelph, Centre for Biodiversity Genomics":                          "BIOUG",
}
