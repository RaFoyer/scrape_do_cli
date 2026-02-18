package cmd

import (
	"context"
)

type InfoCmd struct{}

func (c *InfoCmd) Run(ctx context.Context) error {
	if err := requireToken(ctx); err != nil {
		return err
	}
	resp, err := newClientFromContext(ctx).Info(ctx)
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, nil)
}
