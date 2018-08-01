// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Biodv is a tool for management and analysis of biodiveristy data.
package main

import (
	"github.com/js-arias/biodv/cmdapp"

	// add commands in each metapackage
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy"

	// load drivers
	_ "github.com/js-arias/biodv/driver/gbif"
)

func main() {
	cmdapp.Short = "Biodv is a tool for management and analysis of biodiveristy data."
	cmdapp.Main()
}
