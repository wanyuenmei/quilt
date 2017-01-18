FROM nginx:1.10

ADD default.tmpl /etc/nginx/conf.d/default.tmpl

CMD ["/bin/bash", "-c", "envsubst < /etc/nginx/conf.d/default.tmpl > /etc/nginx/conf.d/default.conf \
    && nginx -g 'daemon off;'"]
