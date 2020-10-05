FROM alpine:3.4

RUN apk add --update ca-certificates && rm -rf /var/cache/apk/*

COPY opssight-cloud-auth .

COPY blackduck-connector-2.2.5-1.tgz /opt/blackduck-connector-2.2.5-1.tgz 

CMD .opssight-cloud-auth
