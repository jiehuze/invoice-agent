APP = $(shell pwd | awk -F '/' '{print $$NF}')
OS = $(shell uname| awk '{print tolower($0)}')
BUILD_INFO = $(shell date  +%Y年%m月%d日-%H:%M:%S)
VERSION_INFO = ${tag}
OUT_PATH = "output"

REGISTRY ?= lavie-mirror-registry.cn-beijing.cr.aliyuncs.com/lavie_service
IMAGE = $(REGISTRY)/$(APP)

ifdef version
	VERSION_INFO = $(version)
endif

ifdef os
	OS = $(os)
endif

LDFLAGS = -ldflags "-X main.BuildInfo=$(BUILD_INFO)	\
				          -X main.VersionInfo=$(VERSION_INFO)"

all: clean $(APP) tarball

$(APP):
	@echo "building $(APP) ..."
	#GOOS=$(OS) GOARCH=amd64 go build $(LDFLAGS) -o $(OUT_PATH)/$(APP)  main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUT_PATH)/$(APP)  main.go

tarball:
	cp -rf conf $(OUT_PATH)/
	cp run.sh $(OUT_PATH)/

clean:
	rm -rf $(APP)
	rm -rf $(OUT_PATH)/*

docker: docker-build

docker-build:
	docker build --rm --build-arg app=$(APP) -t $(IMAGE):$(VERSION_INFO) .

#docker-push:
#	docker push $(IMAGE):$(VERSION_INFO)
#
#docker-clean:
#	docker rmi $(IMAGE):$(VERSION_INFO)
#	docker image prune -f
