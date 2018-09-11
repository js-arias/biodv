// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

package cmdapp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// Inter implements an interactive command line,
// as in the ed unix command.
type Inter struct {
	// Prompt is the function
	// that return the prompt to print
	// when expecting a command.
	prompt func() string

	// R is the reader used
	// for the command input.
	r *bufio.Reader

	mu   sync.Mutex
	cmds map[string]*Cmd
}

// NewInter returns a new command interpreter,
// ready to use.
func NewInter(r io.Reader, prompt func() string) *Inter {
	if r == nil {
		r = os.Stdin
	}
	if prompt == nil {
		prompt = func() string { return "$" }
	}
	i := &Inter{
		r:      bufio.NewReader(r),
		prompt: prompt,
		cmds:   make(map[string]*Cmd),
	}
	hlp := &Cmd{
		Abrev: "h",
		Name:  "help",
		Short: "print command help",
		Long: `
Usage:
    h [<command>]
    help [<command>]
Without parameters, it will print the list of available commands.
If a command is given, it will print the help on that command.
		`,
		Run: helpCmdRun(i),
	}
	i.Add(hlp)
	return i
}

// Add adds a new command to
// the interactive command line.
func (i *Inter) Add(c *Cmd) {
	c.Name = strings.ToLower(c.Name)
	c.Name = strings.Join(strings.Fields(c.Name), "-")
	if c.Name == "" {
		panic("cmdapp: inter: empty command name")
	}
	c.Abrev = strings.ToLower(c.Abrev)
	c.Abrev = strings.Join(strings.Fields(c.Abrev), "-")
	if i.getCmd(c.Name) != nil || i.getCmd(c.Abrev) != nil {
		msg := fmt.Sprintf("cmdapp: inter: repeated command name: %s", c.Name)
		panic(msg)
	}
	if x, _ := utf8.DecodeRuneInString(c.Name); !unicode.IsLetter(x) {
		msg := fmt.Sprintf("cmdapp: inter: invalid command name: %s", c.Name)
		panic(msg)
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	i.cmds[c.Name] = c
	if c.Abrev != "" {
		i.cmds[c.Abrev] = c
	}
}

// GetCmd returns a command with a given name.
func (i *Inter) getCmd(name string) *Cmd {
	name = strings.ToLower(name)
	name = strings.Join(strings.Fields(name), "-")
	if name == "" {
		return nil
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.cmds[name]
}

// A Cmd is a command for
// an interactive command line application.
type Cmd struct {
	// Abrev is an small
	// (usually one letter)
	// invocation of a command.
	Abrev string

	// Name is the full name
	// os the command.
	Name string

	// Short is a short description of the command.
	Short string

	// Long is a long description of the command.
	Long string

	// Run runs the command.
	// It returns true to indicate the end
	// of the command loop.
	Run func(args []string) (bool, error)
}

// Loop is the command loop.
// The loop will end,
// when receive a false from a command,
// or when a io.EOF error is found.
//
// The loop command will ignored empty lines,
// lines with only spaces
// or lines in which the first non space character
// is '#'
// (a handy way to introduce comments).
func (i *Inter) Loop() {
	for {
		fmt.Printf("%s ", i.prompt())
		line, err := i.r.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}

		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}
		cmd := strings.ToLower(args[0])
		args = args[1:]

		c := i.getCmd(cmd)
		if c == nil {
			if x, _ := utf8.DecodeRuneInString(cmd); x == '#' {
				continue
			}
			fmt.Printf("error: unknown command '%s'\n", cmd)
			continue
		}
		end, err := c.Run(args)
		if end {
			return
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}
}

// GetAnswer gets an answer
// that should be responded by the user,
// outside of the command loop.
//
// If empty is true,
// an empty line will be a valid answer.
func (i *Inter) GetAnswer(question string, empty bool) []string {
	for {
		fmt.Printf("%s: ", question)
		ln, err := i.r.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return nil
		}

		ans := strings.Fields(ln)
		if empty {
			return ans
		}
		if len(ans) == 0 {
			continue
		}
		return ans
	}
}

func (i *Inter) printCmds() {
	var cmd []string
	i.mu.Lock()
	defer i.mu.Unlock()
	for j, c := range i.cmds {
		if j != c.Name {
			continue
		}
		cmd = append(cmd, c.Name)
	}
	sort.Strings(cmd)

	fmt.Printf("Commands are:\n")
	for _, nm := range cmd {
		c := i.cmds[nm]
		nm := c.Name
		if c.Abrev != "" {
			nm = fmt.Sprintf("%s, %s", c.Abrev, c.Name)
		}
		fmt.Printf("    %-16s %s\n", nm, c.Short)
	}
	fmt.Printf("\nUse 'h <command>' for more information about a command\n")
}

func helpCmdRun(i *Inter) func(args []string) (bool, error) {
	return func(args []string) (bool, error) {
		if len(args) == 0 {
			i.printCmds()
			return false, nil
		}

		if len(args) != 1 {
			return false, errors.New("to many arguments")
		}

		c := i.getCmd(args[0])
		if c == nil {
			return false, errors.Errorf("unknown command '%s'", args[0])
		}

		fmt.Printf("%s\n", strings.TrimSpace(c.Long))
		return false, nil
	}
}
