package github

import (
	"context"
	"fmt"
	"strings"

	"get.cutie.cafe/isabelle/config"
	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
)

var (
	ghctx context.Context
	gh    *github.Client
)

// Init initializes the GitHub integration
func Init() {
	ghctx = context.Background()

	gh = github.NewClient(oauth2.NewClient(ghctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Config.GitHub.PAT},
	)))
}

// CreateIssue creates a GitHub issue with issueTitle, issueBody
func CreateIssue(issueTitle string, issueBody string) (string, error) {
	newIssue := &github.IssueRequest{
		Title:  &issueTitle,
		Body:   &issueBody,
		Labels: &config.Config.GitHub.Labels}

	issue, _, err := gh.Issues.Create(ghctx, strings.Split(config.Config.GitHub.Repo, "/")[0], strings.Split(config.Config.GitHub.Repo, "/")[1], newIssue)
	if err != nil {
		return "", err
	}
	if issue == nil {
		return "", fmt.Errorf("issue was null")
	}

	return fmt.Sprintf("https://github.com/%s/%d", config.Config.GitHub.Repo, issue.Number), nil
}
