package main

type PagedValues struct {
	Size          int                      `json:"size"`
	Limit         int                      `json:"limit"`
	IsLastPage    bool                     `json:"isLastPage"`
	Values        []map[string]interface{} `json:"values"`
	Start         int                      `json:"start"`
	NextPageStart int                      `json:"nextPageStart"`
}

// Thanks to [github.com/gfleury/go-bitbucket-v1](https://github.com/gfleury/go-bitbucket-v1/blob/master/api_response.go)
/*
MIT License

Copyright (c) 2018 George S. Fleury

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

type SelfLink struct {
	Href string `json:"href"`
}

type CloneLink struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type Links struct {
	Self []SelfLink `json:"self"`
}

type Project struct {
	Key         string `json:"key"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Type        string `json:"type"`
	Links       Links  `json:"links"`
}

type Repository struct {
	Slug          string  `json:"slug"`
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	ScmID         string  `json:"scmId"`
	State         string  `json:"state"`
	StatusMessage string  `json:"statusMessage"`
	Forkable      bool    `json:"forkable"`
	Project       Project `json:"project"`
	Public        bool    `json:"public"`
	Links         struct {
		Clone []CloneLink `json:"clone"`
		Self  []SelfLink  `json:"self"`
	} `json:"links"`
}

type UserWithNameEmail struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
}

type UserWithLinks struct {
	Name        string `json:"name"`
	Email       string `json:"emailAddress"`
	ID          int    `json:"id"`
	DisplayName string `json:"displayName"`
	Active      bool   `json:"active"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Links       Links  `json:"links"`
}

type User struct {
	Name        string `json:"name"`
	Email       string `json:"emailAddress"`
	ID          int    `json:"id"`
	DisplayName string `json:"displayName"`
	Active      bool   `json:"active"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
}

type UserWithMetadata struct {
	User     UserWithLinks `json:"user"`
	Role     string        `json:"role"`
	Approved bool          `json:"approved"`
	Status   string        `json:"status"`
}

type MergeResult struct {
	Outcome string `json:"outcome"`
	Current bool   `json:"current"`
}

type MergeGetResponse struct {
	CanMerge   bool                     `json:"canMerge"`
	Conflicted bool                     `json:"conflicted"`
	Outcome    string                   `json:"outcome"`
	Vetoes     []map[string]interface{} `json:"vetoes"`
}

type PullRequestRef struct {
	ID           string     `json:"id"`
	DisplayID    string     `json:"displayId"`
	LatestCommit string     `json:"latestCommit"`
	Repository   Repository `json:"repository"`
}

type PullRequest struct {
	ID           int                `json:"id"`
	Version      int32              `json:"version"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	State        string             `json:"state"`
	Open         bool               `json:"open"`
	Closed       bool               `json:"closed"`
	CreatedDate  int64              `json:"createdDate"`
	UpdatedDate  int64              `json:"updatedDate"`
	FromRef      PullRequestRef     `json:"fromRef"`
	ToRef        PullRequestRef     `json:"toRef"`
	Locked       bool               `json:"locked"`
	Author       UserWithMetadata   `json:"author"`
	Reviewers    []UserWithMetadata `json:"reviewers"`
	Participants []UserWithMetadata `json:"participants"`
	Properties   struct {
		MergeResult       MergeResult `json:"mergeResult"`
		ResolvedTaskCount int         `json:"resolvedTaskCount"`
		OpenTaskCount     int         `json:"openTaskCount"`
	} `json:"properties"`
	Links Links `json:"links"`
}

type SSHKey struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Label string `json:"label"`
}

type Commit struct {
	ID                 string `json:"id"`
	DisplayID          string `json:"displayId"`
	Author             User   `json:"author"`
	AuthorTimestamp    int64  `json:"authorTimestamp"`
	Committer          User   `json:"committer"`
	CommitterTimestamp int64  `json:"committerTimestamp"`
	Message            string `json:"message"`
	Parents            []struct {
		ID        string `json:"id"`
		DisplayID string `json:"displayId"`
	} `json:"parents"`
}

type BuildStatus struct {
	State       string `json:"state"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
	DateAdded   int64  `json:"dateAdded"`
}

type Diff struct {
	Diffs []struct {
		Source struct {
			Components []string `json:"components"`
			Parent     string   `json:"parent"`
			Name       string   `json:"name"`
			Extension  string   `json:"extension"`
			ToString   string   `json:"toString"`
		} `json:"source"`
		Destination struct {
			Components []string `json:"components"`
			Parent     string   `json:"parent"`
			Name       string   `json:"name"`
			Extension  string   `json:"extension"`
			ToString   string   `json:"toString"`
		} `json:"destination"`
		Hunks []struct {
			SourceLine      int `json:"sourceLine"`
			SourceSpan      int `json:"sourceSpan"`
			DestinationLine int `json:"destinationLine"`
			DestinationSpan int `json:"destinationSpan"`
			Segments        []struct {
				Type  string `json:"type"`
				Lines []struct {
					Destination int    `json:"destination"`
					Source      int    `json:"source"`
					Line        string `json:"line"`
					Truncated   bool   `json:"truncated"`
				} `json:"lines"`
				Truncated bool `json:"truncated"`
			} `json:"segments"`
			Truncated bool `json:"truncated"`
		} `json:"hunks"`
		Truncated    bool    `json:"truncated"`
		ContextLines float64 `json:"contextLines"`
		FromHash     string  `json:"fromHash"`
		ToHash       string  `json:"toHash"`
		WhiteSpace   string  `json:"whiteSpace"`
	} `json:"diffs"`
	Truncated    bool    `json:"truncated"`
	ContextLines float64 `json:"contextLines"`
	FromHash     string  `json:"fromHash"`
	ToHash       string  `json:"toHash"`
	WhiteSpace   string  `json:"whiteSpace"`
}

type Tag struct {
	ID              string `json:"id"`
	DisplayID       string `json:"displayId"`
	Type            string `json:"type"`
	LatestCommit    string `json:"latestCommit"`
	LatestChangeset string `json:"latestChangeset"`
	Hash            string `json:"hash"`
}

type Branch struct {
	ID              string `json:"id"`
	DisplayID       string `json:"displayId"`
	Type            string `json:"type"`
	LatestCommit    string `json:"latestCommit"`
	LatestChangeset string `json:"latestChangeset"`
	IsDefault       bool   `json:"isDefault"`
}
