REPO = quilt/mongo

all: build-scripts build-image

build-scripts:
	cd scripts \
		&& GOOS=linux GOARCH=amd64 go build -o run

build-image: build-scripts
	@docker build -t $(REPO) .

push-image: build-image
	@docker push $(REPO)
