.PHONY: build

#for examples
# Current  version
VERSION ?=latest
# Default image tag

IMG ?= quay.io/jenchen/ptp-event-consumer:$(VERSION)

export GO111MODULE=on
export CGO_ENABLED=1
export GOFLAGS=-mod=vendor
export COMMON_GO_ARGS=-race

OS := $(shell uname -s)
ifeq ($(OS), Darwin)
export GOOS=darwin
else
export GOOS=linux
endif

ifeq (,$(shell go env GOBIN))
  GOBIN=$(shell go env GOPATH)/bin
else
  GOBIN=$(shell go env GOBIN)
endif

export GOPATH=$(shell go env GOPATH)

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Versions
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= v4.5.7
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }


deps-update:
	go mod tidy && \
	go mod vendor

docker-build: #test ## Build docker image with the manager.
	docker build -f Containerfile -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

# Deploy all in the configured Kubernetes cluster in ~/.kube/config
deploy:kustomize
	cd ./deployment && $(KUSTOMIZE) edit set image ptp-event-consumer=${IMG}
	$(KUSTOMIZE) build ./deployment | kubectl apply -f -

undeploy:kustomize
	cd ./deployment && $(KUSTOMIZE) edit set image ptp-event-consumer=${IMG}
	$(KUSTOMIZE) build ./deployment | kubectl delete -f -
