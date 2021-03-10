# grafana-annotations

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

Adding annotations for helm deployments

**Homepage:** <https://github.com/zsuzhengdu/grafana-annotations>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Du Zheng | zsuzhengdu@gmail.com |  |

## Source Code  

* <https://github.com/zsuzhengdu/grafana-annotations>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| auth | string | `"admin:admin"` |  |
| config.helmRegistries.overrideChartNames | object | `{}` |  |
| config.helmRegistries.override[0].allowAllReleases | bool | `true` |  |
| config.helmRegistries.override[0].charts | list | `[]` |  |
| config.helmRegistries.override[0].registry.url | string | `""` |  |
| dashboardID | int | `0` |  |
| fullnameOverride | string | `"kube-prometheus-stack-grafana-annotations"` |  |
| grafanaIP | string | `"127.0.0.1"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"zsuzhengdu/grafana-annotations"` |  |
| image.tag | string | `"0.0.1"` |  |
| imagePullSecrets | list | `[]` |  |
| infoMetric | bool | `true` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths | list | `[]` |  |
| ingress.tls | list | `[]` |  |
| intervalDuration | string | `"5s"` |  |
| latestChartVersion | bool | `true` |  |
| nameOverride | string | `""` |  |
| namespaces | string | `""` |  |
| nodeSelector | object | `{}` |  |
| panelID | int | `0` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| rbac.create | bool | `true` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| service.annotations | object | `{}` |  |
| service.port | int | `9571` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `nil` |  |
| serviceMonitor.additionalLabels | object | `{}` |  |
| serviceMonitor.create | bool | `true` |  |
| serviceMonitor.interval | string | `nil` |  |
| serviceMonitor.namespace | string | `"default"` |  |
| serviceMonitor.scrapeTimeout | string | `nil` |  |
| timestampMetric | bool | `true` |  |
| tolerations | list | `[]` |  |

