FROM amazonlinux AS build
RUN yum -y update && yum -y install tar gzip
RUN curl -o go1.14.3.linux-amd64.tar.gz https://dl.google.com/go/go1.14.3.linux-amd64.tar.gz -s
RUN tar -C /usr/local -xzf go1.14.3.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /src/aws-secrets-manager
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . ./
RUN go build -o /app -v ./cmd/aws-secrets-manager

FROM amazonlinux:latest
RUN yum -y update && yum install -y ca-certificates && rm -rf /var/cache/yum/*
COPY --from=build /app /.
COPY sc-ca-bundle.crt /etc/ssl/certs/
COPY sc-ca-bundle.crt /etc/pki/ca-trust/source/anchors
ENV http_proxy=gateway.zscloud.net:443
ENV https_proxy=gateway.zscloud.net:443
ENTRYPOINT ["/app"]
