FROM golang:1.10.4-alpine3.7 as build
COPY ./ /tmp/app
# save files
RUN mkdir /default \
&&  cp /tmp/app/config.json /default/config.json \
&& mv /tmp/app/entrypoint.sh /entrypoint.sh \
&& chmod +x /entrypoint.sh
# installing required packages
RUN apk update && apk add ca-certificates wget tar git
# compilation
RUN mkdir -p /usr/local/go/src/github.com/khezen/ \
&&  mv /tmp/app /usr/local/go/src/github.com/khezen/espipe \
&&  go build -o /bin/espipe github.com/khezen/espipe

FROM alpine:3.7
COPY --from=build /default/config.json /default/config.json
COPY --from=build /entrypoint.sh /entrypoint.sh
COPY --from=build /bin/espipe /bin/espipe
RUN apk add --no-cache ca-certificates \
&&  mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
ENTRYPOINT ["/entrypoint.sh"]
CMD ["espipe"]
