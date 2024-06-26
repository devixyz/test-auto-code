FROM golang:1.20-alpine as build-stage


WORKDIR /work

RUN go env -w GOPROXY=https://goproxy.cn,direct && go env -w CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o Einstein main.go



FROM alpine

COPY --from=build-stage /work/Einstein  /Einstein
COPY --from=build-stage /work/config  /config
COPY --from=build-stage /work/common  /common



EXPOSE 8012
RUN  chmod +x /Einstein
CMD ["/Einstein","server","-c", "/config/settings-prod.yml"]