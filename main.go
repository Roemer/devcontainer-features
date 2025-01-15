package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/roemer/gotaskr"
	"github.com/roemer/gotaskr/execr"
	"github.com/roemer/gotaskr/gttools"
)

////////////////////////////////////////////////////////////
// Main
////////////////////////////////////////////////////////////

func main() {
	os.Exit(gotaskr.Execute())
}

////////////////////////////////////////////////////////////
// Initialize Tasks
////////////////////////////////////////////////////////////

func init() {
	gotaskr.Task("publish-gonovate", func() error {
		// Build the installer inside the feature
		binaryName := "installer"
		featurePath := "features/src/gonovate"
		cmd := exec.Command("go", "build", "-o", binaryName, "-ldflags", "-w", ".")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GOOS=linux")
		cmd.Env = append(cmd.Env, "GOARCH=amd64")
		cmd.Dir = featurePath
		if err := execr.RunCommand(true, cmd); err != nil {
			return err
		}
		defer os.Remove(filepath.Join(featurePath, binaryName))
		// Build and publish the feature

		name := "roemer"
		token := "xxx"
		os.Setenv("DEVCONTAINERS_OCI_AUTH", fmt.Sprintf("ghcr.io|%s|%s", name, token))

		settings := &gttools.DevContainerCliFeaturesPublishSettings{
			Target:    featurePath,
			Registry:  "ghcr.io",
			Namespace: "roemer/devcontainer-features",
		}
		settings.OutputToConsole = true
		return gotaskr.Tools.DevContainerCli.FeaturesPublish(settings)
	})
}
