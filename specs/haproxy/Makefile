REPO = quilt/haproxy

all: build-image

build-scripts:
	cd -P scripts \
		&& GOOS=linux GOARCH=amd64 go build -o start

build-image: build-scripts
	docker build -t $(REPO) .

push-image: build-image
	docker push $(REPO)
