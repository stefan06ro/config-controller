package lint

import (
	"fmt"
	"reflect"
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
