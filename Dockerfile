FROM alpine:3.6
LABEL description="espipe, elasticsearch, pipeline"
MAINTAINER Guillaume Simonneau
COPY ./ /tmp/app
RUN chmod +x /tmp/app/install_amd64.sh \
&&  sh /tmp/app/install_amd64.sh

FROM alpine:3.6
ENTRYPOINT ["/entrypoint.sh"]
CMD ["espipe"]
