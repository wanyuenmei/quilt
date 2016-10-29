FROM node:6.8

COPY scripts/run /
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

RUN git clone https://github.com/tejasmanohar/node-todo.git .
RUN npm install

CMD ["/run"]
