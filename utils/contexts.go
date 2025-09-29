package utils

import (
	"context"

	"github.com/pubgo/dix/v2"
)

type dixCtx string

var Ctx = dixCtx("dix")

func GetDixCtx(ctx context.Context) *dix.Dix {
	return ctx.Value(Ctx).(*dix.Dix)
}

func CreateDixCtx(ctx context.Context, dix *dix.Dix) context.Context {
	return context.WithValue(ctx, Ctx, dix)
}
