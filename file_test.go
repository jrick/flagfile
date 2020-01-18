package flagfile

import (
	"flag"
	"strings"
	"testing"
)

type flags struct {
	bflag bool
	uflag uint
	iflag int
	sflag string
}

func newFlags() (*flags, *flag.FlagSet) {
	f := new(flags)
	fs := flag.NewFlagSet("", 0)
	fs.BoolVar(&f.bflag, "b", false, "bool flag")
	fs.UintVar(&f.uflag, "u", 0, "uint flag")
	fs.IntVar(&f.iflag, "i", 0, "int flag")
	fs.StringVar(&f.sflag, "s", "", "string flag")
	return f, fs
}

func parseFlags(contents string) (*flags, error) {
	f, fs := newFlags()
	err := Parse(strings.NewReader(contents), fs)
	return f, err
}

func TestSet(t *testing.T) {
	f, err := parseFlags(`
[Ignored Section]
# comment
; comment
b=true
u = 123 ; spaces are allowed
i=-123
s=abc
`)
	if err != nil {
		t.Error(err)
	}
	if f.bflag != true || f.uflag != 123 || f.iflag != -123 || f.sflag != "abc" {
		t.Errorf("parsing produced unexpected result: %+v", f)
	}
}

func TestUnknown(t *testing.T) {
	f, err := parseFlags(`
unknown=value
`)
	if err == nil {
		t.Errorf("expected parse error")
	}
	if *f != (flags{}) {
		t.Errorf("flags were unexpectedly parsed: %+v", f)
	}
}

func TestUnsetBool(t *testing.T) {
	f, fs := newFlags()
	f.bflag = true
	err := Parse(strings.NewReader(`
b=false
`), fs)
	if err != nil {
		t.Error(err)
	}
	if f.bflag {
		t.Errorf("bool flag was not unset from default")
	}
}

func TestBoolZero(t *testing.T) {
	f, fs := newFlags()
	f.bflag = true
	err := Parse(strings.NewReader(`
b=0
`), fs)
	if err != nil {
		t.Error(err)
	}
	if f.bflag {
		t.Errorf("bool flag was not unset from default (with b=0)")
	}
}

func TestBoolOne(t *testing.T) {
	f, fs := newFlags()
	err := Parse(strings.NewReader(`
b=1
`), fs)
	if err != nil {
		t.Error(err)
	}
	if !f.bflag {
		t.Errorf("bool flag was not set (with b=1)")
	}
}

func TestMissingEquals(t *testing.T) {
	_, err := parseFlags(`
b
`)
	if err == nil {
		t.Fatal("parsing should fail with missing equals")
	}
	if !strings.Contains(err.Error(), `"b"`) {
		t.Error("missing equals error does not report bad string")
	}
}
