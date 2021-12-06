/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-2021 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package app

import (
	"fmt"
	"github.com/gatblau/onix/artisan/build"
	"github.com/gatblau/onix/artisan/core"
	"github.com/gatblau/onix/artisan/merge"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func loadFromFile(path string) (*Manifest, error) {
	if len(path) == 0 {
		path = "."
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("cannot convert app manifest to absolute path: %s\n", err)
	}
	filename := filepath.Join(path, "app.yaml")
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot read app manifest: %s\n", err)
	}
	appMan := new(Manifest)
	err = yaml.Unmarshal(file, appMan)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal app manifest: %s\n", err)
	}
	return appMan, nil
}

func loadFromURL(url string) (*Manifest, error) {
	content, err := fetchFile(url)
	if err != nil {
		return nil, err
	}
	m := new(Manifest)
	err = yaml.Unmarshal(content, m)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal remotely fetched app manifest yaml file: %s\n", err)
	}
	return m, nil
}

// loadSvcManFromImage extracts the service manifest from a docker image
func loadSvcManFromImage(svcRef SvcRef) (*SvcManifest, error) {
	var cmd string
	pathLabel := "artisan.svc.manifest"
	containerName := fmt.Sprintf("%s-info", svcRef.Name)
	// create a container instance in stopped state
	cmd = fmt.Sprintf("docker create --name %s %s", containerName, svcRef.Image)
	_, err := build.Exe(cmd, ".", merge.NewEnVarFromSlice([]string{}), false)
	if err != nil {
		return nil, fmt.Errorf("cannot create docker container from image '%s': %s\n", svcRef.Image, err)
	}
	// create a random tmp folder
	tmp, err := tmpPath()
	if err != nil {
		return nil, fmt.Errorf("cannot create tmp working folder: %s\n", err)
	}
	// inspect docker manifest looking for artisan.svc.manifest label value
	cmd = fmt.Sprintf("docker inspect --format '{{ index .Config.Labels \"%s\"}}' %s", pathLabel, containerName)
	svcManPath, err := build.Exe(cmd, ".", merge.NewEnVarFromSlice([]string{}), false)
	if err != nil {
		return nil, fmt.Errorf("cannot extract service manifest path from container label: %s\n", err)
	}
	if len(svcManPath) == 0 {
		return nil, fmt.Errorf("missing service manifest path label '%s' in image '%s': %s\n", pathLabel, svcRef.Image, err)
	}
	// extract service manifest
	svcPath := filepath.Join(tmp, "svc.yaml")
	cmd = fmt.Sprintf("docker cp %s:%s %s", containerName, svcManPath, svcPath)
	_, err = build.Exe(cmd, ".", merge.NewEnVarFromSlice([]string{}), false)
	if err != nil {
		return nil, fmt.Errorf("cannot copy service manifest from container image: %s\n", err)
	}
	// remove the svc-info container
	_, err = build.Exe(fmt.Sprintf("docker rm -f %s", containerName), ".", merge.NewEnVarFromSlice([]string{}), false)
	if err != nil {
		return nil, fmt.Errorf("cannot remove info container image: %s\n", err)
	}
	// load the service manifest
	content, err := os.ReadFile(svcPath)
	if err != nil {
		return nil, fmt.Errorf("cannot load service manifest: %s\n", err)
	}
	// unmarshal manifest
	svcMan := new(SvcManifest)
	err = yaml.Unmarshal(content, svcMan)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal service manifest: %s\n", err)
	}
	// remove tmp folder
	cmd = fmt.Sprintf("rm -rf %s", tmp)
	_, err = build.Exe(cmd, ".", merge.NewEnVarFromSlice([]string{}), false)
	if err != nil {
		return nil, fmt.Errorf("cannot remove tmp working folder: %s\n", err)
	}
	return svcMan, nil
}

// loadSvcManFromURI extracts the service manifest from a remote URI
func loadSvcManFromURI(svc SvcRef, credentials string) (*SvcManifest, error) {
	uri, err := addCredentialsToURI(svc.URI, credentials)
	if err != nil {
		return nil, err
	}
	content, err := fetchFile(uri)
	if err != nil {
		return nil, err
	}
	m := new(SvcManifest)
	err = yaml.Unmarshal(content, m)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal remotely fetched service manifest yaml file: %s\n", err)
	}
	return m, nil
}

func isFile(uri string) (isFile bool, path string) {
	isFile = strings.HasPrefix(uri, "file://")
	path = uri[len("file://"):]
	return
}

func isURL(uri string) bool {
	return strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")
}

// return a temporary path and create the tmp folder
func tmpPath() (string, error) {
	uuid.EnableRandPool()
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	path := filepath.Join(core.TmpPath(), id.String())
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

// fetchFile fetches a file content from an url
// TODO: implement authentication
func fetchFile(url string) ([]byte, error) {
	client := http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("cannot fetch '%s': %s\n", url, resp.Status)
	}
	return ioutil.ReadAll(resp.Body)
}
