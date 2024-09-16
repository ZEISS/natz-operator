package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/mod/semver"
	"helm.sh/helm/pkg/chartutil"
)

type flags struct {
	File    string
	Version string
}

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	f := &flags{}

	pflag.StringVar(&f.File, "file", f.File, "chart")
	pflag.StringVar(&f.Version, "version", f.Version, "version")
	pflag.Parse()

	ok := semver.IsValid(f.Version)
	if !ok {
		log.Fatal(errors.New("updater: no valid version"))
	}

	f.Version = strings.TrimPrefix(f.Version, "v")

	meta, err := chartutil.LoadChartfile(f.File)
	if err != nil {
		log.Fatal(err)
	}

	meta.AppVersion = f.Version
	meta.Version = f.Version

	err = chartutil.SaveChartfile(f.File, meta)
	if err != nil {
		log.Fatal(err)
	}
}
