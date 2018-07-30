// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// This work is derived from the go tool source code
// Copyright 2011 The Go Authors.  All rights reserved.

package cmdapp

import "testing"

func TestCommandName(t *testing.T) {
	cases := []struct {
		c    *Command
		want string
	}{
		{&Command{UsageLine: "name this is a command"}, "name"},
		{&Command{UsageLine: "single"}, "single"},
		{&Command{UsageLine: " empty command entry"}, ""},
	}

	for _, c := range cases {
		got := c.c.Name()
		if got != c.want {
			t.Errorf("command name on usageline %q: %q, want %q", c.c.UsageLine, got, c.want)
		}
	}
}

func TestAddCommand(t *testing.T) {
	cases := []*Command{
		&Command{UsageLine: "first command"},
		&Command{UsageLine: "second"},
		&Command{UsageLine: "THIRD"},
	}
	for _, c := range cases {
		Add(c)
	}

	for _, c := range cases {
		cmd := getCmd(c.Name())
		if cmd == nil {
			t.Errorf("command %q not found", c.Name())
			continue
		}
		if cmd != c {
			t.Errorf("command %q name = %q", c.Name(), cmd.Name())
		}
	}
	if cmd := getCmd("fourth"); cmd != nil {
		t.Errorf("command %q should be nil", "fourth")
	}
}

func TestAddCommandPanic(t *testing.T) {
	cases := []struct {
		c    *Command
		want string
	}{
		{&Command{UsageLine: "a command", Short: "base name"}, ""},
		{&Command{UsageLine: "a repeat command name", Short: "repeat lower case"}, "repeated name"},
		{&Command{UsageLine: "A repeat command name", Short: "repeat upper case"}, "repeated name"},
		{&Command{UsageLine: " empty command name"}, "empty command"},
	}
	for i, c := range cases {
		if i == 0 {
			Add(c.c)
			continue
		}
		func(cmd *Command, want string) {
			defer func() {
				if p := recover(); p == nil {
					t.Errorf("for command %q expecting %q panic", c.c.Name(), want)
				}
			}()
			Add(cmd)
		}(c.c, c.want)
	}
}

func TestCapitalize(t *testing.T) {
	cases := []struct {
		s    string
		want string
	}{
		{"name", "Name"},
		{"NAME", "NAME"},
		{"", ""},
		{"    ", ""},
	}

	for _, c := range cases {
		got := capitalize(c.s)
		if got != c.want {
			t.Errorf("capitalized name %q: want %q", got, c.want)
		}
	}
}

func TestRunnable(t *testing.T) {
	mock := func(*Command, []string) error {
		return nil
	}
	cases := []struct {
		c    *Command
		want bool
	}{
		{&Command{UsageLine: "runnable", Run: mock}, true},
		{&Command{UsageLine: "topic"}, false},
	}

	for _, c := range cases {
		if c.c.runnable() != c.want {
			t.Errorf("%q command runnable() is %v", c.c.Name(), c.c.runnable())
		}
	}
}
