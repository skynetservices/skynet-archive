package config

import (
	"flag"
	"os"
	"testing"
)

func clearEnv() {
	os.Setenv("SKYNET_BIND_IP", "")
	os.Setenv("SKYNET_MIN_PORT", "")
	os.Setenv("SKYNET_MAX_PORT", "")
	os.Setenv("SKYNET_REGION", "")
	os.Setenv("SKYNET_VERSION", "")
}

func TestSplitFlagsetFromArgs(t *testing.T) {
	var a, b, c, e, f, g string
	var d, i bool

	flagset1 := flag.NewFlagSet("TestFlagFilter1", flag.ExitOnError)
	flagset1.StringVar(&a, "a", "x", "a")
	flagset1.StringVar(&b, "b", "x", "b")
	flagset1.StringVar(&c, "c", "x", "c")
	flagset1.BoolVar(&d, "d", false, "d")

	flagset2 := flag.NewFlagSet("TestFlagFilter2", flag.ExitOnError)
	flagset2.StringVar(&e, "e", "x", "e")
	flagset2.StringVar(&f, "f", "x", "f")
	flagset2.StringVar(&g, "g", "x", "g")
	flagset2.BoolVar(&i, "i", false, "i")

	args := []string{
		// in flagset
		"--a=a",
		"-b=b",
		"-c=c",
		"-d",
		// not in flagset
		"--e=e",
		"-f=f",
		"-g=g",
		"-i",
	}

	flagsetArgs, additionalArgs := SplitFlagsetFromArgs(flagset1, args)

	// Validate flags
	if len(flagsetArgs) != 4 {
		t.Error("didn't split flagset")
	}

	if flagsetArgs[0] != args[0] {
		t.Error("didn't properly split flagset")
	}

	if flagsetArgs[1] != args[1] {
		t.Error("didn't properly split flagset")
	}

	if flagsetArgs[2] != args[2] {
		t.Error("didn't properly split flagset")
	}

	if flagsetArgs[3] != args[3] {
		t.Error("didn't properly split flagset")
	}

	if flagset1.Parse(flagsetArgs) != nil {
		t.Error("didn't return proper flagset arguments, flags failed to parse")
	}

	if len(flagset1.Args()) > 0 {
		t.Error("unparsed flags returned from parse")
	}

	if len(additionalArgs) != 4 {
		t.Error("didn't return proper additional arguments")
	}

	if additionalArgs[0] != args[4] {
		t.Error("didn't return proper additional arguments")
	}

	if additionalArgs[1] != args[5] {
		t.Error("didn't return proper additional arguments")
	}
	if additionalArgs[2] != args[6] {
		t.Error("didn't return proper additional arguments")
	}
	if additionalArgs[3] != args[7] {
		t.Error("didn't return proper additional arguments")
	}

	if flagset2.Parse(additionalArgs) != nil {
		t.Error("didn't return proper additional arguments, flags failed to parse")
	}

	if len(flagset2.Args()) > 0 {
		t.Error("unparsed flags returned from parse")
	}
}
