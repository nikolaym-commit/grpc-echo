FROM public.ecr.aws/docker/library/golang:1.23-alpine3.22 AS builder

ENV CGO_ENABLED=0

WORKDIR /srv

RUN apk add --no-cache --update git bash curl tzdata && \
    cp /usr/share/zoneinfo/Asia/Almaty /etc/localtime && \
    rm -rf /var/cache/apk/*

COPY ./echopb/ /srv/echopb
COPY ./main.go /srv/main.go
COPY ./pkg/ /srv/pkg

COPY ./go.mod /srv/go.mod
COPY ./go.sum /srv/go.sum

COPY ./.git/ /srv/.git

RUN \
    export version="$(git describe --tags --long)" && \
    echo "version: $version" && \
    go build -o /go/build/grpc-echo -ldflags "-X 'main.version=${version}' -s -w" /srv/main.go

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/nikolaym-commit/grpc-echo"
LABEL maintainer="NikolayM <nikolaym@comm-it.com>"

COPY --from=builder /go/build/grpc-echo /usr/bin/grpc-echo
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

EXPOSE 8080
ENTRYPOINT ["/usr/bin/grpc-echo", "--addr", ":8080"]
