package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"time"

	models "./models"

	"github.com/hamba/avro"
	"github.com/vmihailenco/msgpack"
	"google.golang.org/protobuf/proto"
	yaml "gopkg.in/yaml.v2"
)

type TestStruct struct {
	Some  string `json:"some" avro:"some"`
	Other string `json:"other" avro:"other"`
}

type Test struct {
	ID               int          `json:"id" avro:"Id"`
	Name             string       `json:"name" avro:"Name"`
	SomeNumericArray []int32      `json:"somenumericarray" avro:"Somenumericarray"`
	Tests            []TestStruct `json:"tests" avro:"Tests"`
}

func XMLSerialise(t Test, silence bool) (int, int, int) {

	if !silence {
		fmt.Println("XML")
	}
	var buff bytes.Buffer

	start := time.Now().Nanosecond()

	encoder := xml.NewEncoder(&buff)
	errors := encoder.Encode(t)

	end := time.Now().Nanosecond()
	elapsed := end - start

	if errors != nil {
		fmt.Println("ERRS", errors)
	}
	ioutil.WriteFile("files/XML", buff.Bytes(), 777)

	size := len(buff.Bytes())

	if !silence {
		fmt.Printf("%X\n size: %d\n", buff.Bytes(), size)
	}

	start = time.Now().Nanosecond()

	var out Test
	file, errors := ioutil.ReadFile("files/XML")
	decoder := xml.NewDecoder(bytes.NewBuffer(file))
	errors = decoder.Decode(&out)

	end = time.Now().Nanosecond()
	elapsed1 := end - start

	if errors != nil {
		fmt.Println("ERRS", errors)
	}

	if !silence {
		fmt.Println(out)
	}

	return size, elapsed, elapsed1
}

func GobSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("GOB")
	}

	start := time.Now().Nanosecond()

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	errors := encoder.Encode(t)

	end := time.Now().Nanosecond()
	elapsed := end - start

	if errors != nil {
		fmt.Println("ERRS", errors)
	}

	ioutil.WriteFile("files/Gob", buff.Bytes(), 777)

	size := len(buff.Bytes())

	if !silence {
		fmt.Printf("%X\n size: %d\n", buff.Bytes(), size)
	}

	start = time.Now().Nanosecond()

	file, errors := ioutil.ReadFile("files/Gob")
	decoder := gob.NewDecoder(bytes.NewBuffer(file))
	var out Test
	errors = decoder.Decode(&out)

	end = time.Now().Nanosecond()
	elapsed1 := end - start

	if errors != nil {
		fmt.Println("ERRS", errors)
	}

	if !silence {
		fmt.Println(out)
	}

	return size, elapsed, elapsed1
}

func JsonSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("JSON")
		fmt.Println(t)
	}

	start := time.Now().Nanosecond()

	jsbytes, errors := json.Marshal(t)

	end := time.Now().Nanosecond()
	elapsed := end - start

	if errors != nil {
		fmt.Println("encoding error", errors)
	}

	ioutil.WriteFile("files/Json", jsbytes, 777)
	size := len(jsbytes)

	if !silence {
		fmt.Printf("%X\n size: %d\n", jsbytes, size)
	}

	start = time.Now().Nanosecond()

	file, errors := ioutil.ReadFile("files/Json")
	result := Test{}
	errors = json.Unmarshal(file, &result)

	end = time.Now().Nanosecond()
	elapsed1 := end - start

	if errors != nil {
		fmt.Println("decoding error", errors)
	}

	if !silence {
		fmt.Println(result)
	}

	return size, elapsed, elapsed1
}

func ProtoSerialise(t models.Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("Proto")
		fmt.Println(t)
	}

	start := time.Now().Nanosecond()

	out, errors := proto.Marshal(&models.Test{
		ID:               int32(t.ID),
		Name:             t.Name,
		Tests:            t.Tests,
		SomeNumericArray: t.SomeNumericArray,
	})

	end := time.Now().Nanosecond()
	elapsed := end - start

	if errors != nil {
		fmt.Println("encoding error", errors)
	}

	ioutil.WriteFile("files/Proto", out, 777)
	size := len(out)

	if !silence {
		fmt.Printf("%X\n size: %d\n", out, size)
	}

	start = time.Now().Nanosecond()

	file, errors := ioutil.ReadFile("files/Proto")
	result := &models.Test{}
	errors = proto.Unmarshal(file, result)

	end = time.Now().Nanosecond()
	elapsed1 := end - start

	if errors != nil {
		fmt.Println("decoding error", errors)
	}

	if !silence {
		fmt.Println(result)
	}

	return size, elapsed, elapsed1
}

func AvroSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("Avro")
		fmt.Println(t)
	}

	start := time.Now().Nanosecond()

	schemaStr, errors := ioutil.ReadFile("./models/schema.avsc")

	if errors != nil {
		fmt.Println("schema reading error", errors)
	}

	schema, errors := avro.Parse(string(schemaStr))

	if errors != nil {
		fmt.Println("schema parsing error", errors)
	}

	out, errors := avro.Marshal(schema, t)

	if errors != nil {
		fmt.Println("Serializing error", errors)
	}

	end := time.Now().Nanosecond()
	elapsed := end - start

	ioutil.WriteFile("files/Avro", out, 777)

	size := len(out)

	if !silence {
		fmt.Printf("%X\n size: %d\n", out, size)
	}

	start = time.Now().Nanosecond()

	file, errors := ioutil.ReadFile("files/Avro")
	result := &Test{}
	errors = avro.Unmarshal(schema, file, result)
	if errors != nil {
		fmt.Println("decoding error", errors)
	}

	end = time.Now().Nanosecond()
	elapsed1 := end - start

	if !silence {
		fmt.Println(result)
	}

	return size, elapsed, elapsed1
}

func YAMLSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("YAML")
		fmt.Println(t)
	}

	start := time.Now().Nanosecond()

	yamlbytes, errors := yaml.Marshal(t)

	if errors != nil {
		fmt.Println("encoding error", errors)
	}

	end := time.Now().Nanosecond()
	elapsed := end - start

	ioutil.WriteFile("files/YAML", yamlbytes, 777)

	size := len(yamlbytes)

	if !silence {
		fmt.Printf("%X\n size: %d\n", yamlbytes, size)
	}

	start = time.Now().Nanosecond()

	file, errors := ioutil.ReadFile("files/YAML")
	result := Test{}
	errors = yaml.Unmarshal(file, &result)

	end = time.Now().Nanosecond()
	elapsed1 := end - start

	if errors != nil {
		fmt.Println("decoding error", errors)
	}

	if !silence {
		fmt.Println(result)
	}

	return size, elapsed, elapsed1
}

func MSGpSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("YAML")
		fmt.Println(t)
	}

	start := time.Now().Nanosecond()

	msg, errors := msgpack.Marshal(t)

	if errors != nil {
		fmt.Println("encoding error", errors)
	}

	end := time.Now().Nanosecond()
	elapsed := end - start

	ioutil.WriteFile("files/MSG", msg, 777)

	size := len(msg)

	if !silence {
		fmt.Printf("%X\n size: %d\n", msg, size)
	}

	start = time.Now().Nanosecond()

	file, errors := ioutil.ReadFile("files/MSG")
	result := Test{}
	errors = msgpack.Unmarshal(file, &result)

	end = time.Now().Nanosecond()
	elapsed1 := end - start
	if errors != nil {
		fmt.Println("decoding error", errors)
	}

	if !silence {
		fmt.Println(result)
	}

	return size, elapsed, elapsed1
}

func main() {
	t := Test{
		ID:               1,
		Name:             "Test",
		SomeNumericArray: []int32{1, 2, 3},
		Tests: []TestStruct{
			{Some: "Slava", Other: "Bebrow"},
			{Some: "Slava2", Other: "Bebrow3"},
		},
	}

	num_runs := 1
	silence := true

	GOBsize, GOBIntime, GOBOuttime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		F, S, T := GobSerialise(t, false)
		GOBsize = F
		if !silence {
			fmt.Printf("RUN #%d \nGOB size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", i, F, S, T, S+T)
		}
		GOBIntime += S
		GOBOuttime += T
	}

	GOBIntime /= num_runs
	GOBOuttime /= num_runs
	fmt.Printf("GOB size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", GOBsize, GOBIntime, GOBOuttime, GOBIntime+GOBOuttime)

	XMLsize, XMLIntime, XMLOuttime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		F, S, T := XMLSerialise(t, false)
		XMLsize = F
		if !silence {
			fmt.Printf("RUN #%d \nXML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", i, F, S, T, S+T)
		}
		XMLIntime += S
		XMLOuttime += T
	}

	XMLIntime /= num_runs
	XMLOuttime /= num_runs

	fmt.Printf("XML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", XMLsize, XMLIntime, XMLOuttime, XMLIntime+XMLOuttime)

	Jsonsize, JsonIntime, JsonOuttime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		F, S, T := JsonSerialise(t, false)
		Jsonsize = F
		if !silence {
			fmt.Printf("RUN #%d \nJson size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", i, F, S, T, S+T)
		}
		JsonIntime += S
		JsonOuttime += T
	}

	JsonIntime /= num_runs
	JsonOuttime /= num_runs

	fmt.Printf("JSON size:  %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", Jsonsize, JsonIntime, JsonOuttime, JsonIntime+JsonOuttime)

	tproto := models.Test{
		ID:               1,
		Name:             "Test",
		SomeNumericArray: []int32{1, 2, 3},
		Tests: []*models.TestStruct{
			{Some: "Slava", Other: "Bebrow"},
			{Some: "Slava2", Other: "Bebrow3"},
		},
	}

	protoSize, protoInTime, protoOutTime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		F, S, T := ProtoSerialise(tproto, false)
		protoSize = F
		if !silence {
			fmt.Printf("RUN #%d \nProto size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", i, F, S, T, S+T)
		}
		protoInTime += S
		protoOutTime += T
	}

	protoInTime /= num_runs
	protoOutTime /= num_runs

	fmt.Printf("Proto size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", protoSize, protoInTime, protoOutTime, protoInTime+protoOutTime)

	AvroSize, AvroInTime, AvroOutTime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		F, S, T := AvroSerialise(t, false)
		AvroSize = F
		if !silence {
			fmt.Printf("RUN #%d \nAvro size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", i, F, S, T, S+T)
		}
		AvroInTime += S
		AvroOutTime += T
	}

	AvroInTime /= num_runs
	AvroOutTime /= num_runs

	fmt.Printf("Avro size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", AvroSize, AvroInTime, AvroOutTime, AvroInTime+AvroOutTime)

	YAMLSize, YAMLInTime, YAMLOutTime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		F, S, T := YAMLSerialise(t, false)
		YAMLSize = F
		if !silence {
			fmt.Printf("RUN #%d \nYAML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", i, F, S, T, S+T)
		}
		YAMLInTime += S
		YAMLOutTime += T
	}

	YAMLInTime /= num_runs
	YAMLOutTime /= num_runs

	fmt.Printf("YAML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", YAMLSize, YAMLInTime, YAMLOutTime, YAMLInTime+YAMLOutTime)

	MSGpSize, MSGpInTime, MSGpOutTime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		F, S, T := MSGpSerialise(t, false)
		MSGpSize = F
		if !silence {
			fmt.Printf("RUN #%d \nYAML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", i, F, S, T, S+T)
		}
		MSGpInTime += S
		MSGpOutTime += T
	}

	MSGpInTime /= num_runs
	MSGpOutTime /= num_runs
	fmt.Printf("MSGp size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", MSGpSize, MSGpInTime, MSGpOutTime, MSGpInTime+MSGpOutTime)
}
