package main

import (
	"log/slog"
	"os"
	"testing"
	"time"

	iterator "github.com/jcchavezs/gh-iterator"
	"github.com/stretchr/testify/require"
)

func TestParseSearchFilter_BasicConditions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("filter by language - match", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.language == "Go"`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:     "test-repo",
			Language: "Go",
		}
		require.True(t, filterFn(repo))
	})

	t.Run("filter by language - no match", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.language == "Go"`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:     "test-repo",
			Language: "Python",
		}
		require.False(t, filterFn(repo))
	})

	t.Run("filter by archived status - not archived", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`!repo.archived`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:     "test-repo",
			Archived: false,
		}
		require.True(t, filterFn(repo))
	})

	t.Run("filter by archived status - archived", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.archived`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:     "archived-repo",
			Archived: true,
		}
		require.True(t, filterFn(repo))
	})

	t.Run("filter by fork status - not fork", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`!repo.fork`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name: "original-repo",
			Fork: false,
		}
		require.True(t, filterFn(repo))
	})

	t.Run("filter by fork status - is fork", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.fork`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name: "forked-repo",
			Fork: true,
		}
		require.True(t, filterFn(repo))
	})
}

func TestParseSearchFilter_NumericConditions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("filter by empty", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.isEmpty`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name: "large-repo",
			Size: 200,
		}
		require.False(t, filterFn(repo))

		smallRepo := iterator.Repository{
			Name: "empty-repo",
			Size: 0,
		}
		require.True(t, filterFn(smallRepo))
	})
}

func TestParseSearchFilter_StringOperations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("filter by visibility", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.visibility == "public"`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:       "public-repo",
			Visibility: "public",
		}
		require.True(t, filterFn(repo))

		privateRepo := iterator.Repository{
			Name:       "private-repo",
			Visibility: "private",
		}
		require.False(t, filterFn(privateRepo))
	})

	t.Run("filter by name contains", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.name.contains("test")`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name: "my-test-repo",
		}
		require.True(t, filterFn(repo))

		noMatchRepo := iterator.Repository{
			Name: "my-repo",
		}
		require.False(t, filterFn(noMatchRepo))
	})

	t.Run("filter by name startsWith", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.name.startsWith("gh-")`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name: "gh-iterator",
		}
		require.True(t, filterFn(repo))

		noMatchRepo := iterator.Repository{
			Name: "iterator-gh",
		}
		require.False(t, filterFn(noMatchRepo))
	})

	t.Run("filter by name endsWith", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.name.endsWith("-cli")`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name: "my-app-cli",
		}
		require.True(t, filterFn(repo))

		noMatchRepo := iterator.Repository{
			Name: "cli-my-app",
		}
		require.False(t, filterFn(noMatchRepo))
	})

	t.Run("filter by empty string language", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.language == ""`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:     "no-language",
			Language: "",
		}
		require.True(t, filterFn(repo))
	})
}

func TestParseSearchFilter_ComplexConditions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("complex condition - AND - all match", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.language == "Go" && !repo.fork && !repo.archived`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:     "go-repo",
			Language: "Go",
			Fork:     false,
			Archived: false,
		}
		require.True(t, filterFn(repo))
	})

	t.Run("complex condition - AND - partial match", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.language == "Go" && !repo.fork && !repo.archived`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:     "go-fork",
			Language: "Go",
			Fork:     true,
			Archived: false,
		}
		require.False(t, filterFn(repo))
	})

	t.Run("complex condition - OR", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.language == "Go" || repo.language == "Python"`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		goRepo := iterator.Repository{
			Name:     "go-repo",
			Language: "Go",
		}
		require.True(t, filterFn(goRepo))

		pythonRepo := iterator.Repository{
			Name:     "python-repo",
			Language: "Python",
		}
		require.True(t, filterFn(pythonRepo))

		javaRepo := iterator.Repository{
			Name:     "java-repo",
			Language: "Java",
		}
		require.False(t, filterFn(javaRepo))
	})

	t.Run("complex condition - multiple operators", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`(repo.language == "Go" || repo.language == "Rust") && !repo.archived`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		matchingRepo := iterator.Repository{
			Name:     "go-large-repo",
			Language: "Go",
			Archived: false,
		}
		require.True(t, filterFn(matchingRepo))
	})
}

func TestParseSearchFilter_TimestampConditions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("filter by pushed at - recent", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`timestamp(repo.pushedAt) > timestamp("2024-01-01T00:00:00Z")`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name:     "recent-repo",
			PushedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		}
		require.True(t, filterFn(repo))

		oldRepo := iterator.Repository{
			Name:     "old-repo",
			PushedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		require.False(t, filterFn(oldRepo))
	})
}

func TestParseSearchFilter_ErrorCases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("empty condition", func(t *testing.T) {
		_, err := parseSearchFilterIn("", logger)
		require.Error(t, err)
	})

	t.Run("invalid CEL expression - incomplete", func(t *testing.T) {
		_, err := parseSearchFilterIn(`repo.Language ==`, logger)
		require.Error(t, err)
	})

	t.Run("invalid CEL expression - single equals", func(t *testing.T) {
		_, err := parseSearchFilterIn(`repo.Language = "Go"`, logger)
		require.Error(t, err)
	})

	t.Run("invalid CEL expression - missing dot", func(t *testing.T) {
		_, err := parseSearchFilterIn(`repository Language == "Go"`, logger)
		require.Error(t, err)
	})

	t.Run("invalid CEL expression - incomplete AND", func(t *testing.T) {
		_, err := parseSearchFilterIn(`repo.Language == "Go" &&`, logger)
		require.Error(t, err)
	})

	//t.Run("non-existent field", func(t *testing.T) {
	//	_, err := parseSearchFilter(`repo.NonExistentField == "value"`, logger)
	//	require.Error(t, err)
	//})
}

func TestParseSearchFilter_NonBooleanExpression(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("expression returns non-boolean", func(t *testing.T) {
		filterFn, err := parseSearchFilterIn(`repo.Size`, logger)
		require.NoError(t, err)
		require.NotNil(t, filterFn)

		repo := iterator.Repository{
			Name: "test-repo",
			Size: 100,
		}
		// Should return false because result is not a boolean
		require.False(t, filterFn(repo))
	})
}
