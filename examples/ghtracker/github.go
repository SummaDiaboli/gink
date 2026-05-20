package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Repo is a GitHub repository summary returned by gh repo list.
type Repo struct {
	Name          string    `json:"name"`
	NameWithOwner string    `json:"nameWithOwner"`
	Description   string    `json:"description"`
	Stars         int       `json:"stargazerCount"`
	IsPrivate     bool      `json:"isPrivate"`
	PushedAt      time.Time `json:"pushedAt"`
}

type ghAuthor struct {
	Login string `json:"login"`
}

type ghLabel struct {
	Name string `json:"name"`
}

// Issue is a GitHub issue summary.
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	Author    ghAuthor  `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	Labels    []ghLabel `json:"labels"`
}

// PR is a GitHub pull request summary.
type PR struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	Author    ghAuthor  `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	IsDraft   bool      `json:"isDraft"`
}

type ghCommitInner struct {
	Message string `json:"message"`
	Author  struct {
		Name string    `json:"name"`
		Date time.Time `json:"date"`
	} `json:"author"`
}

type ghCommitItem struct {
	SHA    string        `json:"sha"`
	Commit ghCommitInner `json:"commit"`
}

// Commit is a Git commit summary.
type Commit struct {
	SHA     string
	Message string
	Author  string
	Date    time.Time
}

// fetchUsername returns the authenticated GitHub username.
func fetchUsername() (string, error) {
	out, err := exec.Command("gh", "api", "user", "--jq", ".login").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// fetchRepos lists the authenticated user's repositories sorted by last push.
func fetchRepos() ([]Repo, error) {
	out, err := exec.Command("gh", "repo", "list",
		"--json", "name,nameWithOwner,description,stargazerCount,isPrivate,pushedAt",
		"--limit", "40",
	).Output()
	if err != nil {
		return nil, err
	}
	var repos []Repo
	return repos, json.Unmarshal(out, &repos)
}

// fetchIssues lists open and closed issues for a repository.
func fetchIssues(repo string) ([]Issue, error) {
	out, err := exec.Command("gh", "issue", "list",
		"--repo", repo,
		"--json", "number,title,state,author,createdAt,labels",
		"--state", "all",
		"--limit", "30",
	).Output()
	if err != nil {
		return nil, err
	}
	var issues []Issue
	return issues, json.Unmarshal(out, &issues)
}

// fetchPRs lists open and closed pull requests for a repository.
func fetchPRs(repo string) ([]PR, error) {
	out, err := exec.Command("gh", "pr", "list",
		"--repo", repo,
		"--json", "number,title,state,author,createdAt,isDraft",
		"--state", "all",
		"--limit", "30",
	).Output()
	if err != nil {
		return nil, err
	}
	var prs []PR
	return prs, json.Unmarshal(out, &prs)
}

// fetchCommits returns the 30 most recent commits for a repository.
func fetchCommits(repo string) ([]Commit, error) {
	out, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/commits?per_page=30", repo),
	).Output()
	if err != nil {
		return nil, err
	}
	var items []ghCommitItem
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, err
	}
	commits := make([]Commit, len(items))
	for i, item := range items {
		sha := item.SHA
		if len(sha) > 7 {
			sha = sha[:7]
		}
		msg := item.Commit.Message
		if idx := strings.IndexByte(msg, '\n'); idx >= 0 {
			msg = msg[:idx]
		}
		commits[i] = Commit{
			SHA:     sha,
			Message: msg,
			Author:  item.Commit.Author.Name,
			Date:    item.Commit.Author.Date,
		}
	}
	return commits, nil
}
