package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/joerx/resolve-semver/pkg/semver"
)

const releaseURL = "https://releases.hashicorp.com"

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "fatal: %v", err)
	os.Exit(1)
}

func main() {
	if err := downloadAndRun(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func downloadAndRun(tfargs []string) error {
	dir := "."
	module, diags := tfconfig.LoadModule(dir)
	if diags.HasErrors() {
		return diags.Err()
	}

	if len(module.RequiredCore) != 1 {
		return fmt.Errorf("cannot reliably determine TF core version, found %d possibilities", len(module.RequiredCore))
	}

	constraint := module.RequiredCore[0]
	log.Printf("Trying to find terraform version matching \"%s\"\n", constraint)

	versions, err := fetchTerraformVersions()
	if err != nil {
		return err
	}

	version, err := semver.FindLatestMatching(constraint, versions)
	if err != nil {
		return err
	}

	log.Printf("Using version %s", version)

	binPath, err := ensureTerraformVersion(version)
	if err != nil {
		return err
	}

	if err := executeTerraform(binPath, tfargs); err != nil {
		return err
	}

	return nil
}

func fetchTerraformVersions() ([]string, error) {
	re := regexp.MustCompile(`terraform_([0-9]+.[0-9]+.[0-9]+(-\w+)?)`)
	versions := make([]string, 0)

	resp, err := http.Get(fmt.Sprintf("%s/terraform/", releaseURL))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if m := re.FindAllSubmatch(line, -1); m != nil {
			v := m[0][1]
			versions = append(versions, string(v))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return versions, nil
}

func executeTerraform(binPath string, args []string) error {
	log.Printf("Exec '%s %s'", binPath, strings.Join(args, " "))

	cmd := exec.Command(binPath, args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", stdoutStderr)

	return nil
}

func ensureTerraformVersion(v string) (string, error) {
	dstBase := fmt.Sprintf("/tmp/tfversions/%s/", v)
	binPath := filepath.Join(dstBase, "terraform")

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		log.Printf("Installing terraform %s", v)

		goos := runtime.GOOS
		arch := runtime.GOARCH

		baseURL := fmt.Sprintf("%s/terraform/%s/terraform_%s", releaseURL, v, v)
		shaURL := fmt.Sprintf("%s_SHA256SUMS", baseURL)
		zipURL := fmt.Sprintf("%s_%s_%s.zip", baseURL, goos, arch)

		log.Printf("Destination: %s", binPath)
		log.Printf("Downloading: %s", zipURL)
		log.Printf("Checksums: %s", shaURL)

		if err := getter.GetFile(binPath, zipURL); err != nil {
			return "", err
		}

		return binPath, nil
	}

	return binPath, nil
}
