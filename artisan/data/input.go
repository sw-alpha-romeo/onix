/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-2021 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
package data

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/gatblau/onix/artisan/core"
	"github.com/gatblau/onix/artisan/crypto"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

// describes exported input information required by functions or runtimes
type Input struct {
	// required PGP keys
	Key []*Key `yaml:"key,omitempty" json:"key,omitempty"`
	// required string value secrets
	Secret []*Secret `yaml:"secret,omitempty" json:"secret,omitempty"`
	// required variables
	Var []*Var `yaml:"var,omitempty" json:"var,omitempty"`
}

func (i *Input) ContainsVar(binding string) bool {
	for _, variable := range i.Var {
		if variable.Name == binding {
			return true
		}
	}
	return false
}

func (i *Input) ContainsSecret(binding string) bool {
	for _, secret := range i.Secret {
		if secret.Name == binding {
			return true
		}
	}
	return false
}

func (i *Input) ContainsKey(binding string) bool {
	for _, key := range i.Key {
		if key.Name == binding {
			return true
		}
	}
	return false
}

func (i *Input) Encrypt(pub *crypto.PGP) {
	encryptInput(i, pub)
}

// extracts the build file Input that is relevant to a function (using its bindings)
func InputFromBuildFile(fxName string, buildFile *BuildFile, prompt bool) *Input {
	if buildFile == nil {
		core.RaiseErr("build file is required")
	}
	// get the build file function to inspect
	fx := buildFile.Fx(fxName)
	if fx == nil {
		core.RaiseErr("function '%s' cannot be found in build file", fxName)
	}
	return getInput(fx.Input, buildFile.Input, prompt)
}

// extracts the package manifest Input in an exported function
func InputFromManifest(name *core.PackageName, fxName string, manifest *Manifest, prompt bool) *Input {
	// get the function in the manifest
	fx := manifest.Fx(fxName)
	if fx == nil {
		core.RaiseErr("function '%s' does not exist in or has not been exported", fxName)
	}
	input := *fx.Input
	// as we need to open this package a verification key is needed
	// then, add the key to the inputs automatically
	input.Key = append(input.Key, &Key{
		Name:        fmt.Sprintf("%s_%s_VERIFICATION_KEY", strings.ToUpper(name.Group), strings.ToUpper(name.Name)),
		Description: fmt.Sprintf("the public PGP key required to open the package %s", name),
		Private:     false,
	})
	if prompt {
		surveyInput(&input)
	}
	return &input
}

func InputFromURI(uri string, prompt bool) *Input {
	response, err := core.Get(uri, "", "")
	core.CheckErr(err, "cannot fetch runtime manifest")
	body, err := ioutil.ReadAll(response.Body)
	core.CheckErr(err, "cannot read runtime manifest http response")
	// need a wrapper object for the input for the unmarshaller to work so using buildfile
	var buildFile = new(BuildFile)
	err = yaml.Unmarshal(body, buildFile)
	if prompt {
		return surveyInput(buildFile.Input)
	}
	return buildFile.Input
}

func surveyInput(input *Input) *Input {
	// makes a shallow copy of the input
	result := *input
	// collect values from command line interface
	for _, v := range result.Var {
		surveyVar(v)
	}
	for _, secret := range result.Secret {
		surveySecret(secret)
	}
	for _, key := range result.Key {
		surveyKey(key)
	}
	// return pointer to new object
	return &result
}

// extract any Input data from the source that have a binding
func getInput(fxInput *InputBinding, sourceInput *Input, prompt bool) *Input {
	// if no bindings then return no Input
	if fxInput == nil {
		return nil
	}
	result := &Input{
		Key:    make([]*Key, 0),
		Secret: make([]*Secret, 0),
		Var:    make([]*Var, 0),
	}
	// collects exported vars
	for _, varBinding := range fxInput.Var {
		for _, variable := range sourceInput.Var {
			if variable.Name == varBinding {
				result.Var = append(result.Var, variable)
				// if interactive mode is enabled then prompt the user to enter the variable value
				if prompt {
					surveyVar(variable)
				}
			}
		}
	}
	// collect exported secrets
	for _, secretBinding := range fxInput.Secret {
		for _, secret := range sourceInput.Secret {
			if secret.Name == secretBinding {
				result.Secret = append(result.Secret, secret)
				// if interactive mode is enabled then prompt the user to enter the variable value
				if prompt {
					surveySecret(secret)
				}
			}
		}
	}
	// collect exported keys
	for _, keyBinding := range fxInput.Key {
		for _, key := range sourceInput.Key {
			if key.Name == keyBinding {
				result.Key = append(result.Key, key)
				// if interactive mode is enabled then prompt the user to enter the variable value
				if prompt {
					surveyKey(key)
				}
			}
		}
	}
	return result
}

// encrypts secret and key values
func encryptInput(input *Input, encPubKey *crypto.PGP) {
	if input == nil {
		return
	}
	for _, secret := range input.Secret {
		// and encrypts the secret value
		err := secret.Encrypt(encPubKey)
		core.CheckErr(err, "cannot encrypt secret")
	}
	for _, key := range input.Key {
		// and encrypts the key value
		err := key.Encrypt(encPubKey)
		core.CheckErr(err, "cannot encrypt PGP key %s: %s", key.Name, err)
	}
}

func surveyVar(variable *Var) {
	// check if the var is defined in the environment
	value := os.Getenv(variable.Name)
	// if it is
	if len(value) > 0 {
		// sets it with  its value and return
		variable.Value = value
		return
	}
	// otherwise prompts the user to enter it
	var validator survey.Validator
	desc := ""
	// if a description is available use it
	if len(variable.Description) > 0 {
		desc = variable.Description
	}
	// prompt for the value
	prompt := &survey.Input{
		Message: fmt.Sprintf("var => %s (%s):", variable.Name, desc),
	}
	// if required then add required validator
	if variable.Required {
		validator = survey.ComposeValidators(survey.Required)
	}
	// add type validators
	switch strings.ToLower(variable.Type) {
	case "path":
		validator = survey.ComposeValidators(validator, isPath)
	case "uri":
		validator = survey.ComposeValidators(validator, isURI)
	case "name":
		validator = survey.ComposeValidators(validator, isPackageName)
	}
	core.HandleCtrlC(survey.AskOne(prompt, &variable.Value, survey.WithValidator(validator)))
}

func surveySecret(secret *Secret) {
	// check if the secret is defined in the environment
	value := os.Getenv(secret.Name)
	// if it is
	if len(value) > 0 {
		// sets it with  its value and return
		secret.Value = value
		return
	}
	desc := ""
	// if a description is available use it
	if len(secret.Description) > 0 {
		desc = secret.Description
	}
	// prompt for the value
	prompt := &survey.Password{
		Message: fmt.Sprintf("secret => %s (%s):", secret.Name, desc),
	}
	core.HandleCtrlC(survey.AskOne(prompt, &secret.Value, survey.WithValidator(survey.Required)))
}

func surveyKey(key *Key) {
	desc := ""
	// if a description is available use it
	if len(key.Description) > 0 {
		desc = key.Description
	}
	// prompt for the value
	prompt := &survey.Input{
		Message: fmt.Sprintf("PGP key => %s PATH (%s):", key.Name, desc),
		Default: "/",
		Help:    "/ indicates root keys; /group-name indicates group level keys; /group-name/package-name indicates package level keys",
	}
	var (
		keyPath, pk, pub string
		keyBytes         []byte
		err              error
	)
	core.HandleCtrlC(survey.AskOne(prompt, &keyPath, survey.WithValidator(keyPathExist)))
	// load the keys
	parts := strings.Split(keyPath, "/")
	switch len(parts) {
	case 2:
		// root level keys
		if len(parts[1]) == 0 {
			pk, pub = crypto.KeyNames(core.KeysPath(), "root", "pgp")
			key.PackageGroup = ""
			key.PackageName = ""
		} else {
			// group level keys
			pk, pub = crypto.KeyNames(path.Join(core.KeysPath(), parts[1]), parts[1], "pgp")
			key.PackageGroup = parts[1]
			key.PackageName = ""
		}
	// package level keys
	case 3:
		pk, pub = crypto.KeyNames(path.Join(core.KeysPath(), parts[1], parts[2]), fmt.Sprintf("%s_%s", parts[1], parts[2]), "pgp")
		key.PackageGroup = parts[1]
		key.PackageName = parts[2]
	// error
	default:
		core.RaiseErr("the provided path %s is invalid", keyPath)
	}
	if key.Private {
		keyBytes, err = ioutil.ReadFile(pk)
		core.CheckErr(err, "cannot read private key from registry")
	} else {
		keyBytes, err = ioutil.ReadFile(pub)
		core.CheckErr(err, "cannot read public key from registry")
	}
	key.Value = string(keyBytes)
}

func keyPathExist(val interface{}) error {
	// the reflect value of the result
	value := reflect.ValueOf(val)

	// if the value passed in is a string
	if value.Kind() == reflect.String {
		if len(value.String()) > 0 {
			if !strings.HasPrefix(value.String(), "/") {
				// it is not a valid package name
				return fmt.Errorf("key path '%s' must start with a forward slash", value.String())
			}
			_, err := os.Stat(filepath.Join(core.KeysPath(), value.String()))
			// if the path to the group does not exist
			if os.IsNotExist(err) {
				// it is not a valid package name
				return fmt.Errorf("key path '%s' does not exist", value.String())
			}
		}
	} else {
		// if the value is not of a string type it cannot be a path
		return fmt.Errorf("key group must be a string")
	}
	return nil
}
