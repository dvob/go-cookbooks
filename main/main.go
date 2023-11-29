package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var (
		mySetting = "my default value"
	)

	flag.StringVar(&mySetting, "my-setting", mySetting, "description of my setting")

	err := readFlagsFromEnv(flag.CommandLine, "MY_APP_")
	if err != nil {
		return err
	}

	flag.Parse()

	// we explicitly set the default logger that it is set for the old log package as well
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	// your main here
	return nil
}

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
