FROM golang:1.13

WORKDIR /serialization-test

COPY main.go ./
COPY models/schema.avsc ./models/
COPY models/test.pb.go ./models/

RUN mkdir files
RUN go get -v "github.com/hamba/avro"
RUN go get -v "github.com/vmihailenco/msgpack"
RUN go get -v "google.golang.org/protobuf/proto"
RUN go get -v "google.golang.org/protobuf/reflect/protoreflect"
RUN go get -v "google.golang.org/protobuf/runtime/protoimpl"
RUN go get -v "gopkg.in/yaml.v2"


RUN go build main.go
ENTRYPOINT ./main 