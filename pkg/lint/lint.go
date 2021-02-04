package lint

import (
	"bytes"
	"fmt"
	"reflect"

	pathmodifier "github.com/giantswarm/valuemodifier/path"
)

const (
	overshadowErrorThreshold  float64 = 0.75
	patchUsedByErrorThreshold float64 = 0.25
)

// TODO: kuba - how about having custom error type
type LinterFunc func(d *Discovery) (errors []string)

func GlobalDuplicateConfigValues(d *Discovery) (errors []string) {
	for path, valuePath := range d.Config.paths {
		for _, overshadowingPatch := range valuePath.OvershadowedBy {
			patchedPath := overshadowingPatch.paths[path]
			if reflect.DeepEqual(valuePath.Value, patchedPath.Value) {
				errors = append(
					errors,
					fmt.Sprintf(
						"path %q in %q is a duplicate of the same path in config.yaml",
						path, overshadowingPatch.filepath,
					),
				)
			}
		}
	}
	return errors
}

func GlobalOvershadowedValues(d *Discovery) (errors []string) {
	if len(d.Installations) == 0 {
		return // avoid division by 0
	}
	for path, valuePath := range d.Config.paths {
		if float64(len(valuePath.OvershadowedBy)/len(d.Installations)) >= overshadowErrorThreshold {
			errors = append(
				errors,
				fmt.Sprintf(
					"path %q in config.yaml is overshadowed by %d/%d patches; consider removing it from config.yaml",
					path, len(valuePath.OvershadowedBy), len(d.Installations),
				),
			)
		}
	}
	return errors
}

func PatchUnusedValues(d *Discovery) (errors []string) {
	for _, configPatch := range d.ConfigPatches {
		if len(d.AppsPerInstallation[configPatch.installation]) == 0 {
			continue // avoid division by 0
		}
		for path, valuePath := range configPatch.paths {
			if len(valuePath.UsedBy) == 0 {
				errors = append(
					errors,
					fmt.Sprintf(
						"path %q in %q is *unused*; consider removing it",
						path, configPatch.filepath,
					),
				)
			} else if float64(len(valuePath.UsedBy)/len(d.AppsPerInstallation[configPatch.installation])) <= patchUsedByErrorThreshold {
				errors = append(
					errors,
					fmt.Sprintf(
						"path %q in %q is used by %d/%d apps; consider moving it to app templates",
						path, configPatch.filepath, len(valuePath.UsedBy), len(d.AppsPerInstallation[configPatch.installation]),
					),
				)
			}
		}
	}
	return errors
}

func GlobalConfigUnusedValues(d *Discovery) (errors []string) {
	if len(d.Installations) == 0 || len(d.Apps) == 0 {
		return // what's the point, nothing is defined
	}
	for path, valuePath := range d.Config.paths {
		if len(valuePath.UsedBy) == 0 {
			errors = append(
				errors,
				fmt.Sprintf(
					"path %q in %q is *unused*; consider removing it",
					path, d.Config.filepath,
				),
			)
		} else if len(valuePath.UsedBy) == 1 {
			errors = append(
				errors,
				fmt.Sprintf(
					"path %q in %q is used by just one app in %q; consider moving it",
					path, d.Config.filepath, valuePath.UsedBy[0].filepath,
				),
			)
		}
	}
	return errors
}

// TODO: linter func might need to return internal errors instead of panicking!
func UnusedPatchableAppValues(d *Discovery) (errors []string) {
	for _, template := range d.Templates {
		for path, value := range template.values {
			if !value.MayBeMissing {
				continue
			}

			used := false
			for _, templatePatch := range d.TemplatePatches {
				// render template with zero data
				t := templatePatch.CopyTemplate().Option("missingkey=zero")
				output := bytes.NewBuffer([]byte{})
				var data interface{}
				err := t.Execute(output, data)
				if err != nil {
					panic(err)
				}

				c := pathmodifier.Config{
					InputBytes: output.Bytes(),
					Separator:  ".",
				}

				svc, err := pathmodifier.New(c)
				if err != nil {
					panic(err)
				}

				_, err = svc.Get(path)
				if err != nil && pathmodifier.IsNotFound(err) {
					continue
				} else if err != nil {
					panic(err)
				}

				used = true
				break
			}

			if used {
				continue
			}
			errors = append(
				errors,
				fmt.Sprintf(
					"path %q in %q is never configured; consider removing it",
					path, template.filepath,
				),
			)
		}
	}
	return errors
}

func UnconfiguredAppValues(d *Discovery) (errors []string) {
	for _, templatePatch := range d.TemplatePatches {
		for path, value := range templatePatch.values {
			if !value.MayBeMissing {
				continue
			}
			errors = append(
				errors,
				fmt.Sprintf(
					"path %q in %q is never configured; consider removing it",
					path, templatePatch.filepath,
				),
			)
		}
	}
	return errors
}
