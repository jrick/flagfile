package flagfile

import (
	"flag"
	"strings"
	"testing"
)

type product struct {
	name  string
	price float64
}

type debug struct {
	enabled bool
	level   int
}

type flagsSection struct {
	bflag bool
	pflag product
	dflag debug
}

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

func newFlagsSections() (*flagsSection, *flag.FlagSet) {
	f := new(flagsSection)
	fs := flag.NewFlagSet("", 0)
	fs.BoolVar(&f.bflag, "b", false, "bool flag")
	fs.StringVar(&f.pflag.name, "product.name", "", "string flag")
	fs.Float64Var(&f.pflag.price, "product.price", 0, "float64 flag")
	fs.BoolVar(&f.dflag.enabled, "debug.enabled", false, "bool flag")
	fs.IntVar(&f.dflag.level, "debug.level", 0, "int var")
	return f, fs
}

func parseFlags(contents string) (*flags, error) {
	f, fs := newFlags()
	var p Parser
	err := p.Parse(strings.NewReader(contents), fs)
	return f, err
}

func parseFlagsSections(contents string) (*flagsSection, error) {
	f, fs := newFlagsSections()
	var p Parser
	p.ParseSections = true
	err := p.Parse(strings.NewReader(contents), fs)
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

func TestSections(t *testing.T) {
	f, err := parseFlagsSections(`
[] ; ignored
b=true

[product]
name=widget
price=5.99

[debug]
enabled=true
level=1
`)
	if err != nil {
		t.Error(err)
	}
	if f.bflag != true || f.pflag.name != "widget" || f.pflag.price != 5.99 || f.dflag.enabled != true ||
		f.dflag.level != 1 {
		t.Errorf("parsing sections produced unexpected result: %+v", f)
	}
}

func TestUnknown(t *testing.T) {
	const cfg = `
unknown=value
`
	f, err := parseFlags(cfg)
	if err == nil {
		t.Errorf("expected parse error")
	}
	if *f != (flags{}) {
		t.Errorf("flags were unexpectedly parsed: %+v", f)
	}

	f, fs := newFlags()
	var p Parser
	p.AllowUnknown = true
	err = p.Parse(strings.NewReader(cfg), fs)
	if err != nil {
		t.Errorf("parsing errored on unknown option with AllowUnknown=true: %v", err)
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
