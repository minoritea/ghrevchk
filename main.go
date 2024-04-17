package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hasura/go-graphql-client"
)

func ownerAndRepo(r string) (string, string) {
	ss := strings.Split(string(r), "/")
	if len(ss) == 2 {
		return ss[0], ss[1]
	} else if len(ss) == 3 && ss[0] == "github.com" {
		return ss[1], ss[2]
	}
	return "", ""
}

func fetch(ctx context.Context, owner, repo, user, state string) error {
	client := graphql.NewClient("https://api.github.com/graphql", nil).
		WithRequestModifier(func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))
		})
	if os.Getenv("DEBUG") != "" {
		client = client.WithDebug(true)
	}
	type Query struct {
		Search struct {
			Nodes []struct {
				PullRequest struct {
					Title string
					URL   string
				} `graphql:"... on PullRequest"`
			}
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
		} `graphql:"search(query: $query, type: ISSUE, first: 100)"`
	}
	result, err := query[Query](ctx, client, map[string]any{
		"query": fmt.Sprintf(
			"is:pr repo:%s/%s state:%s involves:%s",
			owner,
			repo,
			state,
			user,
		),
	})
	if err != nil {
		return err
	}
	for _, node := range result.Search.Nodes {
		fmt.Printf("%s\n> %s\n\n", node.PullRequest.Title, node.PullRequest.URL)
	}
	return nil
}

func query[T any](ctx context.Context, client *graphql.Client, vars map[string]any, options ...graphql.Option) (*T, error) {
	var q T
	return &q, client.Query(ctx, &q, vars, options...)
}

func run() error {
	var (
		user  string
		state string
	)
	flag.StringVar(&user, "user", "", "user to filter by")
	flag.StringVar(&state, "state", "open", "state of PRs to list")
	flag.Parse()

	if len(flag.Args()) < 1 {
		return fmt.Errorf("usage: github-prs <repo>")
	}
	if user == "" {
		return fmt.Errorf("user is required")
	}
	owner, repo := ownerAndRepo(flag.Arg(0))
	return fetch(context.Background(), owner, repo, user, state)
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
