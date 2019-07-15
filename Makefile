test: build-test-image
	@docker run ninep-unittest test ./...

build-test-image:
	@docker build --rm --tag=ninep-unittest .
