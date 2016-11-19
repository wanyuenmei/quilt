export GO15VENDOREXPERIMENT=1
PACKAGES=$(shell govendor list -no-status +local)
NOVENDOR=$(shell find . -path ./specs/**/*/vendor -prune -o -path ./vendor -prune -o -name '*.go' -print)
LINE_LENGTH_EXCLUDE=./constants/awsConstants.go \
		    ./constants/gceConstants.go \
		    ./cluster/provider/cloud_config.go \
		    ./minion/network/link_test.go \
		    ./minion/pb/pb.pb.go \
		    ./api/pb/pb.pb.go \
		    ./stitch/bindings.js.go \
		    ./cluster/provider/mocks/EC2Client.go

REPO = quilt
DOCKER = docker
SHELL := /bin/bash

all:
	cd -P . && go build .

install:
	cd -P . && go install .

check: format-check
	govendor test +local

clean:
	govendor clean -x +local
	rm -f *.cov.coverprofile cluster/*.cov.coverprofile minion/*.cov.coverprofile specs/*.cov.coverprofile
	rm -f *.cov.html cluster/*.cov.html minion/*.cov.html specs/*.cov.html

COV_SKIP= /minion/pb /minion/pprofile /api/pb /constants /scripts /quilt-tester \
		  /quilt-tester/tests/100-basic /quilt-tester/tests/100-basic/check_docker.go \
		  /quilt-tester/tests/100-basic/check_logs.go \
		  /quilt-tester/tests/spark /quilt-tester/tests/spark/check_spark.go \
		  /quiltctl/testutils /cluster/provider/mocks

COV_PKG = $(subst github.com/NetSys/quilt,,$(PACKAGES))
coverage: $(addsuffix .cov, $(filter-out $(COV_SKIP), $(COV_PKG)))
	gover

%.cov:
	go test -coverprofile=.$@.coverprofile .$*
	go tool cover -html=.$@.coverprofile -o .$@.html

format: scripts/format
	gofmt -w -s $(NOVENDOR)
	scripts/format $(filter-out $(LINE_LENGTH_EXCLUDE),$(NOVENDOR))

scripts/format: scripts/format.go
	cd scripts && go build format.go

format-check:
	RESULT=`gofmt -s -l $(NOVENDOR)` && \
	if [[ -n "$$RESULT"  ]] ; then \
	    echo $$RESULT && \
	    exit 1 ; \
	fi

lint: format
	cd -P . && govendor vet +local
	for package in $(PACKAGES) ; do \
		if [[ $$package != *minion/pb* && $$package != *api/pb* ]] ; then \
			golint -min_confidence .25 -set_exit_status $$package || exit 1 ; \
		fi \
	done
	ineffassign .

generate:
	govendor generate +local

providers:
	python3 scripts/gce-descriptions > provider/gceConstants.go

go-get:
	go get -v -u \
	    github.com/golang/protobuf/{proto,protoc-gen-go} \
	    github.com/gordonklaus/ineffassign \
	    github.com/davecgh/go-spew/spew \
	    github.com/golang/lint/golint \
	    github.com/mattn/goveralls \
	    github.com/modocache/gover \
	    github.com/kardianos/govendor \
	    github.com/vektra/mockery

tests:
	cd -P quilt-tester && \
	for suite in tests/* ; do \
		for f in $$suite/* ; do \
			if [ $${f: -3} == ".go" ] ; then \
				CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $${f%???} $$f ; \
			fi \
		done \
	done

docker-build-quilt:
	cd -P . && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build . \
	    && ${DOCKER} build -t ${REPO}/quilt .

docker-build-tester: tests
	cd -P quilt-tester \
	    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	       go build -o bin/quilt-tester . \
	    && ${DOCKER} build -t ${REPO}/tester .

docker-build-ovs:
	cd -P ovs && docker build -t ${REPO}/ovs .

# Include all .mk files so you can have your own local configurations
include $(wildcard *.mk)
