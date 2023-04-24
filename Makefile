
test:
	docker run --rm \
		-w /app \
		-v `pwd`:/app/ \
		docker.chotot.org/golang-builder:1.13.7-alpine \
		go test -v ./... -coverprofile=coverage.out
