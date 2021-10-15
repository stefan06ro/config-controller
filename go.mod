module github.com/giantswarm/config-controller

go 1.14

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/fatih/color v1.13.0
	github.com/ghodss/yaml v1.0.0
	github.com/giantswarm/apiextensions/v3 v3.18.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v5 v5.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v4 v4.0.0
	github.com/giantswarm/valuemodifier v0.3.1
	github.com/go-git/go-billy/v5 v5.3.1
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-test/deep v1.0.7 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/go-retryablehttp v0.6.7 // indirect
	github.com/hashicorp/vault/api v1.2.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/tidwall/pretty v1.0.1 // indirect
	go.mongodb.org/mongo-driver v1.4.2 // indirect
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v3.2.1+incompatible
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2 // CVE-2021-3121
	github.com/opencontainers/runc v0.1.1 => github.com/opencontainers/runc v1.0.0-rc93
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
)
