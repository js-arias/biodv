// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package main

import (
	// initialize records sub-commands
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/add"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/dbadd"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/info"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/mapcmd"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/set"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/table"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/validate"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/records/value"
)
