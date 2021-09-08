FROM alpine:3.6 as support

RUN apk add -U --no-cache ca-certificates mailcap

FROM scratch

ENV PAGES_PORT=3000
ENV PAGES_BIND=0.0.0.0
ENV PAGES_ROOT=/pages/default

WORKDIR /pages/conf

COPY --from=support /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=support /etc/mime.types /etc/
COPY pages /usr/bin/pages
COPY res/docker/Pagesfile /pages/conf/Pagesfile
COPY res/pages/ /pages/

EXPOSE 3000/tcp
VOLUME ["/pages/conf", "/pages/default", "/pages/domains"]

ENTRYPOINT ["/usr/bin/pages"]
CMD ["start"]
