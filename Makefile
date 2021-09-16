.DEFAULT_GOAL := images
DOCKER_IMAGE ?= cr.yandex/yc/fluent-bit-plugin-yandex
PLUGIN_VERSION ?= dev
FLUENT_BIT_1_6?=1.6.10
FLUENT_BIT_1_7?=1.7.9
FLUENT_BIT_1_8?=1.8.6

push-images:
	docker push $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_6)
	docker push $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_7)
	docker push $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_8)

images: mod.vendor
	docker build \
		--build-arg plugin_version=$(PLUGIN_VERSION) \
		--build-arg fluent_bit_version=$(FLUENT_BIT_1_6) \
		-t $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_6) .
	docker build \
		--build-arg plugin_version=$(PLUGIN_VERSION) \
		--build-arg fluent_bit_version=$(FLUENT_BIT_1_7) \
		-t $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_7) .
	docker build \
		--build-arg plugin_version=$(PLUGIN_VERSION) \
		--build-arg fluent_bit_version=$(FLUENT_BIT_1_8) \
		-t $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_8) .

precommit: mod.tidy fmt vet

vet:
	go vet ./...

fmt:
	goimports -w -format-only .

mod.vendor: mod.tidy
	go mod vendor

mod.tidy:
	go mod tidy