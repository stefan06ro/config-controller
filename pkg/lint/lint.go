package lint

import (
	"fmt"
	"reflect"
)

const (
	overshadowErrorThreshold float64 = 0.75
)

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
