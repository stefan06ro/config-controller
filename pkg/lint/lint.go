package lint

import (
	"fmt"
	"reflect"
)

const (
	overshadowErrorThreshold  float64 = 0.75
	patchUsedByErrorThreshold float64 = 0.25
)

// TODO: kuba - how about having custom error type
type LinterFunc func(*Discovery) (errors []string)

func GlobalDuplicateConfigValues(d *Discovery) (errors []string) {
	for path, valuePath := range d.Config.paths {
		for _, overshadowingPatch := range valuePath.OvershadowedBy {
			patchedPath, _ := overshadowingPatch.paths[path]
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
		} else if float64(len(valuePath.UsedBy)/len(d.Apps)) <= patchUsedByErrorThreshold {
			errors = append(
				errors,
				fmt.Sprintf(
					"path %q in %q is used by %d/%d apps; consider moving it to app templates",
					path, d.Config.filepath, len(valuePath.UsedBy), len(d.Apps),
				),
			)
		}
	}
	return errors
}
