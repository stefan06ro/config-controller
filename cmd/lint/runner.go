package lint

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/config-controller/pkg/generator"
	"github.com/giantswarm/config-controller/pkg/github"
	"github.com/giantswarm/config-controller/pkg/lint"
)

const (
	owner = "giantswarm"
	repo  = "config"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if IsLinterFoundIssues(err) {
		os.Exit(1)
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var store generator.Filesystem
	{
		gh, err := github.New(github.Config{
			Token: r.flag.GitHubToken,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		if r.flag.ConfigVersion != "" {
			tag, err := gh.GetLatestTag(ctx, owner, repo, r.flag.ConfigVersion)
			if err != nil {
				return microerror.Mask(err)
			}

			store, err = gh.GetFilesByTag(ctx, owner, repo, tag)
			if err != nil {
				return microerror.Mask(err)
			}

		} else if r.flag.Branch != "" {
			store, err = gh.GetFilesByBranch(ctx, owner, repo, r.flag.Branch)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	discovery, err := lint.NewDiscovery(store)
	if err != nil {
		return microerror.Mask(err)
	}

	messageCount := 0
	linterFuncs := lint.GetFilteredLinterFunctions(r.flag.FilterFunctions)
	fmt.Printf("Linting using %d functions\n\n", len(linterFuncs))

	for _, f := range linterFuncs {
		messages := f(discovery)
		for _, msg := range messages {
			if r.flag.OnlyErrors && !msg.IsError() {
				continue
			}

			fmt.Println(msg.Message(!r.flag.NoFuncNames, !r.flag.NoDescriptions))
			messageCount += 1

			if r.flag.MaxMessages > 0 && messageCount >= r.flag.MaxMessages {
				fmt.Println("-------------------------")
				fmt.Println("Too many messages, skipping the rest of checks")
				fmt.Printf("Run linter with '--%s 0' to see all the errors\n", flagMaxMessages)
				return nil
			}
		}
	}
	fmt.Printf("-------------------------\nFound %d issues\n", messageCount)
	if messageCount > 0 {
		return microerror.Mask(linterFoundIssuesError)
	}
	return nil
}
