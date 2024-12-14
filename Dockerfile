FROM docker.io/golang:1.23-alpine AS build
LABEL MAINTAINER github.com/arizon-dread

WORKDIR /usr/local/go/src/github.com/arizon-dread/plats
RUN ls -la
COPY . .

ENV GOARCH=amd64
ENV GOOS=linux

RUN apk update && apk add --no-cache git 
RUN go build -v -o /usr/local/bin/ ./... 

FROM docker.io/alpine:3.21
WORKDIR /go/bin
COPY --from=build /usr/local/bin/plats/ /go/bin/

ENV environment=development
ENV path=/go/bin/config

EXPOSE 8080
ENTRYPOINT ["./plats"]

