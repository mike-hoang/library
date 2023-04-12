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

package git

import (
	"fmt"
	"net/http"
	"net/url"
)

var (
	GetDoFunc                func(req *http.Request) (*http.Response, error)
	GetNewGitUrlFunc         func(url string) (*MockGitUrl, error)
	GetParseGitUrlFunc       func(url string) error
	GetGitRawFileAPIFunc     func() string
	GetSetTokenFunc          func(token string, httpTimeout *int) error
	GetIsPublicFunc          func(httpTimeout *int) bool
	GetIsGitProviderRepoFunc func() bool
	GetCloneGitRepoFunc      func(destDir string) error
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

type MockGitUrl struct {
	//NewGitUrlFunc func(url string) (*MockGitUrl, error)
	//ParseGitUrlFunc          func(fullUrl string) error
	//GetGitRawFileAPIFunc     func() string
	//SetTokenFunc             func(token string, httpTimeout *int) error
	//IsPublicFunc             func(httpTimeout *int) bool
	//GetIsGitProviderRepoFunc func(url string) bool
	//CloneGitRepoFunc         func(destDir string) error
	Protocol string // URL scheme
	Host     string // URL domain name
	Owner    string // name of the repo owner
	Repo     string // name of the repo
	Branch   string // branch name
	Path     string // path to a directory or file in the repo
	token    string // used for authenticating a private repo
	IsFile   bool   // defines if the URL points to a file in the repo
}

var mockExecute = func(baseDir string, cmd CommandType, args ...string) ([]byte, error) {
	if cmd == GitCommand {
		fmt.Println(url.Parse(args[1]))
		//c := exec.Command(string(cmd), args...)
		//c.Dir = baseDir
		//output, err := c.CombinedOutput()
		output := []byte("test")
		return output, nil
	}

	return []byte(""), fmt.Errorf(unsupportedCmdMsg, string(cmd))
}

func MockNewGitUrl(url string) (*MockGitUrl, error) {
	//return GetNewGitUrlFunc(url)
	return &MockGitUrl{
		Protocol: "https",
		Host:     "github.com",
		Owner:    "devfile",
		Repo:     "registry",
		Branch:   "main",
		Path:     "stacks/go/1.0.2/devfile.yaml",
		IsFile:   true,
	}, nil
}

func (m *MockGitUrl) CloneGitRepo(destDir string) error {
	//return GetCloneGitRepoFunc(destDir)
	fmt.Println("test")
	exist := CheckPathExists(destDir)
	if !exist {
		return fmt.Errorf("failed to clone repo, destination directory: '%s' does not exists", destDir)
	}

	host := m.Host
	if host == RawGitHubHost {
		host = GitHubHost
	}

	var repoUrl string
	if m.token == "" {
		repoUrl = fmt.Sprintf("%s://%s/%s/%s.git", m.Protocol, host, m.Owner, m.Repo)
	} else {
		repoUrl = fmt.Sprintf("%s://token:%s@%s/%s/%s.git", m.Protocol, m.token, host, m.Owner, m.Repo)
		if m.Host == BitbucketHost {
			repoUrl = fmt.Sprintf("%s://x-token-auth:%s@%s/%s/%s.git", m.Protocol, m.token, host, m.Owner, m.Repo)
		}
	}

	//c, err := execute("git", "clone", repoUrl, destDir)
	c, err := execute(destDir, "git", "clone", repoUrl)
	fmt.Println("[clone repo] c: ", string(c))
	fmt.Println("[clone repo] err: ", err)

	///* #nosec G204 -- user input is processed into an expected format for the git clone command */
	//c := exec.Command("git", "clone", repoUrl, destDir)
	//c.Dir = destDir
	//
	//// set env to skip authentication prompt and directly error out
	//c.Env = os.Environ()
	//c.Env = append(c.Env, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=/bin/echo")
	//
	//_, err := c.CombinedOutput()
	if err != nil {
		if m.token == "" {
			return fmt.Errorf("failed to clone repo without a token, ensure that a token is set if the repo is private. error: %v", err)
		} else {
			return fmt.Errorf("failed to clone repo with token, ensure that the url and token is correct. error: %v", err)
		}
	}

	return nil
}

func (m *MockGitUrl) ParseGitUrl(fullUrl string) error {
	return GetParseGitUrlFunc(fullUrl)
}

func (m *MockGitUrl) GitRawFileAPI() string {
	return GetGitRawFileAPIFunc()
}

func (m *MockGitUrl) SetToken(token string, httpTimeout *int) error {
	return GetSetTokenFunc(token, httpTimeout)
}

func (m *MockGitUrl) IsPublic(httpTimeout *int) bool {
	return GetIsPublicFunc(httpTimeout)
}

func (m *MockGitUrl) IsGitProviderRepo() bool {
	return GetIsGitProviderRepoFunc()
}
