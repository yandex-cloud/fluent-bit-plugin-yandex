.DEFAULT_GOAL := images
DOCKER_IMAGE ?= cr.yandex/yc/fluent-bit-plugin-yandex
PLUGIN_VERSION ?= dev
#FLUENT_BIT_1_8?=1.8.15
FLUENT_BIT_1_9?=1.9.10
#FLUENT_BIT_2_0?=2.0.11
FLUENT_BIT_2_1?=2.1.7

push-images:
	#docker push $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_8)
	docker push $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_9)
	#docker push $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_2_0)
	docker push $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_2_1)

images: #mod.vendor
	#docker build \
#		--platform linux/amd64 \
#		--build-arg plugin_version=$(PLUGIN_VERSION) \
#		--build-arg fluent_bit_version=$(FLUENT_BIT_1_8) \
#		-t $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_8) .
	docker build \
		--platform linux/amd64 \
		--build-arg plugin_version=$(PLUGIN_VERSION) \
		--build-arg fluent_bit_version=$(FLUENT_BIT_1_9) \
		-t $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_1_9) .
	#docker build \
#		--platform linux/amd64 \
#		--build-arg plugin_version=$(PLUGIN_VERSION) \
#		--build-arg fluent_bit_version=$(FLUENT_BIT_2_0) \
#		-t $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_2_0) .
	docker build \
		--platform linux/amd64 \
		--build-arg plugin_version=$(PLUGIN_VERSION) \
		--build-arg fluent_bit_version=$(FLUENT_BIT_2_1) \
		-t $(DOCKER_IMAGE):$(PLUGIN_VERSION)-fluent-bit-$(FLUENT_BIT_2_1) .

precommit: mod.tidy fmt vet lint

vet:
	go vet ./...

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

mod.vendor: mod.tidy
	go mod vendor

mod.tidy:
	go mod tidy
