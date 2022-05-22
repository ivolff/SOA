FROM golang:1.18

WORKDIR /usr/local/app

COPY . .

RUN go build .

CMD ["bash", "-c", "while ! curl -s mongodb:27017 > /dev/null; do echo waiting for mongodb; sleep 3; done; ./graphql-mafia"]
