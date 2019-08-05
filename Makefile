REPO=marraison
SERVICE=ninepd
VERSION=0.0.1

help:
	@echo "Usage: make COMMAND\n\nThe commands are:\n"
	@echo "  build-image - build ninepd daemon docker image"
	@echo "  push-image  - push ninepd image to dockerhub"
	@echo "  test        - test package"
	@echo ""

test: build-test-image
	@docker run ninep-unittest test -tags=compat -v ./...

build-test-image:
	@docker build --rm --tag=ninep-unittest \
		--file scripts/Dockerfile.unittest .

build-image:
	@docker build --rm --tag=$(REPO)/$(SERVICE):$(VERSION) .

push-image: build-image
	@docker push $(REPO)/$(SERVICE):$(VERSION)
