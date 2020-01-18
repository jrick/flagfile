package flagfile

import (
	"bufio"
	"errors"
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

// Parse parses an io.Reader for configuration of a flag.FlagSet.
// The reader must contain newline-delimited name=value pairs for each set flag.
// Comments begin at any # or ; character and whitespace is trimmed.
// INI sections (in the form [Section Name]) and empty lines are ignored.
func Parse(r io.Reader, fs *flag.FlagSet) (err error) {
	line := 0
	defer func() {
		if err != nil {
			err = &Error{Line: line, Err: err}
		}
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
		// Ignore INI sections
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
			return err
		}
	}
	return scanner.Err()
}

type config struct {
	fs *flag.FlagSet
}

// ConfigFlag returns a flag.Value which, when set, parses fs with a config file
// read from the set file path.
func ConfigFlag(fs *flag.FlagSet) flag.Value {
	return &config{fs}
}

func (c *config) String() string { return "" }

func (c *config) Set(value string) (err error) {
	fi, err := os.Open(value)
	if err != nil {
		return err
	}
	defer fi.Close()
	err = Parse(fi, c.fs)
	var e *Error
	if errors.As(err, &e) {
		e.File = fi.Name()
	}
	return err
}
