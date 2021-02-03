package lint

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime"

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

	separator = "-------------------------"
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
	if err != nil {
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

	linterFuncs := []lint.LinterFunc{
		lint.GlobalDuplicateConfigValues,
		lint.GlobalOvershadowedValues,
		lint.PatchUnusedValues,
		lint.GlobalConfigUnusedValues,
		lint.UnusedPatchableAppValues,
	}

	errorsFound := 0
	for _, f := range linterFuncs {
		errors := f(discovery)
		if len(errors) > 0 {
			fmt.Println(separator + " " + runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name() + " " + separator)
		}
		for _, e := range errors {
			fmt.Println(e)
			errorsFound += 1
			if r.flag.MaxErrors > 0 && errorsFound >= r.flag.MaxErrors {
				fmt.Println(separator)
				fmt.Println("Too many errors, skipping the rest of checks")
				fmt.Printf("Run linter with '--%s 0' to see all the errors\n", flagMaxErrors)
				return nil
			}
		}
	}
	fmt.Printf("%s\nFound %d errors\n", separator, errorsFound)
	return nil
}
