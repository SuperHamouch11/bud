package controller

import (
	"context"

	"gitlab.com/mnm/bud/example/hn/internal/hn"
)

type Controller struct {
	HN *hn.Client
}

func (c *Controller) Index(ctx context.Context) (news *hn.News, err error) {
	return c.HN.FrontPage(ctx)
}

// Show a comment
func (c *Controller) Show(ctx context.Context, id string) (story *hn.Story, err error) {
	return c.HN.Find(ctx, id)
}