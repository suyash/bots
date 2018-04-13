FROM node:carbon-alpine as js_builder
WORKDIR /usr/src/app
COPY ./static/js/package.json ./package.json
COPY ./static/js/package-lock.json ./package-lock.json
RUN npm install
COPY ./static/js .
RUN npm run build

FROM golang:alpine as builder
WORKDIR /go/src/app
RUN apk add --no-cache git
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY tmpl tmpl
COPY --from=builder /go/bin/app /app
COPY --from=js_builder /usr/src/app/lib /static/js/lib
ENTRYPOINT ./app
LABEL Name=bots Version=0.0.1
EXPOSE 8080
