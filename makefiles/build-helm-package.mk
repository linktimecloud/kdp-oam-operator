##@ Helm package
HELM_CHART         ?= kdp-oam-operator
HELM_CHART_VERSION ?= $(VERSION)

.PHONY: helm-package
helm-package:   ## Helm package
	cd charts && $(HELMBIN) package $(HELM_CHART) --version $(HELM_CHART_VERSION) --app-version $(HELM_CHART_VERSION)



.PHONY: helm-doc-gen
helm-doc-gen: helm-doc  ## helm-doc-gen: Generate helm chart README.md
	readme-generator -v charts/kdp-oam-operator/values.yaml -r charts/kdp-oam-operator/README.md