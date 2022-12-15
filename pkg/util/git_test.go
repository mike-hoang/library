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
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	githubToken    = "fake-github-token"
	gitlabToken    = "fake-gitlab-token"
	bitbucketToken = "fake-bitbucket-token"
	url1           = "https://github.com/mike-hoang/private-repo-example"
	url2           = "https://github.com/mike-hoang/private-repo-example/blob/main/README.md"
	url3           = "https://github.com/mike-hoang/registry/tree/main/tests"
	url4           = "https://raw.githubusercontent.com/devfile/registry/main/stacks/nodejs/devfile.yaml"
)

func TestGitHubParser(t *testing.T) {
	{
		g, err := ParseGitUrl(url1)
		assert.NoError(t, err)
		assert.Equal(t, "github.com", g.Host)
		assert.Equal(t, "kubescape", g.Owner)
		assert.Equal(t, "go-git-url", g.Repo)
		assert.Equal(t, "", g.Branch)
		assert.Equal(t, "", g.Path)
	}
}

func TestParseGitUrl(t *testing.T) {
	defer func() {
		err := os.Unsetenv(githubToken)
		if err != nil {
			fmt.Println("failed to unset github token")
		}
		err = os.Unsetenv(bitbucketToken)
		if err != nil {
			fmt.Println("failed to unset bitbucket token")
		}
		err = os.Unsetenv(gitlabToken)
		if err != nil {
			fmt.Println("failed to unset gitlab token")
		}
	}()

	err := os.Setenv("GITHUB_TOKEN", githubToken)
	if err != nil {
		fmt.Println("failed to set github token env")
	}

	err = os.Setenv("BITBUCKET_TOKEN", bitbucketToken)
	if err != nil {
		fmt.Println("failed to set bitbucket token env")
	}

	err = os.Setenv("GITLAB_TOKEN", gitlabToken)
	if err != nil {
		fmt.Println("failed to set gitlab token env")
	}

	giturls := []string{
		//"https://github.com/mike-hoang/private-repo-example",
		//"https://github.com/mike-hoang/private-repo-example/blob/main/README.md",
		//"https://github.com/mike-hoang/private-repo-example/blob/main",
		//"https://github.com/mike-hoang/registry/tree/main/tests",
		//"https://raw.githubusercontent.com/devfile/registry/main/stacks/nodejs/devfile.yaml",
		//"https://github.com/devfile/registry/blob/main/stacks/nodejs/devfile.yaml",
		//"https://github.com/whilp/git-urls/blob/master/urls_test.go",
		//"https://github.com/go-git/go-git/blob/f9b2cce5c9e6510fefb21ce54c07b59e83b6da8f/plumbing/transport/common.go#L101",
		//"https://github.com/go-git/go-git/tree/master/plumbing/transport",
		//"https://raw.githubusercontent.com/devfile",
		//"https://google.ca",
		//"",
		//" ",
		//"https://raw.githubusercontent.com/",
		//"https://raw.githubusercontent.com/devfile",
		//"https://github.com/mike-hoang",

		//"https://bitbucket.org/mike-hoang/private-repo-example/src/main/README.md",
		//"https://bitbucket.org/mike-hoang/private-repo-example/raw/main/README.md",
		//"https://bitbucket.org/mike-hoang/private-repo-example/src/main/directory/",
		//"https://bitbucket.org/mike-hoang/private-repo-example/src/main/directory/test.txt",
		//"https://bitbucket.org/mike-hoang/private-repo-example/main/directory/test.txt",
		//"https://opensource.ncsa.illinois.edu/bitbucket/projects/U3D/repos/3dutilities/browse",

		"https://gitlab.com/gitlab-org/cloud-native/gitlab-operator",
		"https://gitlab.com/personal2464/private-repo-example",
		"https://gitlab.com/personal2464/private-repo-example/-/raw/main/README.md",
		"https://gitlab.com/personal2464/private-repo-example/-/tree/main/directory",
		"https://gitlab.com/personal2464/private-repo-example/-/blob/main/directory/text.txt",
		"https://gitlab.com/gitlab-org/gitlab-foss",
	}

	for _, giturl := range giturls {
		g, err := ParseGitUrl(giturl)
		fmt.Printf("url: %s\n%v\n", giturl, g)
		fmt.Println("token: ", g.token)
		if err != nil {
			fmt.Printf("err: %s\n", err)
		}
		fmt.Println()
	}
}

func TestCloneGitRepo(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create temp dir: %s, error: %v", tempDir, err)
	}
	defer os.RemoveAll(tempDir)

	invalidGitUrl := GitUrl{
		Protocol: "",
		Host:     "",
		Owner:    "nonexistent",
		Repo:     "nonexistent",
		Branch:   "nonexistent",
	}

	//validGitHubUrl := GitUrl{
	//	Protocol: "https",
	//	Host:     "github.com",
	//	Owner:    "devfile",
	//	Repo:     "library",
	//	Branch:   "main",
	//}

	//https://gitlab.com/redhat/centos-stream/docs/enterprise-docs
	//https://gitlab.com/gitlab-org/gitlab
	//https://gitlab.com/gitlab-org/cloud-native/gitlab-operator.git
	//https://github.com/whilp/git-urls.git
	validGitLabUrl := GitUrl{
		Protocol: "https",
		Host:     "gitlab.com",
		Owner:    "gitlab-org",
		Repo:     "cloud-native/gitlab-operator",
		Branch:   "master",
	}

	tests := []struct {
		name    string
		gitUrl  GitUrl
		destDir string
		wantErr bool
	}{
		{
			name:    "should fail with invalid git url",
			gitUrl:  invalidGitUrl,
			destDir: filepath.Join(os.TempDir(), "nonexistent"),
			wantErr: true,
		},
		//{
		//	name:    "should be able to clone valid github url",
		//	gitUrl:  validGitHubUrl,
		//	destDir: tempDir,
		//	wantErr: false,
		//},
		{
			name:    "should be able to clone valid gitlab url",
			gitUrl:  validGitLabUrl,
			destDir: tempDir,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CloneGitRepo(tt.gitUrl, tt.destDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error: %t, got error: %t", tt.wantErr, err)
			}
		})
	}
}
