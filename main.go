package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeclysm/extract"
	"github.com/posener/complete/v2"
	"github.com/posener/complete/v2/predict"
)

const (
	godir        = "/usr/local/go"
	goSymLink    = "/usr/local/go/active"
	activeDir    = "active"
	instructions = `[WARN] go path is not set in PATH. Add the following to ~/.zshrc file
	export GO_BIN="/usr/local/go/active/bin"
	export PATH="${PATH}:${GO_BIN}"`
	defaultPlatform     = "darwin"
	defaultArch         = "arm64"
	gosourceURL         = "https://go.dev/dl/go%s.%s-%s.tar.gz"
	boldString          = "\033[1m%s\033[0m"
	activeVersionNotice = "\n[ERROR] you are trying to delete active version of go. Change the active version before deleting\n"
)

const (
	cmdUsageNotice = `Usage of %s:
download:
  -arch string
    	os architecture (default "arm64")
  -platform string
    	os platform (default "darwin")
  -version string
    	version to be downloaded

list:
  List the installed go versions

delete <version>: Delete an installed go version
  <version> is the version number to be deleted

set <version>: Set the active go version
  <version> is the version number to be activated
`
)

var (
	availablePlatforms = []string{"darwin", "linux", "freebsd"}
	availableArch      = []string{"amd64", "arm64", "386", "armv6l", "ppc64le", "s390x"}
	letters            = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func deleteVersion(version string) {
	activeVersion, _ := getActiveVersion()
	if activeVersion == version {
		fmt.Print(activeVersionNotice)
		os.Exit(1)
	}

	versions, _ := versionList()
	var found bool
	for _, v := range versions {
		if v == version {
			found = true
		}
	}
	if !found {
		fmt.Printf("\n[ERROR] version not found\n")
		os.Exit(1)
	}

	if err := os.RemoveAll(godir + "/go" + version); err != nil {
		fmt.Printf("\n[ERROR] failed to delete directory: %v\n", err)
		os.Exit(1)
	}
}

func setVersion(version string) {
	versions, errList := versionList()
	if errList != nil {
		fmt.Printf("\n[ERROR] failed to set version: %v", errList)
		os.Exit(1)
	}

	if found := searchStrings(versions, version); !found {
		fmt.Printf("\n[ERROR] unknown go version: %s. List of available versions are: %s\n\n", version, strings.Join(versions, "/"))
		os.Exit(1)
	}

	os.Remove(goSymLink)

	gopath := godir + "/go" + version
	if !pathExists(gopath) {
		fmt.Printf("\n[ERROR] go path not found: %s\n\n", gopath)
		os.Exit(1)
	}

	if err := os.Symlink(gopath, goSymLink); err != nil {
		fmt.Printf("\n[ERROR] failed to symlink go path: %v\n\n", err)
		os.Exit(1)
	}

	out, errExec := exec.Command("go", "version").Output()
	if errExec != nil {
		fmt.Printf("\n[ERROR] failed to execute go version: %v\n", errExec)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func versionList() ([]string, error) {
	fileInfo, err := ioutil.ReadDir(godir)
	if err != nil {
		return nil, fmt.Errorf("unable to list versions: %w", err)
	}

	versions := make([]string, 0)
	for _, file := range fileInfo {
		if file.IsDir() && file.Name() != activeDir {
			version := strings.TrimPrefix(file.Name(), "go")
			versions = append(versions, version)
		}
	}

	return versions, nil
}

func getActiveVersion() (string, error) {
	name, errRead := os.Readlink(goSymLink)
	if errRead != nil {
		return "", fmt.Errorf("failed to read link: %w", errRead)
	}
	name = strings.TrimPrefix(filepath.Base(name), "go")
	return name, nil
}

func download(version, platform, arch string) error {
	if platform == "windows" {
		return fmt.Errorf("unsupported platform")
	}

	if !searchStrings(availablePlatforms, platform) || !searchStrings(availableArch, arch) {
		return fmt.Errorf("unsupported platform or arch, supported platform=%v and arch=%v", availablePlatforms, availableArch)
	}

	versions, errList := versionList()
	if errList != nil {
		return fmt.Errorf("failed to download version: %w", errList)
	}

	if found := searchStrings(versions, version); found {
		return fmt.Errorf("version already exists: %s %v", version, versions)
	}

	url := fmt.Sprintf(gosourceURL, version, platform, arch)
	client := http.Client{Timeout: 60 * time.Second}
	resp, errGet := client.Get(url)
	if errGet != nil {
		return fmt.Errorf("failed to get go source tar ball: %w", errGet)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	buffer := bytes.NewBuffer(data)

	path := "/tmp/" + "go-" + randString(8)
	if err := extract.Gz(context.TODO(), buffer, path, nil); err != nil {
		return fmt.Errorf("failed to extract tar.gz: %w", err)
	}

	os.Rename(path+"/go", godir+"/go"+version)
	os.Remove(path)

	return nil
}

func checkPath() {
	path, _ := os.LookupEnv("PATH")
	if !strings.Contains(path, goSymLink) {
		fmt.Println(instructions)
	}
}

func usage() {
	fmt.Printf(cmdUsageNotice, os.Args[0])
}

func main() {
	flag.Usage = usage
	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	version := downloadCmd.String("version", "", "version to be downloaded")
	platform := downloadCmd.String("platform", "darwin", "os platform")
	arch := downloadCmd.String("arch", "arm64", "os architecture")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)

	setCmd := flag.NewFlagSet("set", flag.ExitOnError)

	versions, _ := versionList()
	cmd := &complete.Command{
		Sub: map[string]*complete.Command{
			"set": {
				Args: predict.Set(versions),
			},
			"delete": {
				Args: predict.Set(versions),
			},
			"list": {},
			"download": {
				Flags: map[string]complete.Predictor{
					"platform": predict.Set(availablePlatforms),
					"arch":     predict.Set(availableArch),
				},
			},
		},
	}

	cmd.Complete("vgo")

	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	switch os.Args[1] {
	case "download":
		downloadCmd.Parse(os.Args[2:])
		if err := download(*version, *platform, *arch); err != nil {
			fmt.Printf("download failed: %v\n", err)
			os.Exit(1)
		}
	case "list":
		listCmd.Parse(os.Args[2:])
		versions, errList := versionList()
		if errList != nil {
			fmt.Printf("\n[ERROR] failed to list versions: %v\n", errList)
			os.Exit(1)
		}
		activeVersion, _ := getActiveVersion()
		for i := 0; i < len(versions); i++ {
			if activeVersion == versions[i] {
				versions[i] = fmt.Sprintf(boldString, versions[i])
			}
		}
		fmt.Printf("\nVersions: %s\n\n", strings.Join(versions, ", "))
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if deleteCmd.NArg() != 1 {
			fmt.Printf("delete command expects 1 argument, received %d\n\n", deleteCmd.NArg())
			flag.Usage()
			os.Exit(1)
		}
		deleteVersion(deleteCmd.Arg(0))
	case "set":
		setCmd.Parse(os.Args[2:])
		if setCmd.NArg() != 1 {
			fmt.Printf("set command expects 1 argument, received %d\n\n", setCmd.NArg())
			flag.Usage()
			os.Exit(1)
		}
		checkPath()
		setVersion(setCmd.Arg(0))
	default:
		fmt.Println("expected 'download' or 'list' or 'delete' or 'set' subcommands")
		flag.Usage()
		os.Exit(1)
	}

	flag.Parse()
}

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

func searchStrings(hay []string, needle string) bool {
	for _, h := range hay {
		if h == needle {
			return true
		}
	}
	return false
}
