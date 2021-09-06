FROM alpine:3.6 as certs
RUN apk add -U --no-cache ca-certificates

FROM scratch

ENV PAGES_PORT=3000
ENV PAGES_BIND=0.0.0.0
ENV PAGES_ROOT=/pages/default

WORKDIR /pages/conf

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY pages /usr/bin/pages
COPY res/docker/Pagesfile /pages/conf/Pagesfile
COPY res/pages/ /pages/

EXPOSE 3000/tcp
VOLUME ["/pages/conf", "/pages/default", "/pages/domains"]

ENTRYPOINT ["/usr/bin/pages"]
CMD ["start"]
