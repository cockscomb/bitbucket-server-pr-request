package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func main() {
	logger := log.New(os.Stderr, "", 0)

	var (
		bitbucketServer = flag.String("bitbucket-server", "", "BitBucket Server `URL`")
		username        = flag.String("username", "", "BitBucket Server username")
		password        = flag.String("password", "", "BitBucket Server password")
		projectName     = flag.String("project", "", "Project name")
		repositoryName  = flag.String("repository", "", "Repository name")
		fromBranchName  = flag.String("from", "", "from `branch name`")
		toBranchName    = flag.String("to", "", "to `branch name`")
	)
	flag.Parse()

	if len(*bitbucketServer) == 0 ||
		len(*username) == 0 ||
		len(*password) == 0 ||
		len(*projectName) == 0 ||
		len(*repositoryName) == 0 ||
		len(*fromBranchName) == 0 ||
		len(*toBranchName) == 0 {
		flag.PrintDefaults()
		logger.Fatalln("Required parameters are missing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	api, err := NewBitBucketAPI(*bitbucketServer, *username, *password)
	if err != nil {
		logger.Fatalln(err)
	}

	// Find release pull-request
	prChan, errChan := api.GetPullRequests(ctx, *projectName, *repositoryName)
	releasePullRequest := func() *PullRequest {
		for {
			select {
			case pr, ok := <-prChan:
				if !ok {
					return nil
				}
				if pr.FromRef.DisplayID == *fromBranchName && pr.ToRef.DisplayID == *toBranchName && pr.Open {
					return &pr
				}
			case err, ok := <-errChan:
				if ok {
					logger.Fatalln(err)
				}
			}
		}
	}()

	// Create new release pull-request
	if releasePullRequest == nil {
		pullRequest, err := api.CreatePullRequest(ctx, *projectName, *repositoryName, *fromBranchName, *toBranchName)
		if err != nil {
			logger.Fatalln(err)
		}
		releasePullRequest = pullRequest
	}

	// Find merge commits
	commitChan, errChan := api.GetPullRequestCommits(ctx, *projectName, *repositoryName, releasePullRequest)
	mergeCommits := func() []Commit {
		commits := make([]Commit, 0)
		for {
			select {
			case commit, ok := <-commitChan:
				if !ok {
					return commits
				}
				if strings.HasPrefix(commit.Message, "Merge pull request ") {
					commits = append(commits, commit)
				}
			case err, ok := <-errChan:
				if ok {
					logger.Fatalln(err)
				}
			}
		}
	}()
	pullRequestIDs := make([]int, 0)
	mergeCommitMessage := regexp.MustCompile(`^Merge pull request #(\d+) `)
	for _, commit := range mergeCommits {
		group := mergeCommitMessage.FindStringSubmatch(commit.Message)
		if len(group) <= 1 {
			logger.Fatalln(errors.New("cannot extract pull-request ID"))
		}
		prID, err := strconv.Atoi(group[1])
		if err != nil {
			logger.Fatalln(err)
		}
		pullRequestIDs = append(pullRequestIDs, prID)
	}

	// Retrieve merged pull-requests
	mergedPullRequests := make([]PullRequest, 0)
	for _, prID := range pullRequestIDs {
		pr, err := api.GetPullRequest(ctx, *projectName, *repositoryName, prID)
		if err != nil {
			logger.Fatalln(err)
		}
		mergedPullRequests = append(mergedPullRequests, *pr)
	}

	// Compose description
	type DescriptionParameters struct {
		PullRequests []PullRequest
	}
	params := DescriptionParameters{PullRequests: mergedPullRequests}
	tmpl, err := template.New("description").Parse(`{{range $pr := .PullRequests}}
- #{{$pr.ID}} {{$pr.Title}}{{end}}`)
	if err != nil {
		logger.Fatalln(err)
	}
	descBuf := new(bytes.Buffer)
	if err := tmpl.Execute(descBuf, params); err != nil {
		logger.Fatalln(err)
	}
	description := descBuf.String()

	// Update pull-request
	now := time.Now()
	releasePullRequest, err = api.UpdatePullRequest(ctx, *projectName, *repositoryName, releasePullRequest, map[string]interface{}{
		"version":     releasePullRequest.Version,
		"title":       fmt.Sprintf("Release (%s)", now.Format("2006/01/02 15:04:05 JST")),
		"description": description,
	})
	if err != nil {
		logger.Fatalln(err)
	}
	fmt.Println(path.Join(*bitbucketServer, fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", *projectName, *repositoryName, releasePullRequest.ID)))
}
