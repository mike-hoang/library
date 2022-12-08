//
// Copyright 2022 Red Hat, Inc.
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
	gitpkg "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	giturl "github.com/whilp/git-urls"
	"net/url"
	"strings"
)

const (
	DEFAULT_HOST string = "github.com"
	RAW_HOST     string = "raw.githubusercontent.com"
)

type GitHubUrl struct {
	host   string
	branch string
	path   string
	token  string
}

/*
// NewGitURL get instance of git parser
func NewGitURL(fullURL string) (IGitURL, error) {
	hostUrl, err := getHost(fullURL)
	if err != nil {
		return nil, err
	}

	if githubparserv1.IsHostGitHub(hostUrl) {
		return githubparserv1.NewGitHubParserWithURL(fullURL)
	}
	if azureparserv1.IsHostAzure(hostUrl) {
		return azureparserv1.NewAzureParserWithURL(fullURL)
	}
	if gitlabparserv1.IsHostGitLab(hostUrl) {
		return gitlabparserv1.NewGitLabParserWithURL(fullURL)
	}
	return nil, fmt.Errorf("repository host '%s' not supported", hostUrl)
}
*/

func isHostGithub(host string) bool {
	return host == DEFAULT_HOST || host == RAW_HOST
}

func ParseGitUrl(url string) error {
	parsedUrl, err := giturl.Parse(url)
	if err != nil {
		return err
	}

	fmt.Println(parsedUrl)
	return err
}

// GetGitUrlComponentsFromRaw converts a raw GitHub file link to a map of the url components
func GetGitUrlComponentsFromRaw(rawGitURL string) (map[string]string, error) {
	var urlComponents map[string]string

	err := ValidateURL(rawGitURL)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(rawGitURL)
	// the url scheme (e.g. https://) is removed before splitting into the 5 components
	urlPath := strings.SplitN(u.Host+u.Path, "/", 5)

	// raw GitHub url: https://raw.githubusercontent.com/devfile/registry/main/stacks/nodejs/devfile.yaml
	// host: raw.githubusercontent.com
	// username: devfile
	// project: registry
	// branch: main
	// file: stacks/nodejs/devfile.yaml
	if len(urlPath) == 5 {
		urlComponents = map[string]string{
			"host":     urlPath[0],
			"username": urlPath[1],
			"project":  urlPath[2],
			"branch":   urlPath[3],
			"file":     urlPath[4],
		}
	}

	return urlComponents, nil
}

// CloneGitRepo clones a GitHub repo to a destination directory
func CloneGitRepo(gitUrlComponents map[string]string, destDir string) error {
	gitUrl := fmt.Sprintf("https://github.com/%s/%s.git", gitUrlComponents["username"], gitUrlComponents["project"])
	branch := fmt.Sprintf("refs/heads/%s", gitUrlComponents["branch"])

	cloneOptions := &gitpkg.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.ReferenceName(branch),
		SingleBranch:  true,
		Depth:         1,
		//Auth: &http.BasicAuth{
		//	Username: "user",
		//	Password: gitUrlComponents["token"],
		//},
	}

	_, err := gitpkg.PlainClone(destDir, false, cloneOptions)
	if err != nil {
		return err
	}
	return nil
}
