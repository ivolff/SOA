package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"flag"
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
		fmt.Println("before serilize:\n", t)
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
		fmt.Printf("Serialized: \n%X\n size: %d\n", buff.Bytes(), size)
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
		fmt.Println("Deserialized: \n", out, "\n")

	}

	return size, elapsed, elapsed1
}

func GobSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("before serilize:\n", t)
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
		fmt.Printf("Serialized: \n%X\n size: %d\n", buff.Bytes(), size)
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
		fmt.Println("Deserialized: \n", out, "\n")

	}

	return size, elapsed, elapsed1
}

func JsonSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("before serilize:\n", t)
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
		fmt.Printf("Serialized: \n%X\n size: %d\n", jsbytes, size)
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
		fmt.Println("Deserialized: \n", result, "\n")
	}

	return size, elapsed, elapsed1
}

func ProtoSerialise(t models.Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("before serilize:\n", t)
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
		fmt.Printf("Serialized: \n%X\n size: %d\n", out, size)
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
		fmt.Println("Deserialized: \n", result, "\n")
	}

	return size, elapsed, elapsed1
}

func AvroSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("before serilize:\n", t)
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
		fmt.Printf("Serialized: \n%X\n size: %d\n", out, size)
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
		fmt.Println("Deserialized: \n", result, "\n")
	}

	return size, elapsed, elapsed1
}

func YAMLSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("before serilize:\n", t)
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
		fmt.Printf("Serialized: \n%X\n size: %d\n", yamlbytes, size)
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
		fmt.Println("Deserialized: \n", result, "\n")
	}

	return size, elapsed, elapsed1
}

func MSGpSerialise(t Test, silence bool) (int, int, int) {
	if !silence {
		fmt.Println("before serilize:\n", t)
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
		fmt.Printf("Serialized: \n%X\n size: %d\n", msg, size)
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
		fmt.Println("Deserialized: \n", result, "\n")
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

	nruns := flag.Int("runs", 1, "number of runs for every serializer(int)")
	s := flag.Bool("s", true, "show every serialization run report(bool)")
	si := flag.Bool("si", true, "show detail report about every serialization run(bool)")
	flag.Parse()

	num_runs := *nruns
	silence := !*s
	silenceInside := !*si

	GOBsize, GOBIntime, GOBOuttime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		if !silence {
			fmt.Printf("------------Gob------RUN #%d---------------- \n", i)
		}
		F, S, T := GobSerialise(t, silenceInside)
		GOBsize = F
		if !silence {
			fmt.Printf("GOB size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", F, S, T, S+T)
		}
		GOBIntime += S
		GOBOuttime += T
	}

	GOBIntime /= num_runs
	GOBOuttime /= num_runs

	fmt.Printf("\nAvarage: ")
	fmt.Printf("GOB size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", GOBsize, GOBIntime, GOBOuttime, GOBIntime+GOBOuttime)

	println("\n\n")

	XMLsize, XMLIntime, XMLOuttime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		if !silence {
			fmt.Printf("------------XML------RUN #%d---------------- \n", i)
		}

		F, S, T := XMLSerialise(t, silenceInside)
		XMLsize = F
		if !silence {
			fmt.Printf("XML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", F, S, T, S+T)
		}
		XMLIntime += S
		XMLOuttime += T
	}

	XMLIntime /= num_runs
	XMLOuttime /= num_runs

	fmt.Printf("\nAvarage: ")
	fmt.Printf("XML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", XMLsize, XMLIntime, XMLOuttime, XMLIntime+XMLOuttime)

	println("\n\n")

	Jsonsize, JsonIntime, JsonOuttime := 0, 0, 0

	for i := 0; i < num_runs; i++ {

		if !silence {
			fmt.Printf("------------JSON------RUN #%d---------------- \n", i)
		}
		F, S, T := JsonSerialise(t, silenceInside)
		Jsonsize = F
		if !silence {
			fmt.Printf("Json size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", F, S, T, S+T)
		}
		JsonIntime += S
		JsonOuttime += T
	}

	JsonIntime /= num_runs
	JsonOuttime /= num_runs

	fmt.Printf("\nAvarage: ")
	fmt.Printf("JSON size:  %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", Jsonsize, JsonIntime, JsonOuttime, JsonIntime+JsonOuttime)
	println("\n\n")

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
		if !silence {
			fmt.Printf("------------Proto------RUN #%d---------------- \n", i)
		}
		F, S, T := ProtoSerialise(tproto, silenceInside)
		protoSize = F
		if !silence {
			fmt.Printf("Proto size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", F, S, T, S+T)
		}
		protoInTime += S
		protoOutTime += T
	}

	protoInTime /= num_runs
	protoOutTime /= num_runs

	fmt.Printf("\nAvarage: ")
	fmt.Printf("Proto size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", protoSize, protoInTime, protoOutTime, protoInTime+protoOutTime)
	println("\n\n")

	AvroSize, AvroInTime, AvroOutTime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		if !silence {
			fmt.Printf("------------Avro------RUN #%d---------------- \n", i)
		}
		F, S, T := AvroSerialise(t, silenceInside)
		AvroSize = F
		if !silence {
			fmt.Printf("Avro size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", F, S, T, S+T)
		}
		AvroInTime += S
		AvroOutTime += T
	}

	AvroInTime /= num_runs
	AvroOutTime /= num_runs

	fmt.Printf("\nAvarage: ")
	fmt.Printf("Avro size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", AvroSize, AvroInTime, AvroOutTime, AvroInTime+AvroOutTime)
	println("\n\n")

	YAMLSize, YAMLInTime, YAMLOutTime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		if !silence {
			fmt.Printf("------------YAML------RUN #%d---------------- \n", i)
		}
		F, S, T := YAMLSerialise(t, silenceInside)
		YAMLSize = F
		if !silence {
			fmt.Printf("YAML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", F, S, T, S+T)
		}
		YAMLInTime += S
		YAMLOutTime += T
	}

	YAMLInTime /= num_runs
	YAMLOutTime /= num_runs

	fmt.Printf("\nAvarage: ")
	fmt.Printf("YAML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", YAMLSize, YAMLInTime, YAMLOutTime, YAMLInTime+YAMLOutTime)
	println("\n\n")

	MSGpSize, MSGpInTime, MSGpOutTime := 0, 0, 0

	for i := 0; i < num_runs; i++ {
		if !silence {
			fmt.Printf("------------MSGPack------RUN #%d---------------- \n", i)
		}
		F, S, T := MSGpSerialise(t, silenceInside)
		MSGpSize = F
		if !silence {
			fmt.Printf("MSGPack size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", F, S, T, S+T)
		}
		MSGpInTime += S
		MSGpOutTime += T
	}

	MSGpInTime /= num_runs
	MSGpOutTime /= num_runs

	fmt.Printf("\nAvarage: ")
	fmt.Printf("MSGp size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", MSGpSize, MSGpInTime, MSGpOutTime, MSGpInTime+MSGpOutTime)
}
