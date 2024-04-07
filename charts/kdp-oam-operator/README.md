[![Helm Chart Pulls](https://img.shields.io/docker/pulls/linktimecloud/kdp-oam-operator-chart?label=Helm%20Chart%20Pulls)](https://hub.docker.com/r/linktimecloud/kdp-oam-operator-chart)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# KDP-OAM-Operator helm chart

The kdp-oam-operator, based on the KDP (Kubernetes Data Platform) bigdata model, provides a highly programmable application-centric delivery model. It allows users to build upon the fundamental capabilities of the kdp-oam-operator to extend and customize their services. The most critical components in the kdp-oam-operator are the XDefinition, Application, BDC, ContextSetting, and ContextSecret. All of these are built upon the fundamental structure of the Kubernetes Custom Resource Definition (CRD) model and manage the resource lifecycle through the implementation of independent controller components.

## Helm Commands

To pull this chart from the repository,

```bash
helm pull oci://registry-1.docker.io/linktimecloud/kdp-oam-operator-chart --version v1.0.0-rc1
```

Other Commands,

```bash
helm show all oci://registry-1.docker.io/linktimecloud/kdp-oam-operator-chart --version v1.0.0-rc1
helm template <my-release> oci://registry-1.docker.io/linktimecloud/kdp-oam-operator-chart --version v1.0.0-rc1
helm install <my-release> oci://registry-1.docker.io/linktimecloud/kdp-oam-operator-chart --version v1.0.0-rc1
helm upgrade <my-release> oci://registry-1.docker.io/linktimecloud/kdp-oam-operator-chart --version <new-version>
```

## Prerequisites
- Kubernetes 1.23+

## Parameters

### kdp-oam-operator core parameters

| Name                         | Description                                                              | Value                  |
| ---------------------------- | ------------------------------------------------------------------------ | ---------------------- |
| `images.tag`                 | Common tag for kdp-oam-operator images. Defaults to `.Chart.AppVersion`. | `""`                   |
| `images.registry`            | Registry to use for the controller                                       | `""`                   |
| `images.pullPolicy`          | imagePullPolicy to apply to all containers                               | `IfNotPresent`         |
| `images.pullSecrets`         | Secrets with credentials to pull images from a private registry          | `[]`                   |
| `systemNamespace.create`     | Specifies whether a system namespace should be created                   | `false`                |
| `systemNamespace.name`       | The name of the system namespace                                         | `kdp-system`           |
| `kdpContextLabel.key`        | The key of the label to use for the kdp context                          | `kdp-operator-context` |
| `kdpContextLabel.value`      | The value of the label to use for the kdp context                        | `KDP`                  |
| `nameOverride`               | String to partially override `kdp-oam-operator.fullname` template        | `""`                   |
| `fullnameOverride`           | String to fully override `kdp-oam-operator.fullname` template            | `kdp-oam-operator`     |
| `rbac.create`                | Specifies whether a RBAC role should be created                          | `true`                 |
| `serviceAccount.name`        | The name of the service account to use.                                  | `kdp-oam-operator-sa`  |
| `serviceAccount.annotations` | Annotations to add to the service account                                | `{}`                   |
| `serviceAccount.create`      | Specifies whether a service account should be created                    | `true`                 |
| `podAnnotations`             | Annotations to add to the pod                                            | `{}`                   |
| `podSecurityContext`         | Security context to add to the pod                                       | `{}`                   |
| `securityContext`            | Security context to add to the pod                                       | `{}`                   |
| `nodeSelector`               | Node selector to add to the pod                                          | `{}`                   |
| `tolerations`                | Tolerations to add to the pod                                            | `[]`                   |
| `affinity`                   | Affinity to add to the pod                                               | `{}`                   |

### kdp-oam-apiserver parameters

| Name                                   | Description                                                                     | Value                        |
| -------------------------------------- | ------------------------------------------------------------------------------- | ---------------------------- |
| `apiserver.enabled`                    | Specifies whether the apiserver should be enabled                               | `true`                       |
| `apiserver.replicaCount`               | Number of replicas for the apiserver                                            | `1`                          |
| `apiserver.service.type`               | Service type for the apiserver                                                  | `ClusterIP`                  |
| `apiserver.service.port`               | Port to use for the apiserver                                                   | `8000`                       |
| `apiserver.image.repository`           | Repository to use for the apiserver                                             | `kdp-oam-operator/apiserver` |
| `apiserver.image.tag`                  | Image tag for the kdp-oam-operator apiserver. Defaults to `.Values.images.tag`. | `""`                         |
| `apiserver.resources`                  | Resources to add to the apiserver                                               | `{}`                         |
| `apiserver.env`                        | Environment variables to add to the apiserver                                   | `[]`                         |
| `apiserver.extraArgs`                  | Extra arguments to pass to the apiserver                                        | `[]`                         |
| `apiserver.serviceAccount.create`      | Specifies whether a service account should be created                           | `true`                       |
| `apiserver.serviceAccount.annotations` | Annotations to add to the service account                                       | `{}`                         |
| `apiserver.serviceAccount.name`        | The name of the service account to use.                                         | `kdp-oam-apiserver-sa`       |

### kdp-oam-controller parameters

| Name                                       | Description                                                                      | Value                               |
| ------------------------------------------ | -------------------------------------------------------------------------------- | ----------------------------------- |
| `controller.replicaCount`                  | Number of replicas for the controller                                            | `1`                                 |
| `controller.image.repository`              | Repository to use for the controller                                             | `kdp-oam-operator/controller`       |
| `controller.image.tag`                     | Image tag for the kdp-oam-operator controller. Defaults to `.Values.images.tag`. | `""`                                |
| `controller.metricService.enabled`         | Specifies whether the metric service should be enabled                           | `true`                              |
| `controller.metricService.type`            | Service type for the metric service                                              | `ClusterIP`                         |
| `controller.metricService.port`            | Port for the metric service                                                      | `8080`                              |
| `controller.healthzService.type`           | Service type for the healthz service                                             | `ClusterIP`                         |
| `controller.healthzService.port`           | Port for the healthz service                                                     | `9440`                              |
| `controller.Resources.limits.cpu`          | CPU limits to add to the controller Requests to add to the controller            | `500m`                              |
| `controller.Resources.limits.memory`       | Memory limits to add to the controller                                           | `500Mi`                             |
| `controller.Resources.requests.cpu`        | CPU requests to add to the controller                                            | `10m`                               |
| `controller.Resources.requests.memory`     | Memory requests to add to the controller                                         | `64Mi`                              |
| `controller.env`                           | Environment variables to add to the controller                                   | `[]`                                |
| `controller.extraArgs`                     | Extra arguments to pass to the controller                                        | `[]`                                |
| `controller.args.concurrentReconciles`     | Number of concurrent reconciles to use for the controller                        | `5`                                 |
| `controller.args.reSyncPeriod`             | Reconcile period to use for the controller                                       | `10s`                               |
| `admissionWebhooks.enabled`                | Specifies whether admission webhooks should be enabled                           | `false`                             |
| `admissionWebhooks.service.type`           | Service type for the admission webhooks                                          | `ClusterIP`                         |
| `admissionWebhooks.service.port`           | Port for the admission webhooks                                                  | `9443`                              |
| `admissionWebhooks.failurePolicy`          | Failure policy for the admission webhooks                                        | `Fail`                              |
| `admissionWebhooks.certManager.enabled`    | Specifies whether cert-manager should be used for the admission webhooks         | `false`                             |
| `admissionWebhooks.certificate.mountPath`  | Path to mount the certificate                                                    | `/k8s-webhook-server/serving-certs` |
| `admissionWebhooks.patch.enabled`          | Specifies whether the patch should be enabled                                    | `true`                              |
| `admissionWebhooks.patch.image.repository` | Repository to use for the patch                                                  | `oamdev/kube-webhook-certgen`       |
| `admissionWebhooks.patch.image.tag`        |                                                                                  | `v2.4.1`                            |
| `admissionWebhooks.patch.image.pullPolicy` | imagePullPolicy to apply to the patch                                            | `IfNotPresent`                      |
| `admissionWebhooks.patch.nodeSelector`     | Node selector to add to the patch                                                | `{}`                                |
| `admissionWebhooks.patch.affinity`         | Affinity to add to the patch                                                     | `{}`                                |
| `admissionWebhooks.patch.tolerations`      | Tolerations to add to the patch                                                  | `[]`                                |
