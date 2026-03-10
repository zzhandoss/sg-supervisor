package app

import (
	"context"
)

func (a *App) ServeAdminUI(ctx context.Context, listen string) error {
	return a.StartService(ctx, "admin-ui")
}
