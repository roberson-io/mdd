package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
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
	LastModified time.Time     `json:"last_modified"`
	Hash         hashAlgorithm `json:"hash,omitempty"`
	MD5          string        `json:"md5,omitempty"`
	SHA1         string        `json:"sha1,omitempty"`
	SHA256       string        `json:"sha256,omitempty"`
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

// Possibly not needed?
func getMetadata() map[string]filter {
	path := filepath.Join(currentDir(), "METADATA.json")
	content, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		config := getConfig()
		metadata := updateMetadata(config.Repo)
		return metadata
	}
	var metadata map[string]filter
	if unmarshalErr := json.Unmarshal(content, &metadata); unmarshalErr != nil {
		log.Fatalf("JSON unmarshaling failed: %s", unmarshalErr)
	}
	return metadata
}

func getHasher(hashAlgorithm string) hash.Hash {
	var hasher hash.Hash
	switch hashAlgorithm {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	default:
		log.Fatalf("Invalid hash algorithm: %s", hashAlgorithm)
	}
	return hasher
}

func currentDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(ex)
	return dir
}

func filterPath() string {
	path := filepath.Join(currentDir(), "filters")
	return path
}
func getInstalled(hashAlg string) map[string]filter {
	content, readErr := ioutil.ReadFile("installed.json")
	if readErr != nil {
		installed := make(map[string]filter)
		fileInfo, err := ioutil.ReadDir(filterPath())
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range fileInfo {
			path := filepath.Join(filterPath(), file.Name())
			f, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			hasher := getHasher(hashAlg)
			if _, err := io.Copy(hasher, f); err != nil {
				log.Fatal(err)
			}
			var hash = hashAlgorithm{
				Alg:    hashAlg,
				Digest: hex.EncodeToString(hasher.Sum(nil)),
			}
			installed[file.Name()] = filter{
				Description:  "",
				LastModified: file.ModTime(),
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

func listRemote(repoURL string) {
	remote := updateMetadata(repoURL)
	printFilters(remote)
}

func isInstalled(target string) bool {
	config := getConfig()
	installed := getInstalled(config.HashAlg)
	_, inInstalledFile := installed[target]
	path := filepath.Join(filterPath(), target)
	fileExists := true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fileExists = false
	}
	return inInstalledFile && fileExists
}

func downloadFilter(target string) {
	config := getConfig()
	url := config.Repo + target
	fmt.Printf("Fetching %s...\n", target)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s filter not found!\n", target)
		os.Exit(1)
	}
	defer response.Body.Close()
	path := filepath.Join(filterPath(), target)
	filter, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer filter.Close()
	_, err = io.Copy(filter, response.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func updateInstalled(target string) {
	config := getConfig()
	metadata := updateMetadata(config.Repo)
	installed := getInstalled(config.HashAlg)
	targetData := metadata[target]
	var digest string
	switch config.HashAlg {
	case "md5":
		digest = targetData.MD5
	case "sha1":
		digest = targetData.SHA1
	case "sha256":
		digest = targetData.SHA256
	default:
		log.Fatalf("Invalid hash algorithm in config.json: %s", config.HashAlg)
	}
	var hash = hashAlgorithm{
		Alg:    config.HashAlg,
		Digest: digest,
	}
	installed[target] = filter{
		Description:  targetData.Description,
		LastModified: targetData.LastModified,
		Hash:         hash,
	}
	path := filepath.Join(currentDir(), "installed.json")
	data, marshalErr := json.MarshalIndent(installed, "", "    ")
	if marshalErr != nil {
		log.Fatalf("JSON marshaling failed: %s", marshalErr)
	}
	writeErr := ioutil.WriteFile(path, data, 0644)
	if writeErr != nil {
		log.Fatal(writeErr)
	}
}

func fetchFilter(target string) {
	if isInstalled(target) {
		fmt.Printf("%s is already installed\n", target)
	} else {
		downloadFilter(target)
		updateInstalled(target)
	}
}

func updateFilters() {
	config := getConfig()
	installed := getInstalled(config.HashAlg)
	metadata := updateMetadata(config.Repo)
	var keys []string
	for k := range installed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, target := range keys {
		_, inMetadata := metadata[target]
		if inMetadata {
			here := installed[target]
			there := metadata[target]
			diff := there.LastModified.Sub(here.LastModified)
			if (diff > 0) && (there.Hash.Digest != here.Hash.Digest) {
				fmt.Printf("Updating %s...\n", target)
				downloadFilter(target)
				updateInstalled(target)
				fmt.Print("Done.\n")
			}
		}
	}
}
