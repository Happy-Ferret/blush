package cmd_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/arsham/blush/blush"
	"github.com/arsham/blush/cmd"
)

var leaveMeHere = "LEAVEMEHERE"

type stdFile struct {
	f *os.File
}

func (s *stdFile) String() string {
	s.f.Seek(0, 0)
	buf := new(bytes.Buffer)
	buf.ReadFrom(s.f)
	return buf.String()
}

func newStdFile(t *testing.T, name string) (*stdFile, func()) {
	f, err := ioutil.TempFile("", name)
	if err != nil {
		t.Fatal(err)
	}
	sf := &stdFile{f}
	return sf, func() {
		f.Close()
		os.Remove(f.Name())
	}
}

func setup(t *testing.T, args string) (stdout, stderr *stdFile, cleanup func()) {
	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldFatalErr := cmd.FatalErr

	stdout, outCleanup := newStdFile(t, "stdout")
	stderr, errCleanup := newStdFile(t, "stderr")
	os.Stdout = stdout.f
	os.Stderr = stderr.f

	os.Args = []string{"blush"}
	if len(args) > 1 {
		os.Args = append(os.Args, strings.Split(args, " ")...)
	}
	cmd.FatalErr = func(s string) {
		fmt.Fprintf(os.Stderr, "%s\n", s)
	}

	cleanup = func() {
		outCleanup()
		errCleanup()
		os.Args = oldArgs
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		cmd.FatalErr = oldFatalErr
	}
	return stdout, stderr, cleanup
}

func TestMainNoArgs(t *testing.T) {
	stdout, stderr, cleanup := setup(t, "")
	defer cleanup()
	cmd.Main()
	if len(stdout.String()) > 0 {
		t.Errorf("didn't expect any stdout, got: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), cmd.ErrNoInput.Error()) {
		t.Errorf("stderr = `%s`, want `%s` in it", stderr.String(), cmd.ErrNoInput.Error())
	}
}

func TestPipeInput(t *testing.T) {
	oldStdin := os.Stdin
	stdout, stderr, cleanup := setup(t, "findme")
	defer func() {
		cleanup()
		os.Stdin = oldStdin
	}()
	file, err := ioutil.TempFile("", "blush_pipe")
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()
	rmFile := func() {
		if err := os.Remove(name); err != nil {
			t.Error(err)
		}
	}
	defer rmFile()
	file.Close()
	rmFile()
	file, err = os.OpenFile(name, os.O_CREATE|os.O_RDWR, os.ModeCharDevice|os.ModeDevice)
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString("you can findme here")
	os.Stdin = file
	file.Seek(0, 0)
	cmd.Main()
	if len(stderr.String()) > 0 {
		t.Errorf("didn't expect printing anything: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "findme") {
		t.Errorf("stdout = `%s`, want `%s` in it", stdout.String(), "findme")
	}
}

func TestMainMatchExact(t *testing.T) {
	match := blush.Colourise("TOKEN", blush.FgBlue)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "../blush/testdata")

	stdout, stderr, cleanup := setup(t, "-b TOKEN "+location)
	defer cleanup()
	cmd.Main()

	if len(stderr.String()) > 0 {
		t.Errorf("len(stderr.String()) = %d, want 0: `%s`", len(stderr.String()), stderr.String())
	}
	if len(stdout.String()) == 0 {
		t.Errorf("len(stdout.String()) = %d, want > 0", len(stdout.String()))
	}
	if !strings.Contains(stdout.String(), match) {
		t.Errorf("want `%s` in `%s`", match, stdout.String())
	}
}

func TestMainMatchIExact(t *testing.T) {
	match := blush.Colourise("TOKEN", blush.FgBlue)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "../blush/testdata")

	stdout, stderr, cleanup := setup(t, "-i -b token "+location)
	defer cleanup()
	cmd.Main()

	if len(stderr.String()) > 0 {
		t.Errorf("len(stderr.String()) = %d, want 0: `%s`", len(stderr.String()), stderr.String())
	}
	if len(stdout.String()) == 0 {
		t.Errorf("len(stdout.String()) = %d, want > 0", len(stdout.String()))
	}
	if !strings.Contains(stdout.String(), match) {
		t.Errorf("want `%s` in `%s`", match, stdout.String())
	}
}

func TestMainMatchRegexp(t *testing.T) {
	match := blush.Colourise("TOKEN", blush.FgBlue)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "../blush/testdata")

	stdout, stderr, cleanup := setup(t, `-b TOK[EN]{2} `+location)
	defer cleanup()
	cmd.Main()

	if len(stderr.String()) > 0 {
		t.Errorf("len(stderr.String()) = %d, want 0: `%s`", len(stderr.String()), stderr.String())
	}
	if len(stdout.String()) == 0 {
		t.Errorf("len(stdout.String()) = %d, want > 0", len(stdout.String()))
	}
	if !strings.Contains(stdout.String(), match) {
		t.Errorf("want `%s` in `%s`", match, stdout.String())
	}
}

func TestMainMatchRegexpInsensitive(t *testing.T) {
	match := blush.Colourise("TOKEN", blush.FgBlue)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "../blush/testdata")

	stdout, stderr, cleanup := setup(t, `-i -b tok[en]{2} `+location)
	defer cleanup()
	cmd.Main()

	if len(stderr.String()) > 0 {
		t.Errorf("len(stderr.String()) = %d, want 0: `%s`", len(stderr.String()), stderr.String())
	}
	if len(stdout.String()) == 0 {
		t.Errorf("len(stdout.String()) = %d, want > 0", len(stdout.String()))
	}
	if !strings.Contains(stdout.String(), match) {
		t.Errorf("want `%s` in `%s`", match, stdout.String())
	}
}

func TestMainMatchNoCut(t *testing.T) {
	matches := []string{"TOKEN", "ONE", "TWO", "THREE", "FOUR"}
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "../blush/testdata")

	stdout, stderr, cleanup := setup(t, fmt.Sprintf("-C -b %s %s", leaveMeHere, location))
	defer cleanup()
	cmd.Main()

	if len(stderr.String()) > 0 {
		t.Errorf("len(stderr.String()) = %d, want 0: `%s`", len(stderr.String()), stderr.String())
	}
	if len(stdout.String()) == 0 {
		t.Errorf("len(stdout.String()) = %d, want > 0", len(stdout.String()))
	}
	for _, s := range matches {
		if !strings.Contains(stdout.String(), s) {
			t.Errorf("want `%s` in `%s`", s, stdout.String())
		}
	}
}

func TestNoFiles(t *testing.T) {
	fileName := "test"
	b, err := cmd.GetBlush([]string{fileName})
	if err == nil {
		t.Error("GetBlush(): err = nil, want error")
	}
	if b != nil {
		t.Errorf("GetBlush(): b = %v, want nil", b)
	}
}

func TestColourArgs(t *testing.T) {
	aaa := "aaa"
	bbb := "bbb"
	tcs := []struct {
		name  string
		input []string
		want  []blush.Locator
	}{
		{"empty", []string{"/"}, []blush.Locator{}},
		{"1-no colour", []string{"aaa", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.DefaultColour),
		}},
		{"1-colour", []string{"-b", "aaa", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgBlue),
		}},
		{"1-colour long", []string{"--blue", "aaa", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgBlue),
		}},
		{"2-no colour", []string{"aaa", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.DefaultColour),
			blush.NewExact(bbb, blush.DefaultColour),
		}},
		{"2-colour", []string{"-b", "aaa", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgBlue),
			blush.NewExact(bbb, blush.FgBlue),
		}},
		{"2-two colours", []string{"-b", "aaa", "-g", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgBlue),
			blush.NewExact(bbb, blush.FgGreen),
		}},
		{"red", []string{"-r", "aaa", "--red", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgRed),
			blush.NewExact(bbb, blush.FgRed),
		}},
		{"green", []string{"-g", "aaa", "--green", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgGreen),
			blush.NewExact(bbb, blush.FgGreen),
		}},
		{"blue", []string{"-b", "aaa", "--blue", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgBlue),
			blush.NewExact(bbb, blush.FgBlue),
		}},
		{"white", []string{"-w", "aaa", "--white", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgWhite),
			blush.NewExact(bbb, blush.FgWhite),
		}},
		{"black", []string{"-bl", "aaa", "--black", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgBlack),
			blush.NewExact(bbb, blush.FgBlack),
		}},
		{"cyan", []string{"-cy", "aaa", "--cyan", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgCyan),
			blush.NewExact(bbb, blush.FgCyan),
		}},
		{"magenta", []string{"-mg", "aaa", "--magenta", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgMagenta),
			blush.NewExact(bbb, blush.FgMagenta),
		}},
		{"yellow", []string{"-yl", "aaa", "--yellow", "bbb", "/"}, []blush.Locator{
			blush.NewExact(aaa, blush.FgYellow),
			blush.NewExact(bbb, blush.FgYellow),
		}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			input := append([]string{"blush"}, tc.input...)
			b, err := cmd.GetBlush(input)
			if err != nil {
				t.Errorf("GetBlush(): err = %s, want nil", err)
			}
			if b == nil {
				t.Error("GetBlush(): b = nil, want *Blush")
			}
			if !argsEqual(b.Locator, tc.want) {
				t.Errorf("(%s): b.Args = %v, want %v", tc.input, b.Locator, tc.want)
			}
		})
	}
}

func argsEqual(a, b []blush.Locator) bool {
	isIn := func(a blush.Locator, haystack []blush.Locator) bool {
		for _, b := range haystack {
			if reflect.DeepEqual(a, b) {
				return true
			}
		}
		return false
	}

	for _, item := range b {
		if !isIn(item, a) {
			return false
		}
	}
	return true
}
