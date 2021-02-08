package lint

import (
	"reflect"
	"regexp"
	"runtime"
)

const (
	overshadowErrorThreshold float64 = 0.75
)

type LinterFunc func(d *Discovery) (messages LinterMessages)

var AllLinterFunctions = []LinterFunc{
	LintUnusedConfigValues,
	LintDuplicateConfigValues,
	LintOvershadowedConfigValues,
	LintUnusedConfigPatchValues,
	LintUndefinedTemplateValues,
	LintUndefinedTemplatePatchValues,
	LintUnusedSecretValues,
	LintUndefinedSecretTemplateValues,
	LintUndefinedSecretTemplatePatchValues,
	LintIncludeFiles,
}

func LintDuplicateConfigValues(d *Discovery) (messages LinterMessages) {
	for path, defaultPath := range d.Config.paths {
		for _, overshadowingPatch := range defaultPath.OvershadowedBy {
			patchedPath := overshadowingPatch.paths[path]
			if reflect.DeepEqual(defaultPath.Value, patchedPath.Value) {
				messages = append(
					messages,
					NewError(overshadowingPatch.filepath, path, "is duplicate of the same path in %s", d.Config.filepath),
				)
			}
		}
	}
	return messages
}

func LintOvershadowedConfigValues(d *Discovery) (messages LinterMessages) {
	if len(d.Installations) == 0 {
		return // avoid division by 0
	}
	for path, valuePath := range d.Config.paths {
		if len(valuePath.OvershadowedBy) == len(d.Installations) {
			messages = append(
				messages,
				NewError(d.Config.filepath, path, "is overshadowed by all config.yaml.patch files"),
			)
		} else if float64(len(valuePath.OvershadowedBy)/len(d.Installations)) >= overshadowErrorThreshold {
			msg := NewMessage(
				d.Config.filepath, path, "is overshadowed by %d/%d patches",
				len(valuePath.OvershadowedBy), len(d.Installations),
			).WithDescription("consider removing it from %s", d.Config.filepath)
			messages = append(messages, msg)
		}
	}
	return messages
}

func LintUnusedConfigPatchValues(d *Discovery) (messages LinterMessages) {
	for _, configPatch := range d.ConfigPatches {
		if len(d.AppsPerInstallation[configPatch.installation]) == 0 {
			continue // avoid division by 0
		}
		for path, valuePath := range configPatch.paths {
			if len(valuePath.UsedBy) > 0 {
				continue
			}
			messages = append(messages, NewError(configPatch.filepath, path, "is unused"))
		}
	}
	return messages
}

func LintUnusedConfigValues(d *Discovery) (messages LinterMessages) {
	if len(d.Installations) == 0 || len(d.Apps) == 0 {
		return // what's the point, nothing is defined
	}
	for path, valuePath := range d.Config.paths {
		if len(valuePath.UsedBy) == 0 {
			messages = append(messages, NewError(d.Config.filepath, path, "is unused"))
		} else if len(valuePath.UsedBy) == 1 {
			msg := NewMessage(d.Config.filepath, path, "is used by just one app: %s", valuePath.UsedBy[0].app).
				WithDescription("consider moving this value to %s template or template patch", valuePath.UsedBy[0].app)
			messages = append(messages, msg)
		}
	}
	return messages
}

func LintUnusedSecretValues(d *Discovery) (messages LinterMessages) {
	if len(d.Installations) == 0 {
		return // what's the point, nothing is defined
	}
	for _, secretFile := range d.Secrets {
		for path, valuePath := range secretFile.paths {
			if len(valuePath.UsedBy) == 0 {
				messages = append(messages, NewError(secretFile.filepath, path, "is unused"))
			} else if len(valuePath.UsedBy) == 1 {
				msg := NewMessage(secretFile.filepath, path, "is used by just one app: %s", valuePath.UsedBy[0].app).
					WithDescription("consider moving this value to %s secret-values patch", valuePath.UsedBy[0].app)
				messages = append(messages, msg)
			}

		}
	}
	return messages
}

func LintUndefinedSecretTemplateValues(d *Discovery) (messages LinterMessages) {
	for _, template := range d.SecretTemplates {
		for path, value := range template.values {
			if !value.MayBeMissing {
				continue
			}

			messages = append(messages, NewError(template.filepath, path, "is templated but never configured"))
		}
	}
	return messages
}

func LintUndefinedSecretTemplatePatchValues(d *Discovery) (messages LinterMessages) {
	for _, template := range d.SecretTemplatePatches {
		for path, value := range template.values {
			if !value.MayBeMissing {
				continue
			}

			messages = append(messages, NewError(template.filepath, path, "is templated but never configured"))
		}
	}
	return messages
}

func LintUndefinedTemplateValues(d *Discovery) (messages LinterMessages) {
	for _, template := range d.Templates {
		for path, value := range template.values {
			if !value.MayBeMissing {
				continue
			}

			used := false
			for _, templatePatch := range d.TemplatePatches {
				if _, ok := templatePatch.paths[path]; ok {
					used = true
					break
				}
			}

			if used {
				continue
			}
			messages = append(messages, NewError(template.filepath, path, "is templated but never configured"))
		}
	}
	return messages
}

func LintUndefinedTemplatePatchValues(d *Discovery) (messages LinterMessages) {
	for _, templatePatch := range d.TemplatePatches {
		for path, value := range templatePatch.values {
			if !value.MayBeMissing {
				continue
			}
			messages = append(messages, NewError(templatePatch.filepath, path, "is templated but never configured"))
		}
	}
	return messages
}

func LintIncludeFiles(d *Discovery) (messages LinterMessages) {
	used := map[string]bool{}
	exist := map[string]bool{}
	for _, includeFile := range d.Include {
		exist[includeFile.filepath] = true
		for _, filepath := range includeFile.includes {
			used[filepath] = true
		}
	}

	for _, template := range d.Templates {
		for _, filepath := range template.includes {
			used[filepath] = true
		}
	}

	for _, templatePatch := range d.TemplatePatches {
		for _, filepath := range templatePatch.includes {
			used[filepath] = true
		}
	}

	if reflect.DeepEqual(exist, used) {
		return messages
	}

	for filepath := range exist {
		if _, ok := used[filepath]; !ok {
			messages = append(messages, NewError(filepath, "*", "is never included"))
		}
	}

	for filepath := range used {
		if _, ok := exist[filepath]; !ok {
			messages = append(messages, NewError(filepath, "*", "is included but does not exist"))
		}
	}

	return messages
}

//------ helper funcs -------
func GetFilteredLinterFunctions(filters []string) []LinterFunc {
	if len(filters) == 0 {
		return AllLinterFunctions
	}

	functions := []LinterFunc{}
	for _, function := range AllLinterFunctions {
		name := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
		for _, filter := range filters {
			re := regexp.MustCompile(filter)
			if re.MatchString(name) {
				functions = append(functions, function)
				break
			}
		}
	}

	return functions
}
