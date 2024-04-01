##@ Helm package
HELM_CHART         ?= kdp-oam-operator
HELM_CHART_VERSION ?= $(VERSION)

.PHONY: helm-package
helm-package:   ## Helm package
	cd charts && $(HELMBIN) package $(HELM_CHART) --version $(HELM_CHART_VERSION) --app-version $(HELM_CHART_VERSION)
