package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	iterator "github.com/jcchavezs/gh-iterator"
	"github.com/jcchavezs/gh-iterator/exec"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
)

var flags struct {
	perPage       int
	page          string
	cloningSubset []string
	searchFilter  string
	command       string
	logLevel      slog.Level
}

func renderCommand(s string, repository string) string {
	return strings.ReplaceAll(s, "{{ .Repository }}", repository)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gh-iterator-run",
		Short: "Filter GitHub repositories using CEL expressions",
		Long: `A CLI tool that iterates over GitHub organization repositories 
and filters them using CEL (Common Expression Language) conditions.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			logHandler := slog.NewJSONHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{Level: flags.logLevel})
			logger := slog.New(logHandler)

			searchFilterIn, err := parseSearchFilterIn(flags.searchFilter, logger)
			if err != nil {
				return err
			}

			var p int
			if flags.page == "all" {
				p = -1
			} else if flags.page != "" {
				if p, err = strconv.Atoi(flags.page); err != nil {
					return err
				}
			}

			res, err := iterator.RunForOrganization(
				ctx, args[0],
				iterator.SearchOptions{
					FilterIn: searchFilterIn,
					PerPage:  flags.perPage,
					Page:     iterator.PageN(p),
				},
				func(ctx context.Context, repository string, isEmpty bool, exec exec.Execer) error {
					if flags.command != "" {
						res, err := exec.Run(ctx, os.Getenv("SHELL"), "-c", renderCommand(flags.command, repository))
						if err != nil {
							io.WriteString(cmd.ErrOrStderr(), res.Stderr)
							return err
						}

						io.WriteString(cmd.OutOrStdout(), res.Stdout)
					}
					return nil
				},
				iterator.Options{
					LogHandler:    logHandler,
					CloningSubset: flags.cloningSubset,
				},
			)

			if err != nil {
				return err
			}

			fmt.Printf("Processed %d repositories\n", res.Processed)
			fmt.Printf("Filtered %d repositories\n", res.Inspected)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&flags.searchFilter, "search-filter", "s", "", "CEL condition(s) to search repositories. By default, it filters out archived, forked, and empty repositories.")
	rootCmd.Flags().StringVarP(&flags.command, "command", "c", "", "CEL condition(s) to search repositories.")
	rootCmd.Flags().StringVar(&flags.page, "page", "all", "Page number to fetch, or 'all' to fetch all pages")
	rootCmd.Flags().IntVar(&flags.perPage, "per-page", 100, "Number of repositories to fetch per page")
	rootCmd.Flags().StringArrayVar(&flags.cloningSubset, "cloning-subset", nil, "")
	rootCmd.PersistentFlags().Var(
		enumflag.New(&flags.logLevel, "string", LevelIds, enumflag.EnumCaseInsensitive),
		"log-level",
		"Sets the log level",
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
