//
// Copyright 2023 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"encoding/json"
	"fmt"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/git"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/devfile/library/v2/pkg/util"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/klog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type MockDevfileCtx struct {

	// devfile ApiVersion
	apiVersion string

	// absolute path of devfile
	absPath string

	// relative path of devfile
	relPath string

	// raw content of the devfile
	rawContent []byte

	// devfile json schema
	jsonSchema string

	// url path of the devfile
	url string

	// token is a personal access token used with a private git repo URL
	token string
	// todo: use token from git interface

	// git is an interface used for git urls
	git git.MockGitUrl

	// filesystem for devfile
	fs filesystem.Filesystem

	// devfile kubernetes components has been converted from uri to inlined in memory
	convertUriToInlined bool
}

// MockNewURLDevfileCtx NewURLDevfileCtx returns a new DevfileCtx type object
func (d *MockDevfileCtx) MockNewURLDevfileCtx(url string) MockDevfileCtx {
	var git = git.MockGitUrl{}
	return MockNewURLDevfileCtxWithGit(url, git)
}

func MockNewURLDevfileCtxWithGit(url string, git git.MockGitUrl) MockDevfileCtx {
	return MockDevfileCtx{
		url: url,
		git: git,
	}
}

func (d *MockDevfileCtx) NewURLDevfileCtxWithGit(url string, git git.MockGitUrl) MockDevfileCtx {
	//gitUrl := GitUrl{
	//	IGitUrl:    nil,
	//	MockGitUrl: git,
	//}
	return MockDevfileCtx{
		url: url,
		git: git,
	}
}

func MockNewPrivateURLDevfileCtxWithGit(url string, token string, git git.MockGitUrl) MockDevfileCtx {
	//gitUrl := GitUrl{
	//	IGitUrl:    nil,
	//	MockGitUrl: git,
	//}
	return MockDevfileCtx{
		url:   url,
		token: token,
		git:   git,
	}
}

// populateDevfile checks the API version is supported and returns the JSON schema for the given devfile API Version
func (d *MockDevfileCtx) populateDevfile() (err error) {

	// Get devfile APIVersion
	if err := d.SetDevfileAPIVersion(); err != nil {
		return err
	}

	// Read and save devfile JSON schema for provided apiVersion
	return d.SetDevfileJSONSchema()
}

// Populate fills the DevfileCtx struct with relevant context info
func (d *MockDevfileCtx) Populate() (err error) {
	if !strings.HasSuffix(d.relPath, ".yaml") {
		if _, err := os.Stat(filepath.Join(d.relPath, "devfile.yaml")); os.IsNotExist(err) {
			if _, err := os.Stat(filepath.Join(d.relPath, ".devfile.yaml")); os.IsNotExist(err) {
				return fmt.Errorf("the provided path is not a valid yaml filepath, and devfile.yaml or .devfile.yaml not found in the provided path : %s", d.relPath)
			} else {
				d.relPath = filepath.Join(d.relPath, ".devfile.yaml")
			}
		} else {
			d.relPath = filepath.Join(d.relPath, "devfile.yaml")
		}
	}
	if err := d.SetAbsPath(); err != nil {
		return err
	}
	klog.V(4).Infof("absolute devfile path: '%s'", d.absPath)
	// Read and save devfile content
	if err := d.SetDevfileContent(); err != nil {
		return err
	}
	return d.populateDevfile()
}

// PopulateFromURL fills the DevfileCtx struct with relevant context info
func (d *MockDevfileCtx) PopulateFromURL() (err error) {
	_, err = url.ParseRequestURI(d.url)
	if err != nil {
		return err
	}
	// Read and save devfile content
	if err := d.SetDevfileContent(); err != nil {
		return err
	}
	return d.populateDevfile()
}

// PopulateFromRaw fills the DevfileCtx struct with relevant context info
func (d *MockDevfileCtx) PopulateFromRaw() (err error) {
	return d.populateDevfile()
}

// Validate func validates devfile JSON schema for the given apiVersion
func (d *MockDevfileCtx) Validate() error {

	// Validate devfile
	return d.ValidateDevfileSchema()
}

// GetGit returns the git object
//func (d *MockDevfileCtx) GetGit() git.MockGitUrl {
//	return d.git
//}

func (d *MockDevfileCtx) GetGit() git.Url {
	return git.Url{
		Protocol: d.git.Protocol,
		Host:     d.git.Host,
		Owner:    d.git.Owner,
		Repo:     d.git.Repo,
		Branch:   d.git.Branch,
		Path:     d.git.Path,
		IsFile:   d.git.IsFile,
	}
}

// GetAbsPath func returns current devfile absolute path
func (d *MockDevfileCtx) GetAbsPath() string {
	return d.absPath
}

// GetURL func returns current devfile absolute URL address
func (d *MockDevfileCtx) GetURL() string {
	return d.url
}

// GetToken func returns current devfile token
func (d *MockDevfileCtx) GetToken() string {
	return d.token
}

// SetAbsPath sets absolute file path for devfile
func (d *MockDevfileCtx) SetAbsPath() (err error) {
	// Set devfile absolute path
	if d.absPath, err = util.GetAbsPath(d.relPath); err != nil {
		return err
	}
	klog.V(2).Infof("absolute devfile path: '%s'", d.absPath)

	return nil

}

// GetConvertUriToInlined func returns if the devfile kubernetes comp has been converted from uri to inlined
func (d *MockDevfileCtx) GetConvertUriToInlined() bool {
	return d.convertUriToInlined
}

// SetConvertUriToInlined sets if the devfile kubernetes comp has been converted from uri to inlined
func (d *MockDevfileCtx) SetConvertUriToInlined(value bool) {
	d.convertUriToInlined = value
}

// SetDevfileAPIVersion returns the devfile APIVersion
func (d *MockDevfileCtx) SetDevfileAPIVersion() error {

	// Unmarshal JSON into map
	var r map[string]interface{}
	err := json.Unmarshal(d.rawContent, &r)
	if err != nil {
		return errors.Wrapf(err, "failed to decode devfile json")
	}

	// Get "schemaVersion" value from map for devfile V2
	schemaVersion, okSchema := r["schemaVersion"]
	var devfilePath string
	if d.GetAbsPath() != "" {
		devfilePath = d.GetAbsPath()
	} else if d.GetURL() != "" {
		devfilePath = d.GetURL()
	}

	if okSchema {
		// SchemaVersion cannot be empty
		if schemaVersion.(string) == "" {
			return fmt.Errorf("schemaVersion in devfile: %s cannot be empty", devfilePath)
		}
	} else {
		return fmt.Errorf("schemaVersion not present in devfile: %s", devfilePath)
	}

	// Successful
	// split by `-` and get the first substring as schema version, schemaVersion without `-` won't get affected
	// e.g. 2.2.0-latest => 2.2.0, 2.2.0 => 2.2.0
	d.apiVersion = strings.Split(schemaVersion.(string), "-")[0]
	klog.V(4).Infof("devfile schemaVersion: '%s'", d.apiVersion)
	return nil
}

// GetApiVersion returns apiVersion stored in devfile context
func (d *MockDevfileCtx) GetApiVersion() string {
	return d.apiVersion
}

// IsApiVersionSupported return true if the apiVersion in DevfileCtx is supported
func (d *MockDevfileCtx) IsApiVersionSupported() bool {
	return data.IsApiVersionSupported(d.apiVersion)
}

// SetDevfileJSONSchema returns the JSON schema for the given devfile apiVersion
func (d *MockDevfileCtx) SetDevfileJSONSchema() error {

	// Check if json schema is present for the given apiVersion
	jsonSchema, err := data.GetDevfileJSONSchema(d.apiVersion)
	if err != nil {
		return err
	}
	d.jsonSchema = jsonSchema
	return nil
}

// ValidateDevfileSchema validate JSON schema of the provided devfile
func (d *MockDevfileCtx) ValidateDevfileSchema() error {
	var (
		schemaLoader   = gojsonschema.NewStringLoader(d.jsonSchema)
		documentLoader = gojsonschema.NewStringLoader(string(d.rawContent))
	)

	// Validate devfile with JSON schema
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return errors.Wrapf(err, "failed to validate devfile schema")
	}

	if !result.Valid() {
		errMsg := "invalid devfile schema. errors :\n"
		for _, desc := range result.Errors() {
			errMsg = errMsg + fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf(errMsg)
	}

	// Sucessful
	klog.V(4).Info("validated devfile schema")
	return nil
}

// SetDevfileContent reads devfile and if devfile is in YAML format converts it to JSON
func (d *MockDevfileCtx) SetDevfileContent() error {

	var err error
	var data []byte
	if d.url != "" {
		// set the client identifier for telemetry
		params := util.HTTPRequestParams{URL: d.url, TelemetryClientName: util.TelemetryClientName}
		if d.token != "" {
			params.Token = d.token
		}
		data, err = util.DownloadInMemory(params)
		if err != nil {
			return errors.Wrap(err, "error getting devfile info from url")
		}
	} else if d.absPath != "" {
		// Read devfile
		fs := d.GetFs()
		data, err = fs.ReadFile(d.absPath)
		if err != nil {
			return errors.Wrapf(err, "failed to read devfile from path '%s'", d.absPath)
		}
	}

	// set devfile content
	return d.SetDevfileContentFromBytes(data)
}

// SetDevfileContentFromBytes sets devfile content from byte input
func (d *MockDevfileCtx) SetDevfileContentFromBytes(data []byte) error {
	// If YAML file convert it to JSON
	var err error
	d.rawContent, err = YAMLToJSON(data)
	if err != nil {
		return err
	}

	// Successful
	return nil
}

// GetDevfileContent returns the devfile content
func (d *MockDevfileCtx) GetDevfileContent() []byte {
	return d.rawContent
}

// GetFs returns the filesystem object
func (d *MockDevfileCtx) GetFs() filesystem.Filesystem {
	return d.fs
}
