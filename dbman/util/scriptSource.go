//   Onix Config Manager - Dbman
//   Copyright (c) 2018-2020 by www.gatblau.org
//   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
//   Contributors to this project, hereby assign copyright in this code to the project,
//   to be licensed under the same terms as the rest of the code.
package util

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gatblau/oxc"
	"net/http"
)

// the source of database scripts
type ScriptSource struct {
	client    *oxc.Client
	index     *Plan
	manifests []Release
	cfg       *AppCfg
}

// factory function
func NewScriptSource(cfg *AppCfg, client *oxc.Client) (*ScriptSource, error) {
	// creates a new struct
	source := new(ScriptSource)
	// setup attributes
	source.cfg = cfg
	source.client = client
	return source, nil
}

// new oxc configuration
func NewClientConf(cfg *AppCfg) *oxc.ClientConf {
	return &oxc.ClientConf{
		BaseURI:            cfg.get(SchemaURI),
		InsecureSkipVerify: false,
		AuthMode:           oxc.None,
	}
}

// access a cached index reference
func (info *ScriptSource) plan() *Plan {
	// if the index is not fetched
	if info.index == nil {
		// fetches it
		err := info.loadPlan()
		if err != nil {
			fmt.Sprintf("cannot retrieve plan, %v", err)
		}
	}
	return info.index
}

// (re)loads the internal index reference
func (info *ScriptSource) loadPlan() error {
	ix, err := info.fetchPlan()
	info.index = ix
	return err
}

// fetches the release index
func (info *ScriptSource) fetchPlan() (*Plan, error) {
	if info.cfg == nil {
		return nil, errors.New("configuration object not initialised when fetching release plan")
	}
	response, err := info.client.Get(fmt.Sprintf("%s/plan.json", info.get(SchemaURI)), info.addHttpHeaders)
	if err != nil {
		return nil, err
	}
	i := &Plan{}
	i, err = i.decode(response)
	defer func() {
		if ferr := response.Body.Close(); ferr != nil {
			err = ferr
		}
	}()
	return i, err
}

// fetches the scripts for a database release
func (info *ScriptSource) fetchRelease(appVersion string) (*Release, error) {
	// if cfg not initialised, no point in continuing
	if info.cfg == nil {
		return nil, errors.New("configuration object not initialised when calling fetching release")
	}
	// get the release information based on the
	ri, err := info.release(appVersion)
	if err != nil {
		// could not find release information in the release index
		return nil, err
	}
	// builds a uri to fetch the specific release manifest
	uri := fmt.Sprintf("%s/%s/release.json", info.get(SchemaURI), ri.Path)
	// fetch the release.json manifest
	response, err := info.client.Get(uri, info.addHttpHeaders)
	// if the request was unsuccessful then return the error
	if err != nil {
		return nil, err
	}
	// request was good so construct a release manifest reference
	r := &Release{}
	r, err = r.decode(response)
	defer func() {
		if ferr := response.Body.Close(); ferr != nil {
			err = ferr
		}
	}()
	return r, err
}

// get the release information for a given application version
func (info *ScriptSource) release(appVersion string) (*Info, error) {
	for _, release := range info.plan().Releases {
		if release.AppVersion == appVersion {
			return &release, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("information for application version '%s' does not exist in the release index", appVersion))
}

// add http headers to the request object
func (info *ScriptSource) addHttpHeaders(req *http.Request, payload oxc.Serializable) error {
	// add headers to disable caching
	req.Header.Add("Cache-Control", `no-cache"`)
	req.Header.Add("Pragma", "no-cache")
	// if there is an access token defined
	if len(info.get(SchemaUsername)) > 0 && len(info.get(SchemaToken)) > 0 {
		credentials := base64.StdEncoding.EncodeToString([]byte(
			fmt.Sprintf("%s:%s", info.get(SchemaUsername), info.get(SchemaToken))))
		req.Header.Add("Authorization", credentials)
	}
	return nil
}

func (info *ScriptSource) get(key string) string {
	return info.cfg.get(key)
}
