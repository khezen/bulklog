FROM golang:1.11.2-alpine3.8 as build
# install additional tools
RUN apk add --no-cache git openssh-client musl-dev gcc curl
# copy files
COPY ./ /tmp/app
# save files
RUN mkdir /default \
&&  cp /tmp/app/config.json /default/config.json \
&& mv /tmp/app/entrypoint.sh /entrypoint.sh \
&& chmod +x /entrypoint.sh
# compilation
RUN mkdir -p /usr/local/go/src/github.com/khezen/ \
&&  mv /tmp/app /usr/local/go/src/github.com/khezen/espipe \
&&  go build -o /bin/espipe github.com/khezen/espipe

FROM alpine:3.8
COPY --from=build /default/config.json /default/config.json
COPY --from=build /entrypoint.sh /entrypoint.sh
COPY --from=build /bin/espipe /bin/espipe
RUN apk add --no-cache ca-certificates
ENTRYPOINT ["/entrypoint.sh"]
CMD ["espipe"]
