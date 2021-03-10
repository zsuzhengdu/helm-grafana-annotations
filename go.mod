module github.com/zsuzhengdu/grafana-annotations

go 1.15

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)

require (
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/flagenv v0.0.0-20160425205200-fcd59fca7456
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/gobs/pretty v0.0.0-20180724170744-09732c25a95b
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/knadh/koanf v0.15.0
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/orcaman/concurrent-map v0.0.0-20210106121528-16402b402231
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.8.0
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.5.2
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.2
)
