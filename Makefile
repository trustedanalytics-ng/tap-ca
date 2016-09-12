# Copyright (c) 2016 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
GOBIN=$(GOPATH)/bin
APP_DIR_LIST=$(shell go list ./... | grep -v /vendor/)
PROJECT_NAME=tap-ca
APP_NAME=tap-api-service

default: tests build_anywhere

fmt:
	go fmt $(APP_DIR_LIST)

build: verify_gopath
	go fmt $(APP_DIR_LIST)
	CGO_ENABLED=0 go install -tags netgo $(APP_DIR_LIST)
	mkdir -p application && cp -f $(GOBIN)/$(PROJECT_NAME) ./application/$(PROJECT_NAME)

deps_update_tap: verify_gopath
	$(GOBIN)/govendor update github.com/trustedanalytics-ng/...
	$(GOBIN)/govendor remove github.com/trustedanalytics-ng/$(PROJECT_NAME)/...
	@echo "Done"

clean:
	rm -rf ./data cfssl caservice temp application $(PROJECT_NAME) mockgen

verify_gopath:
	@if [ -z "$(GOPATH)" ] || [ "$(GOPATH)" = "" ]; then\
			echo "GOPATH not set. You need to set GOPATH before run this command";\
			exit 1 ;\
	fi

analyze_code:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	gometalinter --vendor ./...

prepare_dirs_for_mockgen:
	mkdir -p ./temp/src/github.com/golang/mock
	$(eval REPOFILES=$(shell pwd)/vendor/github.com/golang/mock/*)
	ln -sf $(REPOFILES) ./temp/src/github.com/golang/mock

prepare_dirs:
	rm -Rf application && mkdir application
	mkdir -p ./temp/src/github.com/trustedanalytics-ng/$(PROJECT_NAME)
	$(eval REPOFILES=$(shell pwd)/*)
	ln -sf $(REPOFILES) temp/src/github.com/trustedanalytics-ng/$(PROJECT_NAME)

	mkdir -p ./temp/src/github.com/cloudflare/cfssl
	$(eval REPOFILES=$(shell pwd)/vendor/github.com/cloudflare/cfssl/*)
	ln -sf $(REPOFILES) temp/src/github.com/cloudflare/cfssl

build_anywhere: prepare_dirs
	cp -R templates ./application

	$(eval GOPATH=$(shell cd ./temp; pwd))
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -tags netgo ./temp/src/github.com/cloudflare/cfssl/cmd/cfssl

	$(eval APP_DIR_LIST=$(shell GOPATH=$(GOPATH) go list ./temp/src/github.com/trustedanalytics-ng/$(PROJECT_NAME)/... | grep -v /vendor/))
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -tags netgo $(APP_DIR_LIST)

	mv ./cfssl ./application/cfssl
	mv ./$(PROJECT_NAME) ./application/$(PROJECT_NAME)
	rm -Rf ./temp

kubernetes_deploy: docker_build
	kubectl create -f configmap.yaml
	kubectl create -f service.yaml
	kubectl create -f deployment.yaml

kubernetes_update: docker_build
	kubectl delete -f deployment.yaml
	kubectl create -f deployment.yaml

docker_build: build_anywhere
	docker build -t $(PROJECT_NAME) .

push_docker: docker_build
	docker tag $(PROJECT_NAME) $(REPOSITORY_URL)/$(PROJECT_NAME):latest
	docker push $(REPOSITORY_URL)/$(PROJECT_NAME):latest

mock_update: 
	$(GOPATH)/bin/mockgen -source=engine/processor.go -package=engine -destination=engine/processor_mock.go
	./add_license.sh

tests: verify_gopath
	$(eval APP_DIR_LIST=$(shell go list ./... | grep -v /vendor/ | grep -v /temp))
	go test --cover $(APP_DIR_LIST)

