.DEFAULT_GOAL := image
DOCKER_IMAGE ?= cr.yandex/yc/fluent-bit-plugin-yandex
PLUGIN_VERSION ?= dev
FLUENT_BIT_VERSION?=1.8.6

push-image:
	docker push $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_VERSION)

image: mod.vendor
	docker build \
		--build-arg fluent_bit_version=$(FLUENT_BIT_VERSION) \
		-t $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_VERSION) .

precommit: mod.tidy fmt vet

vet:
	go vet ./...

fmt:
	goimports -w -format-only .

mod.vendor: mod.tidy
	go mod vendor

mod.tidy:
	go mod tidy