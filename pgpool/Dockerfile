FROM golang:alpine
COPY src/main.go /build/
RUN cd /build && go build -o configure-pgpool2 -ldflags "-s -w" .

FROM alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories &&\
    apk --no-cache add pgpool

COPY --from=0 /build/configure-pgpool2 /usr/bin/pgpool2-starter
RUN chmod +x /usr/bin/pgpool2-starter \
    && mkdir /etc/pgpool2 /var/run/pgpool /var/log/pgpool /var/run/postgresql /var/log/postgresql/  /usr/share/pgpool2/ \
    && chown postgres /etc/pgpool2 /var/run/pgpool /var/log/pgpool /var/run/postgresql /var/log/postgresql

COPY conf/pcp.conf.template  /usr/share/pgpool2/
COPY conf/pgpool.conf.template /usr/share/pgpool2/

EXPOSE 9999 9898

CMD "pgpool2-starter"
