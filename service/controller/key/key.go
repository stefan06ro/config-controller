package key

import (
	"regexp"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	// PauseAnnotation stops app-operator from reconciling App CR too soon.
	// Once config has been generated, this annotation should be removed by
	// config-controller.
	PauseAnnotation = "app-operator.giantswarm.io/paused"
)

var (
	tagConfigVersionPattern = regexp.MustCompile(`^(\d+)\.x\.x$`)
)

func ToAppCR(v interface{}) (v1alpha1.App, error) {
	if v == nil {
		return v1alpha1.App{}, microerror.Maskf(wrongTypeError, "expected non-nil, got %#v", v)
	}

	p, ok := v.(*v1alpha1.App)
	if !ok {
		return v1alpha1.App{}, microerror.Maskf(wrongTypeError, "expected %T, got %T", p, v)
	}

	c := p.DeepCopy()

	return *c, nil
}

// TryVersionToTag translates config version: "<major>.x.x" to tagReference:
// "v<major>" if possible. Otherwise returns empty string.
func TryVersionToTag(version string) string {
	matches := tagConfigVersionPattern.FindAllStringSubmatch(version, -1)
	if len(matches) > 0 {
		return "v" + matches[0][1]
	}
	return ""
}

func RemoveAnnotation(annotations map[string]string, key string) map[string]string {
	if annotations == nil {
		return nil
	}

	out := map[string]string{}
	for k, v := range annotations {
		if k == key {
			continue
		}
		out[k] = v
	}

	return out
}
