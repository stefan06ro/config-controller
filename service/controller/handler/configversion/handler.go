package configversion

import (
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/config-controller/service/internal/github"
)

const (
	Name  = "configversion"
	owner = "giantswarm"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	GitHubToken string
}

type Handler struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	gitHub *github.GitHub
}

func New(config Config) (*Handler, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.GitHubToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GitHubToken must not be empty", config)
	}

	gh, err := github.New(github.Config{
		GitHubToken: config.GitHubToken,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	h := &Handler{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
		gitHub:    gh,
	}

	return h, nil
}

func (h *Handler) Name() string {
	return Name
}

type Index struct {
	APIVersion string                  `json:"apiVersion"`
	Entries    map[string][]IndexEntry `json:"entries"`
}

type IndexEntry struct {
	APIVersion    string   `json:"apiVersion"`
	AppVersion    string   `json:"appVersion"`
	ConfigVersion string   `json:"configVersion,omitempty"`
	Created       string   `json:"created"`
	Description   string   `json:"description"`
	Digest        string   `json:"digest"`
	Home          string   `json:"home"`
	Name          string   `json:"name"`
	Urls          []string `json:"urls"`
	Version       string   `json:"version"`
}
