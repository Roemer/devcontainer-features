package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/roemer/gonovate"
	"github.com/roemer/gonovate/pkg/common"
	"github.com/roemer/gotaskr/execr"
	"github.com/roemer/gover"
)

var versionRegex *regexp.Regexp = regexp.MustCompile(`(?m:)^v(\d+).(\d+)\.(\d+)$`)

func main() {
	if err := runMain(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runMain() error {
	// Handle the flags
	version := flag.String("version", "latest", "The version of Gonovate to install.")
	flag.Parse()
	requestedVersionString := *version

	// Get the possible versions
	allVersions, err := getAllVersions()
	if err != nil {
		return err
	}

	// Calculate the version to install
	var versionToInstall *gover.Version
	if requestedVersionString == "latest" {
		versionToInstall = gover.FindMax(allVersions, gover.EmptyVersion, true)
	} else {
		// Get the max according to the reference version
		referenceVersion := gover.ParseSimple(strings.Split(strings.ReplaceAll(requestedVersionString, "-", "."), "."))
		versionToInstall = gover.FindMax(allVersions, referenceVersion, false)
	}

	// Error if no version was found
	if versionToInstall == nil {
		return fmt.Errorf("no version to install found for '%s'", requestedVersionString)
	}

	// Version found, install it
	fmt.Printf("Installing version %s\n", versionToInstall.Raw)
	if err := installVersion(versionToInstall); err != nil {
		return err
	}

	return nil
}

func getAllVersions() ([]*gover.Version, error) {
	// Get the datasource
	ds, err := gonovate.GetDatasource(common.DATASOURCE_TYPE_GITHUB_TAGS, &common.DatasourceSettings{Logger: slog.Default()})
	if err != nil {
		return nil, nil
	}
	// Get the releases
	releases, err := ds.GetReleases(&common.Dependency{Name: "roemer/gonovate"})
	if err != nil {
		return nil, nil
	}
	// Convert the releases to versions
	allVersions := []*gover.Version{}
	for _, release := range releases {
		allVersions = append(allVersions, gover.MustParseVersionFromRegex(release.VersionString, versionRegex))
	}
	// Return the version list
	return allVersions, nil
}

func installVersion(version *gover.Version) error {
	// Download
	localFileName := "gonovate.zip"
	downloadUrl := fmt.Sprintf("https://github.com/Roemer/gonovate/releases/download/%s/gonovate-linux-%s-amd64.zip", version.Raw, version.CoreVersion())
	downloadCmd := exec.Command("curl", "-L", "-o", localFileName, downloadUrl)
	if err := execr.RunCommand(true, downloadCmd); err != nil {
		return err
	}
	// Extract
	extractCmd := exec.Command("unzip", localFileName)
	if err := execr.RunCommand(true, extractCmd); err != nil {
		return err
	}
	// Install
	configureCmd := exec.Command("install", "-m", "0755", "gonovate", "/usr/local/bin/gonovate")
	if err := execr.RunCommand(true, configureCmd); err != nil {
		return err
	}
	// Cleanup
	if err := os.Remove(localFileName); err != nil {
		return err
	}
	if err := os.Remove("gonovate"); err != nil {
		return err
	}
	return nil
}
