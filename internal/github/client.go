package github

import (
    "context"
    "fmt"

    "github.com/google/go-github/v57/github"
    "golang.org/x/oauth2"
)

type Client struct {
    client *github.Client
    owner  string
    repo   string
}

func NewClient(token, owner, repo string) *Client {
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(context.Background(), ts)
    
    return &Client{
        client: github.NewClient(tc),
        owner:  owner,
        repo:   repo,
    }
}

func (c *Client) PostComment(ctx context.Context, prNumber int, body string) error {
    comment := &github.IssueComment{
        Body: &body,
    }

    _, _, err := c.client.Issues.CreateComment(ctx, c.owner, c.repo, prNumber, comment)
    if err != nil {
        return fmt.Errorf("failed to post comment: %w", err)
    }

    return nil
}