package lint

import (
	"fmt"
	"sort"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/generator"
)

type Discovery struct {
	Config        *ValueFile
	ConfigPatches []*ValueFile
	Secrets       []*ValueFile

	Templates             []*TemplateFile
	SecretTemplates       []*TemplateFile
	TemplatePatches       []*TemplateFile
	SecretTemplatePatches []*TemplateFile
	Include               []*TemplateFile

	Installations []string
	Apps          []string

	AppsPerInstallation                  map[string][]string
	ConfigPatchesPerInstallation         map[string]*ValueFile
	SecretsPerInstallation               map[string]*ValueFile
	TemplatesPerApp                      map[string]*TemplateFile
	SecretTemplatesPerApp                map[string]*TemplateFile
	TemplatePatchesPerInstallation       map[string][]*TemplateFile
	SecretTemplatePatchesPerInstallation map[string][]*TemplateFile
}

func (d Discovery) GetAppTemplatePatch(installation, app string) (*TemplateFile, bool) {
	templatePatches, ok := d.TemplatePatchesPerInstallation[installation]
	if !ok {
		return nil, false
	}
	for _, patch := range templatePatches {
		if patch.app == app {
			return patch, true
		}
	}
	return nil, false
}

func (d Discovery) GetAppSecretTemplatePatch(installation, app string) (*TemplateFile, bool) {
	templatePatches, ok := d.SecretTemplatePatchesPerInstallation[installation]
	if !ok {
		return nil, false
	}
	for _, patch := range templatePatches {
		if patch.app == app {
			return patch, true
		}
	}
	return nil, false
}

func NewDiscovery(fs generator.Filesystem) (*Discovery, error) {
	d := &Discovery{
		ConfigPatches: []*ValueFile{},
		Secrets:       []*ValueFile{},

		Templates:             []*TemplateFile{},
		SecretTemplates:       []*TemplateFile{},
		TemplatePatches:       []*TemplateFile{},
		SecretTemplatePatches: []*TemplateFile{},

		Installations: []string{},
		Apps:          []string{},

		AppsPerInstallation:                  map[string][]string{},
		ConfigPatchesPerInstallation:         map[string]*ValueFile{},
		SecretsPerInstallation:               map[string]*ValueFile{},
		TemplatesPerApp:                      map[string]*TemplateFile{},
		SecretTemplatesPerApp:                map[string]*TemplateFile{},
		TemplatePatchesPerInstallation:       map[string][]*TemplateFile{},
		SecretTemplatePatchesPerInstallation: map[string][]*TemplateFile{},
	}

	// collect config.yaml
	{
		filepath := "default/config.yaml"
		body, err := fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		d.Config, err = NewValueFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	uniqueInstallations := map[string]bool{}
	uniqueApps := map[string]bool{}

	installationDirs, err := fs.ReadDir("installations/")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// collect installations/*/config.yaml.patch & installations/*/secret.yaml files
	for _, inst := range installationDirs {
		if !inst.IsDir() {
			continue
		}
		uniqueInstallations[inst.Name()] = true
		filepath := fmt.Sprintf("installations/%s/config.yaml.patch", inst.Name())
		body, err := fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		patch, err := NewValueFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		d.ConfigPatches = append(d.ConfigPatches, patch)
		d.ConfigPatchesPerInstallation[inst.Name()] = patch

		filepath = fmt.Sprintf("installations/%s/secret.yaml", inst.Name())
		body, err = fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		secret, err := NewValueFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		d.Secrets = append(d.Secrets, secret)
		d.SecretsPerInstallation[inst.Name()] = secret
	}

	// collect default/apps/*/{configmap,secret}-values.yaml.template files
	defaultAppDirs, err := fs.ReadDir("default/apps/")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	for _, app := range defaultAppDirs {
		if !app.IsDir() {
			continue
		}
		uniqueApps[app.Name()] = true
		filepath := fmt.Sprintf("default/apps/%s/configmap-values.yaml.template", app.Name())
		body, err := fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		template, err := NewTemplateFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		d.Templates = append(d.Templates, template)
		d.TemplatesPerApp[app.Name()] = template

		filepath = fmt.Sprintf("default/apps/%s/secret-values.yaml.template", app.Name())
		body, err = fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		secret, err := NewTemplateFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		d.SecretTemplates = append(d.SecretTemplates, secret)
		d.SecretTemplatesPerApp[app.Name()] = secret
	}

	// collect installations/*/apps/*/{configmap,secret}-values.yaml.patch files
	for _, inst := range installationDirs {
		if !inst.IsDir() {
			continue
		}
		d.AppsPerInstallation[inst.Name()] = []string{}
		d.TemplatePatchesPerInstallation[inst.Name()] = []*TemplateFile{}
		d.SecretTemplatePatchesPerInstallation[inst.Name()] = []*TemplateFile{}
		appDirs, err := fs.ReadDir("default/apps/")
		if err != nil {
			return nil, microerror.Mask(err)
		}
		for _, app := range appDirs {
			if !app.IsDir() {
				continue
			}
			uniqueApps[app.Name()] = true
			d.AppsPerInstallation[inst.Name()] = append(d.AppsPerInstallation[inst.Name()], app.Name())

			filepath := fmt.Sprintf("installations/%s/apps/%s/configmap-values.yaml.patch", inst.Name(), app.Name())
			body, err := fs.ReadFile(filepath)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			templatePatch, err := NewTemplateFile(filepath, body)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			d.TemplatePatches = append(d.TemplatePatches, templatePatch)
			d.TemplatePatchesPerInstallation[inst.Name()] = append(
				d.TemplatePatchesPerInstallation[inst.Name()],
				templatePatch,
			)

			filepath = fmt.Sprintf("installations/%s/apps/%s/secret-values.yaml.patch", inst.Name(), app.Name())
			body, err = fs.ReadFile(filepath)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			secretPatch, err := NewTemplateFile(filepath, body)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			d.SecretTemplatePatches = append(d.SecretTemplatePatches, secretPatch)
			d.SecretTemplatePatchesPerInstallation[inst.Name()] = append(
				d.SecretTemplatePatchesPerInstallation[inst.Name()],
				secretPatch,
			)
		}
	}

	// collect include files
	includeFiles, err := fs.ReadDir("include/")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	for _, includeFile := range includeFiles {
		if includeFile.IsDir() {
			continue
		}
		filepath := fmt.Sprintf("include/%s", includeFile.Name())
		body, err := fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		includeTemplate, err := NewTemplateFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		d.Include = append(d.Include, includeTemplate)
	}

	for k := range uniqueInstallations {
		d.Installations = append(d.Installations, k)
	}
	for k := range uniqueApps {
		d.Apps = append(d.Apps, k)
	}
	sort.Strings(d.Installations)
	sort.Strings(d.Apps)

	if err := d.populateValuePaths(); err != nil {
		return nil, microerror.Mask(err)
	}

	return d, nil
}

// populateValuePaths fills UsedBy and OvershadowedBy fields in all ValuePath
// structs in d.Config and d.ConfigPatches. This allows linter to find unused
// values easier.
func (d *Discovery) populateValuePaths() error {
	// 1. Mark all overshadowed valuePaths in config.yaml
	for _, configPatch := range d.ConfigPatches {
		for path := range configPatch.paths {
			if original, ok := d.Config.paths[path]; ok {
				original.OvershadowedBy = append(original.OvershadowedBy, configPatch)
			}
		}
	}
	// 2. Check templates for all apps x installations, then set UsedBy fields
	// in Config, ConfigPatches
	for _, installation := range d.Installations {
		configPatch, ok := d.ConfigPatchesPerInstallation[installation]
		if !ok {
			configPatch = nil
		}

		for _, app := range d.Apps {
			// mark all fields used by app template's patch
			if templatePatch, ok := d.GetAppTemplatePatch(installation, app); ok {
				populatePathsWithUsedBy(templatePatch, d.Config, configPatch)
			}

			// mark all fields used by the app's default template
			if defaultTemplate, ok := d.TemplatesPerApp[app]; ok {
				populatePathsWithUsedBy(defaultTemplate, d.Config, configPatch)
			}
		}
	}

	// 3. Check SECRET templates for all apps x installations, then set UsedBy fields
	for _, installation := range d.Installations {
		secret, ok := d.SecretsPerInstallation[installation]
		if !ok {
			continue
		}
		for _, app := range d.Apps {
			templatePatch, ok := d.GetAppSecretTemplatePatch(installation, app)
			if ok {
				populateSecretPathsWithUsedBy(templatePatch, secret, nil)
			} else {
				templatePatch = nil
			}

			if defaultTemplate, ok := d.SecretTemplatesPerApp[app]; ok {
				populateSecretPathsWithUsedBy(defaultTemplate, secret, templatePatch)
			}
		}
	}

	return nil
}

func populatePathsWithUsedBy(source *TemplateFile, config, configPatch *ValueFile) {
	for path, templatePath := range source.values {
		if configPatch != nil {
			valuePath, valuePathOk := configPatch.paths[path]
			if valuePathOk {
				// config patch exists and contains the path
				valuePath.UsedBy = appendUniqueUsedBy(valuePath.UsedBy, source)
				continue
			}
		}

		valuePath, valuePathOk := config.paths[path]
		if valuePathOk {
			// the value comes from default config
			valuePath.UsedBy = appendUniqueUsedBy(valuePath.UsedBy, source)
			continue
		}

		// value is missing from config; linter will check if it's patched
		templatePath.MayBeMissing = true
	}
}

func populateSecretPathsWithUsedBy(source *TemplateFile, installationSecret *ValueFile, secretTemplatePatch *TemplateFile) {
	for path, templatePath := range source.values {
		if secretTemplatePatch != nil {
			if _, ok := secretTemplatePatch.paths[path]; ok {
				// path was already checked in the patch
				continue
			}
		}

		valuePath, valuePathOk := installationSecret.paths[path]
		if valuePathOk {
			// config patch exists and contains the path
			valuePath.UsedBy = appendUniqueUsedBy(valuePath.UsedBy, source)
			continue
		}

		// value is missing from secrets
		templatePath.MayBeMissing = true
	}
}

func appendUniqueUsedBy(list []*TemplateFile, t *TemplateFile) []*TemplateFile {
	for _, v := range list {
		if v == t {
			return list
		}
	}
	return append(list, t)
}
