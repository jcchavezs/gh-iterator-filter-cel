package main

import (
	"log/slog"

	"github.com/google/cel-go/cel"
	iterator "github.com/jcchavezs/gh-iterator"
)

var defaultSearchFilterIn = func(r iterator.Repository) bool {
	return !r.Archived && !r.Fork && r.Size > 0
}

func parseSearchFilterIn(cond string, l *slog.Logger) (func(iterator.Repository) bool, error) {
	if cond == "" {
		return defaultSearchFilterIn, nil
	}

	env, err := cel.NewEnv(
		cel.Variable("repo", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(cond)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	return func(r iterator.Repository) bool {
		repoMap := map[string]any{
			"name":       r.Name,
			"archived":   r.Archived,
			"language":   r.Language,
			"visibility": r.Visibility,
			"fork":       r.Fork,
			"isEmpty":    r.Size == 0,
			"pushedAt":   r.PushedAt,
		}

		out, _, err := prg.Eval(map[string]any{"repo": repoMap})
		if err != nil {
			l.Error("Failed to evaluate CEL expression", "error", err)
			return false
		}

		result, ok := out.Value().(bool)
		return ok && result
	}, nil
}
