FROM golang:alpine as build
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk update
ENV GO111MODULE=on \
	GOPROXY="https://goproxy.cn,direct"
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o sso

FROM alpine
WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
COPY --from=build /app/sso ./sso
ENV SERVER=0.0.0.0:80
EXPOSE 80
CMD [ "./sso" ]
