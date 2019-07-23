package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type mddConfig struct {
	HashAlg string `json:"hash_alg"`
	Repo    string `json:"repo"`
}

type hashAlgorithm struct {
	Alg    string `json:"alg"`
	Digest string `json:"digest"`
}

type filter struct {
	Description  string        `json:"description"`
	LastModified string        `json:"last_modified"`
	Hash         hashAlgorithm `json:"hash,omitempty"`
	MD5          string        `json:"md5,omitempty"`
	SHA1         string        `json:"sha1,omitempty"`
	SHA256       string        `json:"sha256,omitempty"`
}

func usage(progName string) {
	fmt.Printf("usage: %s <calculate|lookup|fromfile|filters> <filterfile> <file1> [file2 ...]\n", progName)
	os.Exit(1)
}

func calculate() {
	return
}

func lookup() {
	return
}

func getConfig() mddConfig {
	content, readErr := ioutil.ReadFile("config.json")
	if readErr != nil {
		var defaultConfig = mddConfig{
			HashAlg: "sha256",
			Repo:    "https://github.com/roberson-io/mdd_filters/raw/master/repo/",
		}
		data, marshalErr := json.MarshalIndent(defaultConfig, "", "    ")
		if marshalErr != nil {
			log.Fatalf("JSON marshaling failed: %s", marshalErr)
		}
		writeErr := ioutil.WriteFile("config.json", data, 0644)
		if writeErr != nil {
			log.Fatal(writeErr)
		}
		return defaultConfig
	}
	var config mddConfig
	if unmarshalErr := json.Unmarshal(content, &config); unmarshalErr != nil {
		log.Fatalf("JSON unmarshaling failed: %s", unmarshalErr)
	}
	return config
}

func getInstalled(hashAlg string) map[string]filter {
	content, readErr := ioutil.ReadFile("installed.json")
	if readErr != nil {
		var installed map[string]filter
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		dir := filepath.Dir(ex)
		path := filepath.Join(dir, "filters")
		fileInfo, err := ioutil.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range fileInfo {
			var hash = hashAlgorithm{
				Alg:    hashAlg,
				Digest: "",
			}
			installed[file.Name()] = filter{
				Description:  "",
				LastModified: "",
				Hash:         hash,
			}
		}
		data, marshalErr := json.MarshalIndent(installed, "", "    ")
		if marshalErr != nil {
			log.Fatalf("JSON marshaling failed: %s", marshalErr)
		}
		writeErr := ioutil.WriteFile("installed.json", data, 0644)
		if writeErr != nil {
			log.Fatal(writeErr)
		}
		return installed
	}
	var installed map[string]filter
	if unmarshalErr := json.Unmarshal(content, &installed); unmarshalErr != nil {
		log.Fatalf("JSON unmarshaling failed: %s", unmarshalErr)
	}
	return installed
}

func printFilters(filters map[string]filter) {
	fmt.Printf("%-20s%-40s%-20s\n", "Filter", "Description", "Last Modified")
	fmt.Printf("%s\n", strings.Repeat("-", 90))
	var keys []string
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf(
			"%-20s%-40s%-20s\n",
			key,
			filters[key].Description,
			filters[key].LastModified,
		)
	}
}

func listLocal(hashAlg string) {
	installed := getInstalled(hashAlg)
	printFilters(installed)
}

func updateMetadata(baseURL string) map[string]filter {
	var client = &http.Client{Timeout: 10 * time.Second}
	url := baseURL + "METADATA.json"
	response, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var metadata map[string]filter
	if unmarshalErr := json.Unmarshal(content, &metadata); unmarshalErr != nil {
		log.Fatalf("JSON unmarshaling failed: %s", unmarshalErr)
	}
	writeErr := ioutil.WriteFile("METADATA.json", content, 0644)
	if writeErr != nil {
		log.Fatal(writeErr)
	}
	return metadata
}

func listRemote(repoURL string) {
	remote := updateMetadata(repoURL)
	printFilters(remote)
}

func isInstalled() {
	return
}

func downloadFilter() {
	return
}

func updateInstalled() {
	return
}

func fetchFilter() {
	return
}

func updateFilters() {
	return
}

func filters(args []string) {
	command := args[2]
	switch command {
	case "fetch":
		fetchFilter()
	case "list":
		config := getConfig()
		if len(args) > 3 {
			target := args[3]
			if target == "remote" {
				listRemote(config.Repo)
			} else {
				listRemote(target)
			}
		} else {
			listLocal(config.HashAlg)
		}
	case "update":
		updateFilters()
	default:
		fmt.Printf("Invalid command: %s\n", command)
		usage(args[0])
	}
}

func fromfile() {
	return
}

func main() {
	progName := os.Args[0]
	if len(os.Args) < 3 {
		usage(progName)
	}
	command := os.Args[1]
	switch command {
	case "calculate":
		calculate()
	case "lookup":
		lookup()
	case "filters":
		filters(os.Args)
	case "fromfile":
		fromfile()
	default:
		fmt.Printf("Invalid command: %s\n", command)
		usage(progName)
	}
}
