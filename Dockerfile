ARG ARCH=amd64
FROM --platform=linux/${ARCH} gcr.io/distroless/static-debian11

ADD rqd /usr/local/bin
ADD rqctl /usr/local/bin

WORKDIR /var/rqd/
WORKDIR /var/lib/rqd/

EXPOSE 5290 5230

CMD ["/usr/local/bin/rqd"]