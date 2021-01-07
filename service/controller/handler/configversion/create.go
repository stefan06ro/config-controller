package configversion

import (
	"context"

	"github.com/ghodss/yaml"
	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/microerror"

	controllerkey "github.com/giantswarm/config-controller/service/controller/key"
)

func (h *Handler) EnsureCreated(ctx context.Context, obj interface{}) error {
	app, err := controllerkey.ToAppCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	value, ok := app.GetAnnotations()[annotation.ConfigVersion]
	if ok {
		h.logger.Debugf(ctx, "App CR %#q is already annotated with %#q: %s", app.Name, annotation.ConfigVersion, value)
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}

	if app.Spec.Catalog == "" {
		h.logger.Debugf(ctx, "App CR %#q has no .Spec.Catalog set", app.Name)
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}

	if app.Spec.Catalog == "releases" {
		h.logger.Debugf(ctx, "App CR %#q has a \"releases\" catalog set", app.Name)
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}

	h.logger.Debugf(ctx, "setting App %#q config version", app.Spec.Name)

	h.logger.Debugf(ctx, "resolving config version for App %#q from %#q catalog", app.Name, app.Spec.Catalog)
	var configVersion string

	var index Index
	{
		store, err := h.gitHub.GetFilesByBranch(ctx, owner, app.Spec.Catalog, "master")
		if err != nil {
			return microerror.Mask(err)
		}

		indexYamlBytes, err := store.ReadFile("index.yaml")
		if err != nil {
			return microerror.Mask(err)
		}

		err = yaml.Unmarshal(indexYamlBytes, &index)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	entries, ok := index.Entries[app.Spec.Name]
	if !ok || len(entries) == 0 {
		h.logger.Debugf(ctx, "App %#q has no entries in %#q's index.yaml", app.Spec.Name, app.Spec.Catalog)
		h.logger.Debugf(ctx, "cancelling handler")
		return nil
	}

	configVersion = entries[0].ConfigVersion
	if configVersion == "" {
		configVersion = "0.0.0"
	}
	h.logger.Debugf(ctx, "resolved config version for App %#q from %#q catalog", app.Name, app.Spec.Catalog)

	annotations := app.GetAnnotations()
	annotations[annotation.ConfigVersion] = configVersion
	app.SetAnnotations(annotations)

	err = h.k8sClient.CtrlClient().Update(ctx, &app)
	if err != nil {
		return microerror.Mask(err)
	}
	h.logger.Debugf(ctx, "set App %#q config version to %#q", app.Spec.Name, configVersion)

	return nil
}
