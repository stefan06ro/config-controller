package lint

import (
	"fmt"
	"sort"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/generator"
)

type Discovery struct {
	Config          *ValueFile
	ConfigPatches   []*ValueFile
	Templates       []*TemplateFile
	TemplatePatches []*TemplateFile

	Installations       []string
	Apps                []string
	AppsPerInstallation map[string][]string
}

func (d Discovery) GetConfigPatch(installation string) (*ValueFile, bool) {
	for _, patch := range d.ConfigPatches {
		if patch.installation == installation {
			return patch, true
		}
	}
	return nil, false
}

func (d Discovery) GetAppTemplate(app string) (*TemplateFile, bool) {
	for _, template := range d.Templates {
		if template.app == app {
			return template, true
		}
	}
	return nil, false
}

func (d Discovery) GetAppTemplatePatch(installation, app string) (*TemplateFile, bool) {
	for _, template := range d.TemplatePatches {
		if template.installation == installation && template.app == app {
			return template, true
		}
	}
	return nil, false
}

func (d *Discovery) populateValuePaths() error {
	// 1. Mark all overshadowed valuePaths in config.yaml
	for _, configPatch := range d.ConfigPatches {
		for path := range configPatch.paths {
			if original, ok := d.Config.paths[path]; ok {
				original.OvershadowedBy = append(original.OvershadowedBy, configPatch)
			}
		}
	}
	// 2. Render templates for all apps x installations, then set UsedBy fields
	// in config or config patches.
	for _, installation := range d.Installations {
		for _, app := range d.Apps {
			configPatch, ok := d.GetConfigPatch(installation)
			if !ok {
				configPatch = nil
			}

			// mark all fields used by the templatePatch
			if templatePatch, ok := d.GetAppTemplatePatch(installation, app); ok {
				populatePathsWithSource(templatePatch, d.Config, configPatch)
			}

			// mark all fields used by the defaultTemplate
			if defaultTemplate, ok := d.GetAppTemplate(app); ok {
				populatePathsWithSource(defaultTemplate, d.Config, configPatch)
			}
		}
	}

	return nil
}

func populatePathsWithSource(source *TemplateFile, config, configPatch *ValueFile) {
	for path, templatePath := range source.values {
		valuePath, valuePathOk := configPatch.paths[path]
		if configPatch != nil && valuePathOk {
			// config patch exists and contains the path
			valuePath.UsedBy = appendUniqueUsedBy(valuePath.UsedBy, source)
			continue
		}

		valuePath, valuePathOk = config.paths[path]
		if valuePathOk {
			// the value comes from default config
			valuePath.UsedBy = appendUniqueUsedBy(valuePath.UsedBy, source)
			continue
		}

		// value is missing from config; linter will check if it's patched
		templatePath.MayBeMissing = true
	}
}

func NewDiscovery(fs generator.Filesystem) (*Discovery, error) {
	d := &Discovery{
		ConfigPatches:   []*ValueFile{},
		Templates:       []*TemplateFile{},
		TemplatePatches: []*TemplateFile{},

		Installations:       []string{},
		Apps:                []string{},
		AppsPerInstallation: map[string][]string{},
	}

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
	}

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
	}

	for _, inst := range installationDirs {
		if !inst.IsDir() {
			continue
		}
		d.AppsPerInstallation[inst.Name()] = []string{}
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
		}
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

func appendUniqueUsedBy(list []*TemplateFile, t *TemplateFile) []*TemplateFile {
	for _, v := range list {
		if v == t {
			return list
		}
	}
	return append(list, t)
}
