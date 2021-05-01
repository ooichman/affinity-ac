FROM golang:alpine as builder

RUN mkdir -p /opt/app-root/src/affinity-ac/
WORKDIR /opt/app-root/src/affinity-ac/
ENV GOPATH="/opt/app-root/"
ENV PATH="${PATH}:/opt/app-root/src/go/bin"

COPY src/affinity-ac .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o affinity-ac

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /opt/app-root/src/affinity-ac/affinity-ac /usr/bin/

USER nobody

EXPOSE 8080 8443

CMD ["/usr/bin/affinity-ac"]