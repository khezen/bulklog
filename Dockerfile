FROM golang:1.8.3-alpine3.6 as build
COPY ./ /tmp/app
RUN chmod +x /tmp/app/install_amd64.sh \
&&  sh /tmp/app/install_amd64.sh

FROM alpine:3.6
COPY --from=build /default/config.json /default/config.json
COPY --from=build /entrypoint.sh /entrypoint.sh
COPY --from=build /bin/espipe /bin/espipe
RUN apk add --no-cache ca-certificates \
&&  mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
ENTRYPOINT ["/entrypoint.sh"]
CMD ["espipe"]
