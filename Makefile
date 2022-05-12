# Current  Version
VERSION ?= v0.0.1-alpha
REGISTRY ?= changjjjjjjjj

# Image URL to use all building/pushing image targets
IMG_USER_MANAGER ?= $(REGISTRY)/user-manager:$(VERSION)
IMG_POST_MANAGER ?= $(REGISTRY)/post-manager:$(VERSION)

# Build the docker image
.PHONY: docker-build
docker-build: docker-build-user-manager docker-build-post-manager

docker-build-user-manager:
	docker build . -f usermanagerservice/Dockerfile -t ${IMG_USER_MANAGER}

docker-build-post-manager:
	docker build . -f postmanagerservice/Dockerfile -t ${IMG_POST_MANAGER}

# Push the docker image
.PHONY: docker-push
docker-push: docker-push-user-manager docker-push-post-manager

docker-push-user-manager:
	docker push ${IMG_USER_MANAGER}

docker-push-post-manager:
	docker push ${IMG_POST_MANAGER}

# Test code lint
test-lint:
	golint ./...

# Unit test
test-unit: test-unit-user-manager test-unit-post-manager

test-unit-user-manager:
	go test -v ./usermanagerservice/...

test-unit-post-manager:
	go test -v ./postmanagerservice/...
