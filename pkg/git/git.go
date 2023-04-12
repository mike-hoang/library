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
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	GitHubHost    string = "github.com"
	RawGitHubHost string = "raw.githubusercontent.com"
	GitLabHost    string = "gitlab.com"
	BitbucketHost string = "bitbucket.org"
)

//type IGitUrl interface {
//	NewGitUrl(url string) (*Url, error)
//	ParseGitUrl(fullUrl string) error
//	GitRawFileAPI() string
//	SetToken(token string, httpTimeout *int) error
//	IsPublic(httpTimeout *int) bool
//	IsGitProviderRepo() bool
//	CloneGitRepo(destDir string) error
//}

type Url struct {
	Protocol string // URL scheme
	Host     string // URL domain name
	Owner    string // name of the repo owner
	Repo     string // name of the repo
	Branch   string // branch name
	Path     string // path to a directory or file in the repo
	token    string // used for authenticating a private repo
	IsFile   bool   // defines if the URL points to a file in the repo
}

type CommandType string

const (
	GitCommand        CommandType = "git"
	unsupportedCmdMsg             = "Unsupported command \"%s\" "
)

// Execute is exposed as a global variable for the purpose of running mock tests
// only "git" is supported
/* #nosec G204 -- used internally to execute various git actions and eventual cleanup of artifacts.  Calling methods validate user input to ensure commands are used appropriately */
var execute = func(baseDir string, cmd CommandType, args ...string) ([]byte, error) {
	if cmd == GitCommand {
		c := exec.Command(string(cmd), args...)
		c.Dir = baseDir
		output, err := c.CombinedOutput()
		return output, err
	}

	return []byte(""), fmt.Errorf(unsupportedCmdMsg, string(cmd))
}

////NewGitUrl creates a GitUrl from a string url
//func NewGitUrl(url string) (*Url, error) {
//	var g = Url{}
//	if err := g.ParseGitUrl(url); err != nil {
//		return &g, err
//	}
//	return &g, nil
//}

//NewGitUrl creates a GitUrl from a string url
func (g *Url) NewGitUrl(url string) (*Url, error) {
	if err := g.ParseGitUrl(url); err != nil {
		return g, err
	}
	return g, nil
}

// ParseGitUrl extracts information from a support git url
// Only supports git repositories hosted on GitHub, GitLab, and Bitbucket
func (g *Url) ParseGitUrl(fullUrl string) error {
	err := ValidateURL(fullUrl)
	if err != nil {
		return err
	}

	parsedUrl, err := url.Parse(fullUrl)
	if err != nil {
		return err
	}

	if len(parsedUrl.Path) == 0 {
		return fmt.Errorf("url path should not be empty")
	}

	if parsedUrl.Host == RawGitHubHost || parsedUrl.Host == GitHubHost {
		err = g.parseGitHubUrl(parsedUrl)
	} else if parsedUrl.Host == GitLabHost {
		err = g.parseGitLabUrl(parsedUrl)
	} else if parsedUrl.Host == BitbucketHost {
		err = g.parseBitbucketUrl(parsedUrl)
	} else {
		err = fmt.Errorf("url host should be a valid GitHub, GitLab, or Bitbucket host; received: %s", parsedUrl.Host)
	}

	return err
}

func (g *Url) parseGitHubUrl(url *url.URL) error {
	var splitUrl []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host

	if g.Host == RawGitHubHost {
		g.IsFile = true
		// raw GitHub urls don't contain "blob" or "tree"
		splitUrl = strings.SplitN(url.Path[1:], "/", 4)
		if len(splitUrl) == 4 {
			g.Owner = splitUrl[0]
			g.Repo = splitUrl[1]
			g.Branch = splitUrl[2]
			g.Path = splitUrl[3]
		} else {
			// raw GitHub urls have to be a file
			err = fmt.Errorf("raw url path should contain <owner>/<repo>/<branch>/<path/to/file>, received: %s", url.Path[1:])
		}
		return err
	}

	if g.Host == GitHubHost {
		splitUrl = strings.SplitN(url.Path[1:], "/", 5)
		if len(splitUrl) < 2 {
			err = fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
		} else {
			g.Owner = splitUrl[0]
			g.Repo = splitUrl[1]

			// url doesn't contain a path to a directory or file
			if len(splitUrl) == 2 {
				return nil
			}

			switch splitUrl[2] {
			case "tree":
				g.IsFile = false
			case "blob":
				g.IsFile = true
			default:
				return fmt.Errorf("url path to directory or file should contain 'tree' or 'blob'")
			}

			// url has a path to a file or directory
			if len(splitUrl) == 5 {
				g.Branch = splitUrl[3]
				g.Path = splitUrl[4]
			} else {
				err = fmt.Errorf("url path should contain <owner>/<repo>/<tree or blob>/<branch>/<path/to/file/or/directory>, received: %s", url.Path[1:])
			}
		}
	}

	return err
}

func (g *Url) parseGitLabUrl(url *url.URL) error {
	var splitFile, splitOrg []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.IsFile = false

	// GitLab urls contain a '-' separating the root of the repo
	// and the path to a file or directory
	split := strings.Split(url.Path[1:], "/-/")

	splitOrg = strings.SplitN(split[0], "/", 2)
	if len(splitOrg) < 2 {
		return fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
	} else {
		g.Owner = splitOrg[0]
		g.Repo = splitOrg[1]
	}

	// url doesn't contain a path to a directory or file
	if len(split) == 1 {
		return nil
	}

	// url may contain a path to a directory or file
	if len(split) == 2 {
		splitFile = strings.SplitN(split[1], "/", 3)
	}

	if len(splitFile) == 3 {
		if splitFile[0] == "blob" || splitFile[0] == "tree" || splitFile[0] == "raw" {
			g.Branch = splitFile[1]
			g.Path = splitFile[2]
			ext := filepath.Ext(g.Path)
			if ext != "" {
				g.IsFile = true
			}
		} else {
			err = fmt.Errorf("url path should contain 'blob' or 'tree' or 'raw', received: %s", url.Path[1:])
		}
	} else {
		return fmt.Errorf("url path to directory or file should contain <blob or tree or raw>/<branch>/<path/to/file/or/directory>, received: %s", url.Path[1:])
	}

	return err
}

func (g *Url) parseBitbucketUrl(url *url.URL) error {
	var splitUrl []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.IsFile = false

	splitUrl = strings.SplitN(url.Path[1:], "/", 5)
	if len(splitUrl) < 2 {
		err = fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
	} else if len(splitUrl) == 2 {
		g.Owner = splitUrl[0]
		g.Repo = splitUrl[1]
	} else {
		g.Owner = splitUrl[0]
		g.Repo = splitUrl[1]
		if len(splitUrl) == 5 {
			if splitUrl[2] == "raw" || splitUrl[2] == "src" {
				g.Branch = splitUrl[3]
				g.Path = splitUrl[4]
				ext := filepath.Ext(g.Path)
				if ext != "" {
					g.IsFile = true
				}
			} else {
				err = fmt.Errorf("url path should contain 'raw' or 'src', received: %s", url.Path[1:])
			}
		} else {
			err = fmt.Errorf("url path should contain path to directory or file, received: %s", url.Path[1:])
		}
	}

	return err
}

// SetToken validates the token with a get request to the repo before setting the token
// Defaults token to empty on failure.
func (g *Url) SetToken(token string, httpTimeout *int) error {
	err := g.validateToken(HTTPRequestParams{Token: token, Timeout: httpTimeout})
	if err != nil {
		g.token = ""
		return fmt.Errorf("failed to set token. error: %v", err)
	}
	g.token = token
	return nil
}

//func (g *Url) GetToken() {
//	return g.token
//}

// IsPublic checks if the GitUrl is public with a get request to the repo using an empty token
// Returns true if the request succeeds
func (g *Url) IsPublic(httpTimeout *int) bool {
	err := g.validateToken(HTTPRequestParams{Token: "", Timeout: httpTimeout})
	if err != nil {
		return false
	}
	return true
}

// validateToken makes a http get request to the repo with the GitUrl token
// Returns an error if the get request fails
func (g *Url) validateToken(params HTTPRequestParams) error {
	var apiUrl string

	switch g.Host {
	case GitHubHost, RawGitHubHost:
		apiUrl = fmt.Sprintf("https://api.github.com/repos/%s/%s", g.Owner, g.Repo)
	case GitLabHost:
		apiUrl = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s", g.Owner, g.Repo)
	case BitbucketHost:
		apiUrl = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", g.Owner, g.Repo)
	default:
		apiUrl = fmt.Sprintf("%s://%s/%s/%s.git", g.Protocol, g.Host, g.Owner, g.Repo)
	}

	params.URL = apiUrl
	res, err := HTTPGetRequest(params, 0)
	if len(res) == 0 || err != nil {
		return err
	}

	return nil
}

// GitRawFileAPI returns the endpoint for the git providers raw file
func (g *Url) GitRawFileAPI() string {
	var apiRawFile string

	switch g.Host {
	case GitHubHost, RawGitHubHost:
		apiRawFile = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", g.Owner, g.Repo, g.Branch, g.Path)
	case GitLabHost:
		apiRawFile = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/repository/files/%s/raw", g.Owner, g.Repo, g.Path)
	case BitbucketHost:
		apiRawFile = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/src/%s/%s", g.Owner, g.Repo, g.Branch, g.Path)
	}

	return apiRawFile
}

//// IsGitProviderRepo checks if the url matches a repo from a supported git provider
//func IsGitProviderRepo(url string) bool {
//	if strings.Contains(url, RawGitHubHost) || strings.Contains(url, GitHubHost) ||
//		strings.Contains(url, GitLabHost) || strings.Contains(url, BitbucketHost) {
//		return true
//	}
//	return false
//}

// IsGitProviderRepo checks if the url matches a repo from a supported git provider
func (g *Url) IsGitProviderRepo() bool {
	switch g.Host {
	case GitHubHost, RawGitHubHost, GitLabHost, BitbucketHost:
		return true
	default:
		return false
	}
}

// CloneGitRepo clones a git repo to a destination directory (either an absolute or relative path)
func (g *Url) CloneGitRepo(destDir string) error {
	exist := CheckPathExists(destDir)
	if !exist {
		return fmt.Errorf("failed to clone repo, destination directory: '%s' does not exists", destDir)
	}

	host := g.Host
	if host == RawGitHubHost {
		host = GitHubHost
	}

	var repoUrl string
	if g.token == "" {
		repoUrl = fmt.Sprintf("%s://%s/%s/%s.git", g.Protocol, host, g.Owner, g.Repo)
	} else {
		repoUrl = fmt.Sprintf("%s://token:%s@%s/%s/%s.git", g.Protocol, g.token, host, g.Owner, g.Repo)
		if g.Host == BitbucketHost {
			repoUrl = fmt.Sprintf("%s://x-token-auth:%s@%s/%s/%s.git", g.Protocol, g.token, host, g.Owner, g.Repo)
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
		if g.token == "" {
			return fmt.Errorf("failed to clone repo without a token, ensure that a token is set if the repo is private. error: %v", err)
		} else {
			return fmt.Errorf("failed to clone repo with token, ensure that the url and token is correct. error: %v", err)
		}
	}

	return nil
}
