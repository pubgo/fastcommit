package githubclient

import (
	"context"
	"net/http"

	"github.com/google/go-github/v71/github"
	"github.com/samber/lo"
)

func NewPublicRelease(owner, repo string) *PublicRelease {
	return &PublicRelease{
		client: github.NewClient(http.DefaultClient),
		owner:  owner,
		repo:   repo,
	}
}

type PublicRelease struct {
	client      *github.Client
	owner, repo string
}

func (g PublicRelease) List(ctx context.Context, pageSize ...int) ([]*github.RepositoryRelease, error) {
	size := lo.FirstOr(pageSize, 100)
	releases, _, err := g.client.Repositories.ListReleases(ctx, g.owner, g.repo, &github.ListOptions{PerPage: size})
	return releases, err
}

func (g PublicRelease) Latest(ctx context.Context) (*github.RepositoryRelease, error) {
	rsp, _, err := g.client.Repositories.GetLatestRelease(ctx, g.owner, g.repo)
	return rsp, err
}
