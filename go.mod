module github.com/giantswarm/config-controller

go 1.14

require (
	github.com/CloudyKit/jet v2.1.3-0.20180809161101-62edd43e4f88+incompatible // indirect
	github.com/Joker/jade v1.0.1-0.20190614124447-d475f43051e7 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/fatih/color v1.9.0 // indirect
	github.com/flosch/pongo2 v0.0.0-20190707114632-bbf5a6c351f4 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions/v3 v3.13.0
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v5 v5.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.4.0
	github.com/giantswarm/operatorkit/v4 v4.2.0
	github.com/giantswarm/valuemodifier v0.3.0
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-test/deep v1.0.7 // indirect
	github.com/google/go-cmp v0.5.4
	github.com/hashicorp/go-retryablehttp v0.6.7 // indirect
	github.com/hashicorp/vault/api v1.0.5-0.20201001211907-38d91b749c77
	github.com/hashicorp/vault/sdk v0.1.14-0.20201109203410-5e6e24692b32 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/iris-contrib/i18n v0.0.0-20171121225848-987a633949d0 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/mediocregopher/mediocre-go-lib v0.0.0-20181029021733-cb65787f37ed // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/tidwall/pretty v1.0.1 // indirect
	go.mongodb.org/mongo-driver v1.4.2 // indirect
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a // indirect
	golang.org/x/tools v0.0.0-20200521155704-91d71f6c2f04 // indirect
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
	sigs.k8s.io/controller-runtime v0.6.4
)

replace sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
