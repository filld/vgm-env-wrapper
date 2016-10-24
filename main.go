package main

import (
	"github.com/channelmeter/vault-gatekeeper-mesos/gatekeeper"
	"os"
	"os/exec"
	"strings"
	"net/http"
	"net/url"
	"github.com/Jeffail/gabs"
	"io/ioutil"
	"errors"
	"log"
	"syscall"
	"path/filepath"
	"runtime"
)

type EnvMap struct {
	VaultPath string
	VaultKey  string
	EnvVar    string
	File      *os.File
}

func getSecret(c *gatekeeper.Client, token string, vaultPath string, vaultKey string) (string, error) {
	vaultAddr, err := url.Parse(c.VaultAddress)
	if err != nil {
		return "", err
	}
	vaultAddr.Path = "/v1/" + vaultPath

	req, err := http.NewRequest("GET", vaultAddr.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Vault-Token", token)

	vaultResp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer vaultResp.Body.Close()

	body, err := ioutil.ReadAll(vaultResp.Body)

	parsed, err := gabs.ParseJSON(body)
	if err != nil {
		return "", err
	}
	value, ok := parsed.Search("data", vaultKey).Data().(string)
	if !ok {
		return "", errors.New("Faile to parse JSON response from vault: " + string(body))
	}
	if len(value) == 0 {
		return "", errors.New("Empty value")
	}
	return value, nil
}

const Version = "0.1.0"

func main() {
	log.SetFlags(0) // no timestamps on our logs

	if len(os.Args) <= 1 {
		self := filepath.Base(os.Args[0])
		log.Printf("Usage: %s command [args]", self)
		log.Println()
		log.Printf("%s version: %s (%s on %s/%s; %s)", self, Version, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
		log.Println()
		os.Exit(1)
	}

	var env []string
	env = os.Environ()

	var mapping []EnvMap;

	for _, value := range env {
		name := strings.Split(value, "=")
		v := name[0]
		parts := strings.Split(name[1], ":")
		if len(parts) == 3 && len(parts[1]) > 0 && len(parts[2]) > 0 {
			if parts[0] == "vgm" {
				mapping = append(mapping,
					EnvMap{
						VaultPath:parts[1],
						VaultKey: parts[2],
						EnvVar: v,
					});
			}
			if parts[0] == "vgm_file" {
				tmpfile, err := ioutil.TempFile("", "vgm")
				if err != nil {
					log.Fatal("Failed to save temp file: " + err.Error())
				}
				mapping = append(mapping,
					EnvMap{
						VaultPath:parts[1],
						VaultKey: parts[2],
						EnvVar: v,
						File: tmpfile,
					});
			}
		}
	}
	token, err := gatekeeper.EnvRequestVaultToken()
	if err != nil {
		log.Fatal("Could not get vault token: " + err.Error());
	}
	for _, mapping := range mapping {
		secret, err := getSecret(gatekeeper.DefaultClient, token, mapping.VaultPath, mapping.VaultKey)
		if err != nil {
			log.Fatal("Could not get secret for " + mapping.EnvVar + " : " + err.Error());
		}
		if mapping.File != nil {
			if _, err := mapping.File.Write([]byte(secret)); err != nil {
				log.Fatal(err)
			}
			os.Setenv(mapping.EnvVar, mapping.File.Name());
		} else {
			os.Setenv(mapping.EnvVar, secret);
		}
	}

	name, err := exec.LookPath(os.Args[1])
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err = syscall.Exec(name, os.Args[1:], os.Environ()); err != nil {
		log.Fatalf("error: exec failed: %v", err)
	}
}
