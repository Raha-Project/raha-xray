FROM golang:1.21-alpine AS builder
WORKDIR /app
ARG TARGETARCH 
RUN apk --no-cache --update add build-base gcc wget unzip
COPY . .
RUN env CGO_ENABLED=1 go build -o build/raha-xray main.go

FROM alpine
LABEL org.opencontainers.image.authors="alireza7@gmail.com"
ENV TZ=Asia/Tehran
WORKDIR /app

RUN apk add  --no-cache --update ca-certificates tzdata

COPY --from=builder  /app/build/ /app/
VOLUME [ "raha-xray" ]
CMD [ "./raha-xray" ]