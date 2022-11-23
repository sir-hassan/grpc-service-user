PKG_LIST := $(shell go list ./...)
IMG_NAME := grpc-service-user
BUF_VERSION:=latest


test:
	test -z '$(shell gofmt -s -l .)'
	go vet ./...
	
	go test ./... -v
	mkdir -p bin
	go test ./... -v --coverprofile bin/coverage.txt

	go test -race -short $(PKG_LIST)
	go tool cover -func bin/coverage.txt

	@echo "✅  Successful test run."

test-e2e: composer-down
	docker build -t $(IMG_NAME) .
	IMG_NAME=$(IMG_NAME) docker-compose run e2e || make composer-down

build: test
	mkdir -p bin
	CGO_ENABLED=0 GOARCH=amd64 go  build -gcflags="all=-N -l" -o bin/app cmd/*

lint:
	gofumpt -w . && golangci-lint run --fix


composer-down:
	IMG_NAME=$(IMG_NAME) docker-compose down --remove-orphans -v

composer-up: composer-down
	docker build -t $(IMG_NAME) .
	IMG_NAME=$(IMG_NAME) docker-compose up || make composer-down

generate:
	docker run -v $$(pwd):/app -w /app bufbuild/buf:$(BUF_VERSION) generate