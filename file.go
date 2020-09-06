package flagfile

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Error reports the filename and line number where parsing errors occured.
type Error struct {
	File string
	Line int
	Err  error
}

func (e *Error) Error() string {
	if e.File == "" {
		return fmt.Sprintf("flagfile: line %d: %v", e.Line, e.Err)
	}
	return fmt.Sprintf("%v:%d: %v", e.File, e.Line, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// Parser provides parsing of INI-like config files to modify values in a
// flag.FlagSet.  Parser behavior is configured by its struct fields.
type Parser struct {
	// AllowUnknown specifies whether Parse should skip unknown flag errors,
	// or return these errors to the caller.
	AllowUnknown bool
}

// Parse parses an io.Reader for configuration of a flag.FlagSet.
// The reader must contain newline-delimited name=value pairs for each set flag.
// Comments begin at any # or ; character and whitespace is trimmed.
// INI section headers (in the form [Section Name]) and empty lines are ignored.
func (p *Parser) Parse(r io.Reader, fs *flag.FlagSet) (err error) {
	line := 0
	defer func() {
		if err == nil {
			return
		}
		var file string
		if fi, ok := r.(*os.File); ok {
			file = fi.Name()
		}
		err = &Error{File: file, Line: line, Err: err}
	}()
	scanner := bufio.NewScanner(r)
	for ; scanner.Scan(); line++ {
		s := scanner.Text()
		comment := strings.IndexAny(s, "#;")
		if comment != -1 {
			s = s[:comment]
		}
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		// Ignore INI section headers
		if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
			continue
		}
		equals := strings.IndexByte(s, '=')
		if equals == -1 {
			return fmt.Errorf("parse error: %q", s)
		}
		k := strings.TrimSpace(s[:equals])
		v := strings.TrimSpace(s[equals+1:])
		err := fs.Set(k, v)
		if err != nil {
			// The flag package returns fmt.Errorf for unrecognized
			// flags, making this the best detection possible.
			if p.AllowUnknown && strings.HasPrefix(err.Error(), "no such flag -") {
				continue
			}
			return err
		}
	}
	return scanner.Err()
}

// ConfigFlag returns a flag.Value which, when set, parses fs with a config file
// read from the set file path.  Parse options are specified through p.
func (p *Parser) ConfigFlag(fs *flag.FlagSet) flag.Value {
	return &config{p, fs}
}

var defaultParser Parser

// Parse uses the default Parser to parse an io.Reader for configuration of a flag.FlagSet.
// The reader must contain newline-delimited name=value pairs for each set flag.
// Comments begin at any # or ; character and whitespace is trimmed.
// INI section headers (in the form [Section Name]) and empty lines are ignored.
func Parse(r io.Reader, fs *flag.FlagSet) (err error) {
	return defaultParser.Parse(r, fs)
}

type config struct {
	p  *Parser
	fs *flag.FlagSet
}

// ConfigFlag returns a flag.Value which, when set, parses fs with a config file
// read from the set file path.  The default (zero-value) parser config is used.
func ConfigFlag(fs *flag.FlagSet) flag.Value {
	return &config{&defaultParser, fs}
}

func (c *config) String() string { return "" }

func (c *config) Set(value string) (err error) {
	fi, err := os.Open(value)
	if err != nil {
		return err
	}
	defer fi.Close()
	return c.p.Parse(fi, c.fs)
}
