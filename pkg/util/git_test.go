//
// Copyright 2021-2022 Red Hat, Inc.
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

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestParseGitUrl(t *testing.T) {
	defer func() {
		err := os.Unsetenv("GITHUB_TOKEN")
		if err != nil {
			fmt.Println("failed to unset github token")
		}
	}()

	err := os.Setenv("GITHUB_TOKEN", "")
	if err != nil {
		fmt.Println("failed to set github token env")
	}

	missingEnv := os.Getenv("MISSING_GH_TOKEN")
	fmt.Printf("missing github token: [%s]\n", missingEnv)

	// https://github.com/mike-hoang/private-repo-example
	//privateRepo := map[string]string{
	//	"username": "mike-hoang",
	//	"project":  "private-repo-example",
	//	"token":    os.Getenv("GITHUB_TOKEN"),
	//	"branch":   "main",
	//}

	//publicRepo := map[string]string{
	//	"username": "mike-hoang",
	//	"project":  "devfile-stack-go",
	//	"branch":   "main",
	//	"token":    "",
	//}

	//err = CloneGitRepo(publicRepo, "./temp")
	//if err != nil {
	//	fmt.Println("failed to clone; ", err)
	//}

	err = ParseGitUrl("https://github.com/mike-hoang/private-repo-example")
	if err != nil {
		return
	}

	err = ParseGitUrl("https://raw.githubusercontent.com/devfile/registry/main/stacks/nodejs/devfile.yaml")
	if err != nil {
		return
	}

}

func TestGetGitUrlComponentsFromRaw(t *testing.T) {
	validRawGitUrl := "https://raw.githubusercontent.com/username/project/branch/file/path"
	invalidUrl := "github.com/not/valid/url"

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "should be able to get git url components",
			url:     validRawGitUrl,
			wantErr: false,
		},
		{
			name:    "should fail with invalid raw git url",
			url:     invalidUrl,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetGitUrlComponentsFromRaw(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error: %t, got error: %t", tt.wantErr, err)
			}
		})
	}
}

func TestCloneGitRepo(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create temp dir: %s, error: %v", tempDir, err)
	}
	defer os.RemoveAll(tempDir)

	invalidGitUrl := map[string]string{
		"username": "devfile",
		"project":  "nonexistent",
		"branch":   "nonexistent",
	}
	validGitUrl := map[string]string{
		"username": "devfile",
		"project":  "library",
		"branch":   "main",
	}

	tests := []struct {
		name             string
		gitUrlComponents map[string]string
		destDir          string
		wantErr          bool
	}{
		{
			name:             "should fail with invalid git url",
			gitUrlComponents: invalidGitUrl,
			destDir:          filepath.Join(os.TempDir(), "nonexistent"),
			wantErr:          true,
		},
		{
			name:             "should be able to clone valid git url",
			gitUrlComponents: validGitUrl,
			destDir:          tempDir,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CloneGitRepo(tt.gitUrlComponents, tt.destDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error: %t, got error: %t", tt.wantErr, err)
			}
		})
	}
}
