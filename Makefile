# Current  Version
VERSION ?= v0.0.1-alpha
REGISTRY ?= changjjjjjjjj

# Image URL to use all building/pushing image targets
IMG_USER_MANAGER ?= $(REGISTRY)/user-manager:$(VERSION)

# Build the docker image
.PHONY: docker-build
docker-build: docker-build-user-manager

docker-build-user-manager:
	docker build . -f build/Dockerfile -t ${IMG_USER_MANAGER}

# Push the docker image
.PHONY: docker-push
docker-push: docker-push-user-manager

docker-push-user-manager:
	docker push ${IMG_USER_MANAGER}

# Test code lint
test-lint:
	golint ./src/...

# Unit test
test-unit:
	go test -v ./...
