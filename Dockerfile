FROM golang:1.13 AS build

WORKDIR /ytdlweb
COPY . .

ENV CGO_ENABLED=0
RUN go generate && go build

FROM alpine:3.11

RUN apk add --no-cache --allow-untrusted \
    --repository https://pkgs.alpinelinux.org/package/edge/community/ \
    youtube-dl

WORKDIR /ytdlweb
COPY --from=build /ytdlweb/ytdlweb .

CMD ["/ytdlweb/ytdlweb"]
