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
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	HTTPRequestResponseTimeout = 30 * time.Second // HTTPRequestTimeout configures timeout of all HTTP requests
)

// httpCacheDir determines directory where odo will cache HTTP responses
var httpCacheDir = filepath.Join(os.TempDir(), "odohttpcache")

// HTTPRequestParams holds parameters of forming http request
type HTTPRequestParams struct {
	URL                 string
	Token               string
	Timeout             *int
	TelemetryClientName string //optional client name for telemetry
}

// HTTPGetRequest gets resource contents given URL and token (if applicable)
// cacheFor determines how long the response should be cached (in minutes), 0 for no caching
func HTTPGetRequest(request HTTPRequestParams, cacheFor int) ([]byte, error) {
	// Build http request
	req, err := http.NewRequest("GET", request.URL, nil)
	if err != nil {
		return nil, err
	}
	if request.Token != "" {
		bearer := "Bearer " + request.Token
		req.Header.Add("Authorization", bearer)
	}

	//add the telemetry client name
	req.Header.Add("Client", request.TelemetryClientName)

	overriddenTimeout := HTTPRequestResponseTimeout
	timeout := request.Timeout
	if timeout != nil {
		//if value is invalid, the default will be used
		if *timeout > 0 {
			//convert timeout to seconds
			overriddenTimeout = time.Duration(*timeout) * time.Second
			klog.V(4).Infof("HTTP request and response timeout overridden value is %v ", overriddenTimeout)
		} else {
			klog.V(4).Infof("Invalid httpTimeout is passed in, using default value")
		}

	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ResponseHeaderTimeout: overriddenTimeout,
		},
		Timeout: overriddenTimeout,
	}

	klog.V(4).Infof("HTTPGetRequest: %s", req.URL.String())

	if cacheFor > 0 {
		// if there is an error during cache setup we show warning and continue without using cache
		cacheError := false
		httpCacheTime := time.Duration(cacheFor) * time.Minute

		// make sure that cache directory exists
		err = os.MkdirAll(httpCacheDir, 0750)
		if err != nil {
			cacheError = true
			klog.WarningDepth(4, "Unable to setup cache: ", err)
		}
		err = cleanHttpCache(httpCacheDir, httpCacheTime)
		if err != nil {
			cacheError = true
			klog.WarningDepth(4, "Unable to clean up cache directory: ", err)
		}

		if !cacheError {
			httpClient.Transport = httpcache.NewTransport(diskcache.New(httpCacheDir))
			klog.V(4).Infof("Response will be cached in %s for %s", httpCacheDir, httpCacheTime)
		} else {
			klog.V(4).Info("Response won't be cached.")
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Header.Get(httpcache.XFromCache) != "" {
		klog.V(4).Infof("Cached response used.")
	}

	// We have a non 1xx / 2xx status, return an error
	if (resp.StatusCode - 300) > 0 {
		return nil, errors.Errorf("failed to retrieve %s, %v: %s", request.URL, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Process http response
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes, err
}

// ValidateURL validates the URL
func ValidateURL(sourceURL string) error {
	u, err := url.Parse(sourceURL)
	if err != nil {
		return err
	}

	if len(u.Host) == 0 || len(u.Scheme) == 0 {
		return errors.New("URL is invalid")
	}

	return nil
}

// cleanHttpCache checks cacheDir and deletes all files that were modified more than cacheTime back
func cleanHttpCache(cacheDir string, cacheTime time.Duration) error {
	cacheFiles, err := ioutil.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, f := range cacheFiles {
		if f.ModTime().Add(cacheTime).Before(time.Now()) {
			klog.V(4).Infof("Removing cache file %s, because it is older than %s", f.Name(), cacheTime.String())
			err := os.Remove(filepath.Join(cacheDir, f.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CheckPathExists checks if a path exists or not
func CheckPathExists(path string) bool {
	return checkPathExistsOnFS(path, filesystem.DefaultFs{})
}

func checkPathExistsOnFS(path string, fs filesystem.Filesystem) bool {
	if _, err := fs.Stat(path); !os.IsNotExist(err) {
		// path to file does exist
		return true
	}
	klog.V(4).Infof("path %s doesn't exist, skipping it", path)
	return false
}
