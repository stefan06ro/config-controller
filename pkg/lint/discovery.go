package lint

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/generator"
)

type Discovery struct {
	Config          *ValueFile
	ConfigPatches   []*ValueFile
	Templates       []*TemplateFile
	TemplatePatches []*TemplateFile

	Installations []string
	Apps          []string
}

func NewDiscovery(fs generator.Filesystem) (*Discovery, error) {
	d := &Discovery{
		ConfigPatches:   []*ValueFile{},
		Templates:       []*TemplateFile{},
		TemplatePatches: []*TemplateFile{},

		Installations: []string{},
		Apps:          []string{},
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
		filepath := fmt.Sprintf("installation/%s/config.yaml.patch", inst.Name())
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
		appDirs, err := fs.ReadDir("default/apps/")
		if err != nil {
			return nil, microerror.Mask(err)
		}
		for _, app := range appDirs {
			if !app.IsDir() {
				continue
			}
			uniqueApps[app.Name()] = true
			filepath := fmt.Sprintf("installation/%s/apps/%s/configmap-values.yaml.patch", inst.Name(), app.Name())
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

	for k, _ := range uniqueInstallations {
		d.Installations = append(d.Installations, k)
	}
	for k, _ := range uniqueApps {
		d.Apps = append(d.Apps, k)
	}

	return d, nil
}

func (d Discovery) X() {

}
