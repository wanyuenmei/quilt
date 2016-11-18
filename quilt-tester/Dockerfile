From quilt/quilt
Maintainer Ethan J. Jackson

RUN VER=1.6.2 \
&& export GOROOT=/tmp/build/go GOPATH=/tmp/build/gowork \
&& export PATH=$PATH:$GOROOT/bin \
&& apt-get update \
&& apt-get install -y --no-install-recommends \
    wget bash git python curl openssh-client \
&& mkdir /tmp/build && cd /tmp/build \
&& wget https://storage.googleapis.com/golang/go$VER.linux-amd64.tar.gz \
&& gunzip -c go$VER.linux-amd64.tar.gz | tar x \
&& go get -u github.com/NetSys/quilt \
&& cd $GOPATH/src/github.com/NetSys/quilt/quilt-tester \
&& go build -o /bin/quilt-tester . \
&& mkdir /tests \
&& for suite in tests/* ; do \
	for f in $suite/* ; do \
		extension=$(echo "$f" | awk -F'.' '{print $NF}') ; \
		if [ $extension = "go" ] ; then \
			go build -v -o ${f%???} $f ; \
		fi; \
	done; \
	cp -r $suite /tests ; \
done \
&& rm -rf /tmp/build
Copy config/id_rsa /root/.ssh/id_rsa
RUN chmod 0600 /root/.ssh/id_rsa

Entrypoint ["quilt-tester"]
