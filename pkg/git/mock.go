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
	"io/ioutil"
	"net/url"
	"os"
	"path"
)

type MockGitUrl struct {
	Protocol string // URL scheme
	Host     string // URL domain name
	Owner    string // name of the repo owner
	Repo     string // name of the repo
	Branch   string // branch name
	Path     string // path to a directory or file in the repo
	token    string // used for authenticating a private repo
	IsFile   bool   // defines if the URL points to a file in the repo
}

func MockNewGitUrlWithURL(url string) (*MockGitUrl, error) {
	g, err := ParseGitUrl(url)
	m := ConvertUrlToMockUrl(g)
	return &m, err
}

func (m *MockGitUrl) GetProtocol() string {
	return m.Protocol
}

func (m *MockGitUrl) GetHost() string {
	return m.Host
}

func (m *MockGitUrl) GetOwner() string {
	return m.Owner
}

func (m *MockGitUrl) GetRepo() string {
	return m.Repo
}

func (m *MockGitUrl) GetBranch() string {
	return m.Branch
}

func (m *MockGitUrl) GetPath() string {
	return m.Path
}

func (m *MockGitUrl) GetToken() string {
	return m.token
}

func (m *MockGitUrl) GetIsFile() bool {
	return m.IsFile
}

var mockExecute = func(baseDir string, cmd CommandType, args ...string) ([]byte, error) {
	if cmd == GitCommand {
		// todo: finish this implementation
		// if token is found return no err
		// else return err
		fmt.Println(url.Parse(args[1]))
		//c := exec.Command(string(cmd), args...)
		//c.Dir = baseDir
		//output, err := c.CombinedOutput()
		output := []byte("test")
		return output, nil
	}

	return []byte(""), fmt.Errorf(unsupportedCmdMsg, string(cmd))
}

func (m *MockGitUrl) CloneGitRepo(destDir string) error {
	fmt.Println("test: mock clone")
	exist := CheckPathExists(destDir)
	if !exist {
		return fmt.Errorf("failed to clone repo, destination directory: '%s' does not exists", destDir)
	}

	host := m.GetHost()
	if host == RawGitHubHost {
		host = GitHubHost
	}

	var repoUrl string
	if m.GetToken() == "" {
		repoUrl = fmt.Sprintf("%s://%s/%s/%s.git", m.GetProtocol(), host, m.GetOwner(), m.GetRepo())
	} else {
		repoUrl = fmt.Sprintf("%s://token:%s@%s/%s/%s.git", m.GetProtocol(), m.GetToken(), host, m.GetOwner(), m.GetRepo())
		if m.GetHost() == BitbucketHost {
			repoUrl = fmt.Sprintf("%s://x-token-auth:%s@%s/%s/%s.git", m.GetProtocol(), m.GetToken(), host, m.GetOwner(), m.GetRepo())
		}
	}

	_, err := mockExecute(destDir, "git", "clone", repoUrl, ".")
	fmt.Println("m.Token: ", m.GetToken())
	fmt.Println("[mock execute] repoUrl: ", repoUrl)

	if err != nil {
		if m.GetToken() == "" {
			return fmt.Errorf("failed to clone repo without a token, ensure that a token is set if the repo is private. error: %v", err)
		} else {
			return fmt.Errorf("failed to clone repo with token, ensure that the url and token is correct. error: %v", err)
		}
	}

	return nil
}

func (m *MockGitUrl) DownloadGitRepoResources(url string, destDir string, httpTimeout *int, token string) error {
	//gitUrl, err := MockNewGitUrlWithURL(url)
	gitUrl := m
	if gitUrl.IsGitProviderRepo() && gitUrl.IsFile {
		stackDir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("git-resources"))
		if err != nil {
			return fmt.Errorf("failed to create dir: %s, error: %v", stackDir, err)
		}
		defer os.RemoveAll(stackDir)

		if !gitUrl.IsPublic(httpTimeout) {
			err = m.SetToken(token, httpTimeout)
			if err != nil {
				return err
			}
		}

		err = gitUrl.CloneGitRepo(stackDir)
		if err != nil {
			return err
		}

		dir := path.Dir(path.Join(stackDir, gitUrl.GetPath()))
		err = CopyAllDirFiles(dir, destDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MockGitUrl) SetToken(token string, httpTimeout *int) error {
	m.token = token
	return nil
}

func (m *MockGitUrl) IsPublic(httpTimeout *int) bool {
	return false
}

func (m *MockGitUrl) GitRawFileAPI() string {
	return ""
}

func (m *MockGitUrl) IsGitProviderRepo() bool {
	return true
}

func ConvertUrlToMockUrl(g Url) MockGitUrl {
	m := MockGitUrl{}
	m.Protocol = g.Protocol
	m.Host = g.Host
	m.Owner = g.Owner
	m.Repo = g.Repo
	m.Branch = g.Branch
	m.Path = g.Path
	m.token = g.token
	m.IsFile = g.IsFile
	return m
}
