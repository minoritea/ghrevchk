package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v61/github"
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

	var (
		prs         []*github.PullRequest
		page        int
		ctx         = context.Background()
		client      = github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
		owner, repo = ownerAndRepo(flag.Arg(0))
	)
	for {
		list, res, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
			State: state,
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		if err != nil {
			return err
		}
		prs = append(prs, list...)
		if res.NextPage == 0 {
			break
		}
		page = res.NextPage
	}
	var requested []*github.PullRequest
	for _, pr := range prs {
		for _, reviewer := range pr.RequestedReviewers {
			if reviewer.GetLogin() == user {
				requested = append(requested, pr)
			}
		}
	}
	for _, pr := range requested {
		fmt.Printf("%s\n>  %s\n\n", pr.GetTitle(), pr.GetHTMLURL())
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
