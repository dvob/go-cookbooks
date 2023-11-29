package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// settings
	var (
		// settings for built-in flag types
		myString   = "my string default"
		myBool     = false
		myInt      = 42
		myFloat    = 1.0
		myDuration = time.Second * 2

		// settings for types which implement TextMarshaler/TextUnmarshaler
		myLogLevel = slog.LevelInfo
		myTime     = time.Date(2023, time.January, 1, 0, 0, 0, 0, time.Local)
		myRegex    = regexp.MustCompile("default pattern")
		myIpPrefix = netip.MustParsePrefix("192.168.0.1/24")

		// flags with custom parse function
		myURL *url.URL

		// custom flag values for example slices or maps
		myMapping = map[string]string{
			"key1": "value1",
			"key2": "value2",
		}

		myItems = []string{
			"one",
			"two",
		}
	)

	// built-in flag types
	flag.StringVar(&myString, "string", myString, "description for my string")
	flag.BoolVar(&myBool, "bool", myBool, "description for my bool")
	flag.IntVar(&myInt, "int", myInt, "description for my int")
	flag.Float64Var(&myFloat, "float", myFloat, "description for my float")
	flag.DurationVar(&myDuration, "duration", myDuration, "description for my float")

	// flags for types which implement TextMarshaler/TextUnmarshaler
	flag.TextVar(&myLogLevel, "log-level", &myLogLevel, "log level (DEBUG, INFO, WARN, ERROR)")
	flag.TextVar(&myTime, "time", &myTime, "description for my time")
	flag.TextVar(myRegex, "regex", myRegex, "description for my regex")
	flag.TextVar(&myIpPrefix, "ip-prefix", &myIpPrefix, "description for my ip prefix")

	// flags with custom parse function
	flag.Func("url", "description for my url", func(s string) error {
		u, err := url.Parse(s)
		if err != nil {
			return err
		}
		myURL = u
		return nil
	})

	// flags with custom type which implement the flag.Value interface
	flag.Var(newMapValue(myMapping, "=", ","), "map", "my mapping")
	flag.Var(newSliceValue(&myItems, ","), "item", "my items")

	err := readFlagsFromEnv(flag.CommandLine, "MY_APP_")
	if err != nil {
		return err
	}

	flag.Parse()

	fmt.Printf("string: %s\n", myString)
	fmt.Printf("bool: %t\n", myBool)
	fmt.Printf("int: %d\n", myInt)
	fmt.Printf("float: %f\n", myFloat)
	fmt.Printf("duration: %s\n", myDuration)
	fmt.Printf("log-level: %s\n", myLogLevel)
	fmt.Printf("time: %s\n", myTime)
	fmt.Printf("regex: %s\n", myRegex)
	fmt.Printf("ip-prefix: %s\n", myIpPrefix)
	fmt.Printf("url: %s\n", myURL)

	out, _ := json.MarshalIndent(myMapping, "", "  ")
	fmt.Printf("mapping: \n%s\n", out)

	out, _ = json.MarshalIndent(myItems, "", "  ")
	fmt.Printf("items: \n%s\n", out)
	return nil
}

// readFlagsFromEnv reads configuration values from environment. This function
// can be called before flag.Parse() to read settings from the environment but
// allow to override settings by using flags.
func readFlagsFromEnv(fs *flag.FlagSet, prefix string) error {
	errs := []error{}
	fs.VisitAll(func(f *flag.Flag) {
		envVarName := prefix + f.Name
		envVarName = strings.ReplaceAll(envVarName, "-", "_")
		envVarName = strings.ToUpper(envVarName)
		val, ok := os.LookupEnv(envVarName)
		if !ok {
			return
		}
		err := f.Value.Set(val)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid value '%s' in %s: %w", val, envVarName, err))
		}
	})
	return errors.Join(errs...)
}
