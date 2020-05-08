APP_NAME = mdserver
BUILD_PATH = /tmp/${APP_NAME}
EPOCH_TIME = $(shell date +%s)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
GO_LDFLAGS = '-X "github.com/clintjedwards/${APP_NAME}/cmd.appVersion=$(VERSION)" \
			   -X "github.com/clintjedwards/${APP_NAME}/service.appVersion=$(VERSION)"'
SEMVER = 0.0.1
SHELL = /bin/bash
VERSION = ${SEMVER}_${EPOCH_TIME}_${GIT_COMMIT}

## build: run tests and compile full app in production mode
build-prod:
	go mod tidy
	go generate
	go build -ldflags $(GO_LDFLAGS) -o $(BUILD_PATH)

## help: prints this help message
help:
	@echo "Usage: "
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## install: build application and install on system
install: build-prod
	sudo mv BUILD_PATH /usr/local/bin/
	chmod +x /usr/local/bin/${APP_NAME}

## run: build application and run server; useful for dev
run: export DEBUG=true
run:
	go mod tidy
	go generate
	go build -ldflags $(GO_LDFLAGS) -o /tmp/${APP_NAME} && /tmp/${APP_NAME} server localhost:8080 -d /home/romeo/Documents/reference
