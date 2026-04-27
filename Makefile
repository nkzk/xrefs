GOLANGCI_LINT_VERSION ?= v2.11.4
GOLANGCI_LINT         = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

dev:
	go run ./cmd

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)

.PHONY: lint
lint: golangci-lint
		$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint
		$(GOLANGCI_LINT) run --fix

.PHONY: fmt
fmt:
		go fmt ./...

.PHONY: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): | $(LOCALBIN)
		$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))



# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
# $4 - build tags
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
if [ -n "$(4)" ]; then \
    GOBIN=$(LOCALBIN) go install -tags=$(4) $${package}; \
else \
    GOBIN=$(LOCALBIN) go install $${package}; \
fi ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
