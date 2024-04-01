/*
Copyright 2023 KDP(Kubernetes Data Platform).

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package application

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	"kdp-oam-operator/pkg/utils"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"time"
)

var _ = Describe("Test Application Controller", func() {
	ctx := context.Background()

	BeforeEach(func() {
		var validDef = utils.ReadContent(filepath.Join("../../../../..", "charts", "kdp-oam-operator", "templates", "deftemplate", "application-def.yaml"))

		// Create namespace
		ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kdp-system"}}

		Eventually(func() error {
			return k8sClient.Create(ctx, &ns)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))

		// Create bdc definition map into configmap
		bdcDefinitionMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "application-definition-map",
				Namespace: "kdp-system",
			},
			Data: map[string]string{
				"default-application": "application-def",
			},
		}
		Eventually(func() error {
			return k8sClient.Create(ctx, &bdcDefinitionMap)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))

		// Create xDefinition
		var def bdcv1alpha1.XDefinition
		err := yaml.Unmarshal([]byte(validDef), &def)
		if err != nil {
			return
		}

		Eventually(func() error {
			return k8sClient.Create(ctx, &def)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil()))
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "application-def"}, &def)).Should(Succeed())
	})

	Context("When the application dependent xDefinition doesn't exist, should occur error", func() {
		It("Applying application", func() {
			By("Apply application")

			var validApplicationInstance = `
apiVersion: bdc.kdp.io/v1alpha1
kind: Application
metadata:
  annotations:
    "bdc.kdp.io/org": "bdctestorg"
  name: application-sample1
spec:
  name: application-sample1
  type: default
  properties:
    apiVersion: core.oam.dev/v1alpha1
    kind: Application
    metadata:
      name: zookeeper
      namespace: bdctestorg
    spec:
      components:
        - name: zookeeper
          properties:
            chart: zookeeper
            releaseName: zookeeper
            repoType: helm
            version: 7.6.2
            targetNamespace: admin
            url: https://addrres:port/repository/helm-hosted
            values:
              affinity:
                podAntiAffinity:
                  requiredDuringSchedulingIgnoredDuringExecution:
                    - labelSelector:
                        matchExpressions:
                          - key: app.kubernetes.io/instance
                            operator: In
                            values:
                              - zookeeper
                      topologyKey: kubernetes.io/hostname
              autopurge:
                purgeInterval: 1
                snapRetainCount: 3
              customLivenessProbe:
                exec:
                  command:
                    - /bin/bash
                    - '-c'
                    - curl -s -m 2 http://localhost:8080/commands/ruok | grep ruok
                failureThreshold: 6
                initialDelaySeconds: 30
                periodSeconds: 10
                successThreshold: 1
                timeoutSeconds: 5
              customReadinessProbe:
                exec:
                  command:
                    - /bin/bash
                    - '-c'
                    - >-
                      curl -s -m 2 http://localhost:8080/commands/ruok | grep error
                      | grep null
                failureThreshold: 6
                initialDelaySeconds: 5
                periodSeconds: 10
                successThreshold: 1
                timeoutSeconds: 5
              extraVolumeMounts:
                - mountPath: /opt/bitnami/zookeeper/logs
                  name: logs
              extraVolumes:
                - emptyDir: {}
                  name: logs
                - configMap:
                    defaultMode: 420
                    name: promtail-conf
                  name: promtail-conf
              global:
                imageRegistry: docker-image-registry
                storageClass: lvm-localpv
              heapSize: 384
              image:
                tag: 3.8.2
              livenessProbe:
                enabled: false
              log4jProp: INFO, ROLLINGFILE
              logLevel: INFO
              metrics:
                enabled: true
              persistence:
                enabled: true
                size: 8Gi
              readinessProbe:
                enabled: false
              replicaCount: 3
              resources:
                limits:
                  cpu: '1.0'
                  memory: 512Mi
                requests:
                  cpu: '0.1'
                  memory: 512Mi
                  image: docker-image-registry/grafana/promtail:2.5.0
                  name: logs-promtail-sidecar
                  resources:
                    limits:
                      cpu: 100m
                      memory: 128Mi
                    requests:
                      cpu: 100m
                      memory: 128Mi
                  volumeMounts:
                    - mountPath: /var/log/admin-zookeeper
                      name: logs
                    - mountPath: /etc/promtail
                      name: promtail-conf
`
			var application bdcv1alpha1.Application
			yaml.Unmarshal([]byte(validApplicationInstance), &application)
			Expect(yaml.Unmarshal([]byte(validApplicationInstance), &application)).Should(BeNil())
			Expect(k8sClient.Create(ctx, &application)).Should(Succeed())
		})
	})
})
