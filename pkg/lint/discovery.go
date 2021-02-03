package lint

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/config-controller/pkg/generator"
)

type Discovery struct {
	config          *ValueFile
	configPatches   []*ValueFile
	templates       []*TemplateFile
	templatePatches []*TemplateFile
}

func NewDiscovery(fs generator.Filesystem) (*Discovery, error) {
	d := &Discovery{
		configPatches:   []*ValueFile{},
		templates:       []*TemplateFile{},
		templatePatches: []*TemplateFile{},
	}

	{
		filepath := "default/config.yaml"
		body, err := fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Maskf(err, "cannot find %q", filepath)
		}
		d.config, err = NewValueFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	installationDirs, err := fs.ReadDir("installations/")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, inst := range installationDirs {
		if !inst.IsDir() {
			continue
		}
		filepath := fmt.Sprintf("installation/%s/config.yaml.patch", inst.Name())
		body, err := fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Maskf(err, "cannot find %q", filepath)
		}
		patch, err := NewValueFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		d.configPatches = append(d.configPatches, patch)
	}

	defaultAppDirs, err := fs.ReadDir("default/apps/")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	for _, app := range defaultAppDirs {
		if !app.IsDir() {
			continue
		}
		filepath := fmt.Sprintf("default/apps/%s/configmap-values.yaml.template", app.Name())
		body, err := fs.ReadFile(filepath)
		if err != nil {
			return nil, microerror.Maskf(err, "cannot find %q", filepath)
		}
		template, err := NewTemplateFile(filepath, body)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		d.templates = append(d.templates, template)
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
			filepath := fmt.Sprintf("installation/%s/apps/%s/configmap-values.yaml.patch", inst.Name(), app.Name())
			body, err := fs.ReadFile(filepath)
			if err != nil {
				return nil, microerror.Maskf(err, "cannot find %q", filepath)
			}
			templatePatch, err := NewTemplateFile(filepath, body)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			d.templatePatches = append(d.templatePatches, templatePatch)
		}
	}

	return d, nil
}
