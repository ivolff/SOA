FROM golang:1.18

RUN mkdir /root/app/
WORKDIR /root/app/

COPY . .

RUN go mod download

RUN go build -o /worker

EXPOSE 8080

CMD ["bash", "-c", "while ! curl -s rabbitmq:15672 > /dev/null; do echo waiting for rabbitmq; sleep 3; done; /worker $ID"]
