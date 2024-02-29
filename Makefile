.PHONY: \
	build \
	deploy \
	test

build:
	docker buildx build --platform=linux/amd64 \
		-f Dockerfile \
		-t dotcom \
		../../

test: 
	go test -timeout 10s -cover -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out | \
		perl -an -E 'die "$$F[2] coverage does not meet threshold of ${COVERAGE_MIN}
		%\n" if /total/ && $$F[2] < ${COVERAGE_MIN}'