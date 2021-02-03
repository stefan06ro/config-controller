package lint

import (
	"log"
	"reflect"
)

type LinterFunc func(*Discovery) (errors []string)

func GlobalDuplicateConfigValues(d *Discovery) (errors []string) {
	for path, valuePath := range d.Config.paths {
		for _, overshadowingPatch := range valuePath.OvershadowedBy {
			patchedPath, _ := overshadowingPatch.paths[path]
			if reflect.DeepEqual(valuePath.Value, patchedPath.Value) {
				log.Printf(
					"path %q in %q is duplicates same path in config.yaml",
					path, overshadowingPatch.filepath,
				)
			}
		}
	}
	return []string{}
}
