##@ Build apiserver swagger docs

.PHONY: build-swagger
build-swagger: ## Generate api server swagger docs(./docs/apidoc/swagger.json)
	go run ./cmd/apiserver/main.go build-swagger ./docs/apidoc/swagger.json