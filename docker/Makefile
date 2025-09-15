GENCERT_CMD := $(PWD)/../scripts/gencert.sh
BUILD_OPTIONS ?= -ldflags "-s -w" -trimpath
CN ?= localhost
APP_PORT ?= 8080

all: up

.build.done:
	-rm -fr ./bin
	GOOS=linux GOARCH=amd64 go build $(BUILD_OPTIONS) -o ./bin/app ../.
	#upx --lzma --force-overwrite -q --no-progress ./bin/app
	@touch $@

.cert.done:
	mkdir -pm 0700 cert
	cd cert && CN=$(CN) $(GENCERT_CMD)
	chmod 0755 ./cert/cert
	chmod 0644 ./cert/cert/*
	@touch $@

.build-image.done:
	docker build -t green-api-proxy:latest .
	@touch $@

build: .build.done

cert: .cert.done

build-image: .build-image.done

up: cert build build-image
	docker run -d --restart=unless-stopped -p $(APP_PORT):$(APP_PORT) \
		-e SERVER_ADDR=:$(APP_PORT) -e LOG_LEVEL=$(LOG_LEVEL) \
		--name green-api-proxy green-api-proxy

down: stop
	-docker container rm green-api-proxy

clean: down
	-docker image rm green-api-proxy
	-rm -fr ./bin ./cert ./.*.done

start:
	docker start green-api-proxy
	
stop:
	-docker stop green-api-proxy