package cmd

import "context"

type runtimeKey struct{}

func withRuntimeOptions(ctx context.Context, opts runtimeOptions) context.Context {
	return context.WithValue(ctx, runtimeKey{}, opts)
}

func runtimeFromContext(ctx context.Context) runtimeOptions {
	if ctx == nil {
		return runtimeOptions{}
	}
	if v := ctx.Value(runtimeKey{}); v != nil {
		if opts, ok := v.(runtimeOptions); ok {
			return opts
		}
	}
	return runtimeOptions{}
}
