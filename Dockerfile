FROM golang:1.10-alpine
#RUN apk add --no-cache curl git
#RUN curl https://glide.sh/get | sh
WORKDIR /go/src/github.com/noodlensk/go-greenhouse-gpio/
COPY . .
#RUN glide i
RUN go test -v ./...
RUN GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -o go-greenhouse-gpio

FROM alpine:3.7
RUN apk --no-cache add ca-certificates tzdata
RUN cp /usr/share/zoneinfo/Asia/Novosibirsk /etc/localtime
COPY --from=0 /go/src/github.com/noodlensk/go-greenhouse-gpio/go-greenhouse-gpio /go-greenhouse-gpio
#COPY --from=0 /go/src/github.com/noodlensk/go-greenhouse-gpio/config.yaml /config.yaml

ENV "TELEGRAM_BOT_TOKEN" "MyAwesomeBotToken"

CMD [ "/go-greenhouse-gpio" ]