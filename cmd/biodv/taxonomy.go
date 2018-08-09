// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package main

import (
	// initialize taxonomy sub-commands
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/add"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/dbfill"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/dbsync"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/dbupdate"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/info"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/list"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/set"
	_ "github.com/js-arias/biodv/cmd/biodv/internal/taxonomy/val"
)
