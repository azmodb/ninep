test: build-test-image
	@docker run ninep-unittest test -tags=compat -v ./...

build-test-image:
	@docker build --rm --tag=ninep-unittest .
