package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
)

type BitBucketAPI struct {
	url                *url.URL
	username, password string
	client             *http.Client
}

func NewBitBucketAPI(urlStr, username, password string) (*BitBucketAPI, error) {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, err
	}
	if len(username) == 0 {
		return nil, errors.New("missing username")
	}
	if len(password) == 0 {
		return nil, errors.New("missing password")
	}
	return &BitBucketAPI{
		url:      parsedURL,
		username: username,
		password: password,
		client:   http.DefaultClient,
	}, nil
}

var userAgent = fmt.Sprintf("bitbucket-server-pr-release/%s (%s)", "1.0", runtime.Version())

func (api *BitBucketAPI) newRequest(ctx context.Context, method, spath string, query map[string]string, body map[string]interface{}) (*http.Request, error) {
	u := *api.url
	u.Path = path.Join(api.url.Path, spath)
	q := u.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	buffer := new(bytes.Buffer)
	if body != nil {
		encoder := json.NewEncoder(buffer)
		if err := encoder.Encode(body); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buffer)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.SetBasicAuth(api.username, api.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}

func (api *BitBucketAPI) GetPullRequests(ctx context.Context, project, repository string) (<-chan PullRequest, <-chan error) {
	out := make(chan PullRequest)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)

		isLastPage := false
		pageStart := 0
		for !isLastPage {
			spath := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests", project, repository)
			req, err := api.newRequest(ctx, "GET", spath, map[string]string{"start": strconv.Itoa(pageStart)}, nil)
			if err != nil {
				errc <- err
				return
			}

			res, err := api.client.Do(req)
			if err != nil {
				errc <- err
				return
			}

			if res.StatusCode != 200 {
				errc <- makeUnexpectedResponseError(res)
				return
			}

			var page PagedValues
			if err := decodeBody(res, &page); err != nil {
				errc <- err
				return
			}
			isLastPage = page.IsLastPage
			if !isLastPage {
				pageStart = page.NextPageStart
			}
			var pullRequests []PullRequest
			if err := mapstructure.Decode(page.Values, &pullRequests); err != nil {
				errc <- err
				return
			}
			for _, pullRequest := range pullRequests {
				out <- pullRequest
			}
		}
	}()

	return out, errc
}

func (api *BitBucketAPI) GetPullRequest(ctx context.Context, project, repository string, pullRequestID int) (*PullRequest, error) {
	spath := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests/%d", project, repository, pullRequestID)
	req, err := api.newRequest(ctx, "GET", spath, map[string]string{}, nil)
	if err != nil {
		return nil, err
	}

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, makeUnexpectedResponseError(res)
	}

	var pullRequest PullRequest
	if err := decodeBody(res, &pullRequest); err != nil {
		return nil, err
	}

	return &pullRequest, nil
}

func (api *BitBucketAPI) CreatePullRequest(ctx context.Context, project, repository, from, to string) (*PullRequest, error) {
	spath := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests", project, repository)
	req, err := api.newRequest(ctx, "POST", spath, map[string]string{}, map[string]interface{}{
		"title":  "Temporary pull-request for release",
		"state":  "OPEN",
		"open":   true,
		"closed": false,
		"fromRef": map[string]string{
			"id": "refs/heads/" + from,
		},
		"toRef": map[string]string{
			"id": "refs/heads/" + to,
		},
		"locked": false,
	})
	if err != nil {
		return nil, err
	}

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 201 {
		return nil, makeUnexpectedResponseError(res)
	}

	var pullRequest PullRequest
	if err := decodeBody(res, &pullRequest); err != nil {
		return nil, err
	}

	return &pullRequest, nil
}

func (api *BitBucketAPI) UpdatePullRequest(ctx context.Context, project, repository string, pullRequest *PullRequest, update map[string]interface{}) (*PullRequest, error) {
	spath := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests/%d", project, repository, pullRequest.ID)
	req, err := api.newRequest(ctx, "PUT", spath, map[string]string{}, update)
	if err != nil {
		return nil, err
	}

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, makeUnexpectedResponseError(res)
	}

	var updated PullRequest
	if err := decodeBody(res, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

func (api *BitBucketAPI) GetPullRequestCommits(ctx context.Context, project, repository string, pullRequest *PullRequest) (<-chan Commit, <-chan error) {
	out := make(chan Commit)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)

		isLastPage := false
		pageStart := 0
		for !isLastPage {
			spath := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests/%d/commits", project, repository, pullRequest.ID)
			req, err := api.newRequest(ctx, "GET", spath, map[string]string{"start": strconv.Itoa(pageStart)}, nil)
			if err != nil {
				errc <- err
				return
			}

			res, err := api.client.Do(req)
			if err != nil {
				errc <- err
				return
			}

			if res.StatusCode != 200 {
				errc <- makeUnexpectedResponseError(res)
				return
			}

			var page PagedValues
			if err := decodeBody(res, &page); err != nil {
				errc <- err
				return
			}
			isLastPage = page.IsLastPage
			if !isLastPage {
				pageStart = page.NextPageStart
			}
			var commits []Commit
			if err := mapstructure.Decode(page.Values, &commits); err != nil {
				errc <- err
				return
			}
			for _, commit := range commits {
				out <- commit
			}
		}
	}()

	return out, errc
}

func makeUnexpectedResponseError(res *http.Response) error {
	bodyBuf := new(bytes.Buffer)
	_, _ = bodyBuf.ReadFrom(res.Body)
	bodyStr := bodyBuf.String()
	return errors.New(fmt.Sprintf("unexpected response: [%d] %s", res.StatusCode, bodyStr))
}
