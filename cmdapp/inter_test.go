// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package cmdapp

import "testing"

func TestAddCmd(t *testing.T) {
	cases := []*Cmd{
		{Name: "NAME"},
		{Name: "composite name"},
		{Name: "   name with spaces "},
		{Name: "ok"},
	}
	i := NewInter(nil)
	for _, c := range cases {
		i.Add(c)
	}

	for _, c := range cases {
		cmd := i.getCmd(c.Name)
		if cmd == nil {
			t.Errorf("command %q not found", c.Name)
			continue
		}
		if cmd != c {
			t.Errorf("command %q name = %q", c.Name, cmd.Name)
		}
	}

	if i.getCmd("no-command") != nil {
		t.Errorf("command %q should be nil", "no-command")
	}
}

func TestAddCmdPanic(t *testing.T) {
	cases := []*Cmd{
		{Name: "name"},
		{Name: "name"},
		{Name: "NAME"},
		{Name: "   "},
		{},
		{Name: "# a comment"},
		{Name: "1 start with a number"},
	}

	i := NewInter(nil)
	for j, c := range cases {
		if j == 0 {
			i.Add(c)
			continue
		}
		func(cmd *Cmd) {
			defer func() {
				if p := recover(); p == nil {
					t.Errorf("for command %q expecting a panic", c.Name)
				}
			}()
			i.Add(cmd)
		}(c)
	}
}
