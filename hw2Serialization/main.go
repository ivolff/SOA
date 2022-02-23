package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"
	"unicode"

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
	ID               int32        `json:"id" avro:"Id"`
	Name             string       `json:"name" avro:"Name"`
	SomeNumericArray []int32      `json:"somenumericarray" avro:"Somenumericarray"`
	SomeFloatArray   []float32    `json:"somefloatarray" avro:"Somefloatarray"`
	Tests            []TestStruct `json:"tests" avro:"Tests"`
}

func randString() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789абвгдеёжзийклмнопрстуфхцчшщъыьэюяАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ")
	s := make([]rune, rand.Intn(32))
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
		for i == 0 && unicode.IsDigit(s[i]) {
			s[i] = letters[rand.Intn(len(letters))]
		}
	}
	return string(s)
}

func generateTest(simple bool) (Test, models.Test) {
	arrInt := make([]int32, rand.Intn(256))
	len := rand.Intn(256)
	arrTestStruct := make([]TestStruct, len)
	arrTestTestStruct := make([]*models.TestStruct, len)
	arrFloat := make([]float32, rand.Intn(256))

	if simple {
		arrInt = make([]int32, rand.Intn(3))
		len = rand.Intn(3)
		arrTestStruct = make([]TestStruct, len)
		arrTestTestStruct = make([]*models.TestStruct, len)
		arrFloat = make([]float32, rand.Intn(3))
	}

	for i := range arrInt {
		arrInt[i] = rand.Int31()
	}

	for i := range arrTestStruct {
		str := randString()
		str1 := randString()
		arrTestStruct[i] = TestStruct{Some: str, Other: str1}
		arrTestTestStruct[i] = &models.TestStruct{Some: str, Other: str1}
	}
	for i := range arrFloat {
		arrFloat[i] = rand.Float32()
	}

	t := Test{
		ID:               rand.Int31(),
		Name:             randString(),
		Tests:            arrTestStruct,
		SomeFloatArray:   arrFloat,
		SomeNumericArray: arrInt,
	}
	pt := models.Test{
		ID:               t.ID,
		Name:             t.Name,
		SomeNumericArray: arrInt,
		Tests:            arrTestTestStruct,
		SomeFloatArray:   arrFloat,
	}
	return t, pt
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
	ioutil.WriteFile("files/XML", buff.Bytes(), 0644)

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

	ioutil.WriteFile("files/Gob", buff.Bytes(), 0644)

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

	ioutil.WriteFile("files/Json", jsbytes, 0644)
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

	ioutil.WriteFile("files/Proto", out, 0644)
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

	ioutil.WriteFile("files/Avro", out, 0644)

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

	ioutil.WriteFile("files/YAML", yamlbytes, 0644)

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

	ioutil.WriteFile("files/MSG", msg, 0644)

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

	nruns := flag.Int("runs", 1, "number of runs for every test(int)")
	ntestsPtr := flag.Int("tests", 1, "number of tests")
	s := flag.Bool("s", false, "show every run report(bool)")
	si := flag.Bool("si", false, "show detail report about every run run(bool)")
	simplePtr := flag.Bool("simpleTest", false, "all arrays in test struct have less then 4 elements")
	flag.Parse()

	num_runs := *nruns
	silence := !*s
	silenceInside := !*si
	simple := *simplePtr
	ntests := *ntestsPtr

	Gob := make(map[string]int)
	Gob["size"] = 0
	Gob["inTime"] = 0
	Gob["outTime"] = 0

	XML := make(map[string]int)
	XML["size"] = 0
	XML["inTime"] = 0
	XML["outTime"] = 0

	JSON := make(map[string]int)
	JSON["size"] = 0
	JSON["inTime"] = 0
	JSON["outTime"] = 0

	Proto := make(map[string]int)
	Proto["size"] = 0
	Proto["inTime"] = 0
	Proto["outTime"] = 0

	Avro := make(map[string]int)
	Avro["size"] = 0
	Avro["inTime"] = 0
	Avro["outTime"] = 0

	YAML := make(map[string]int)
	YAML["size"] = 0
	YAML["inTime"] = 0
	YAML["outTime"] = 0

	MSG := make(map[string]int)
	MSG["size"] = 0
	MSG["inTime"] = 0
	MSG["outTime"] = 0

	for j := 0; j < ntests; j++ {

		fmt.Printf("=============================================Test=#%d=================================================\n", j)

		t, tproto := generateTest(simple)

		GOBsize, GOBIntime, GOBOuttime := 0, 0, 0

		for i := 0; i < num_runs; i++ {
			if !silence || !silenceInside {
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

		Gob["size"] += GOBsize
		Gob["inTime"] += GOBIntime
		Gob["outTime"] += GOBOuttime

		fmt.Printf("Avarage: ")
		fmt.Printf("GOB size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d", GOBsize, GOBIntime, GOBOuttime, GOBIntime+GOBOuttime)

		println("\n\n")

		XMLsize, XMLIntime, XMLOuttime := 0, 0, 0

		for i := 0; i < num_runs; i++ {
			if !silence || !silenceInside {
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

		XML["size"] += XMLsize
		XML["inTime"] += XMLIntime
		XML["outTime"] += XMLOuttime

		fmt.Printf("Avarage: ")
		fmt.Printf("XML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d", XMLsize, XMLIntime, XMLOuttime, XMLIntime+XMLOuttime)

		println("\n\n")

		Jsonsize, JsonIntime, JsonOuttime := 0, 0, 0

		for i := 0; i < num_runs; i++ {

			if !silence || !silenceInside {
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

		JSON["size"] += Jsonsize
		JSON["inTime"] += JsonIntime
		JSON["outTime"] += JsonOuttime

		fmt.Printf("Avarage: ")
		fmt.Printf("JSON size:  %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d", Jsonsize, JsonIntime, JsonOuttime, JsonIntime+JsonOuttime)
		println("\n\n")

		protoSize, protoInTime, protoOutTime := 0, 0, 0

		for i := 0; i < num_runs; i++ {
			if !silence || !silenceInside {
				fmt.Printf("------------Proto------RUN #%d---------------- \n", i)
			}
			F, S, T := ProtoSerialise(tproto, silenceInside)
			protoSize = F
			if !silence {
				fmt.Printf("Proto size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d", F, S, T, S+T)
			}
			protoInTime += S
			protoOutTime += T
		}

		protoInTime /= num_runs
		protoOutTime /= num_runs

		Proto["size"] += protoSize
		Proto["inTime"] += protoInTime
		Proto["outTime"] += protoOutTime

		fmt.Printf("Avarage: ")
		fmt.Printf("Proto size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", protoSize, protoInTime, protoOutTime, protoInTime+protoOutTime)
		println("\n\n")

		AvroSize, AvroInTime, AvroOutTime := 0, 0, 0

		for i := 0; i < num_runs; i++ {
			if !silence || !silenceInside {
				fmt.Printf("------------Avro------RUN #%d---------------- \n", i)
			}
			F, S, T := AvroSerialise(t, silenceInside)
			AvroSize = F
			if !silence {
				fmt.Printf("Avro size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d", F, S, T, S+T)
			}
			AvroInTime += S
			AvroOutTime += T
		}

		AvroInTime /= num_runs
		AvroOutTime /= num_runs

		Avro["size"] += AvroSize
		Avro["inTime"] += AvroInTime
		Avro["outTime"] += AvroOutTime

		fmt.Printf("Avarage: ")
		fmt.Printf("Avro size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", AvroSize, AvroInTime, AvroOutTime, AvroInTime+AvroOutTime)
		println("\n\n")

		YAMLSize, YAMLInTime, YAMLOutTime := 0, 0, 0

		for i := 0; i < num_runs; i++ {
			if !silence || !silenceInside {
				fmt.Printf("------------YAML------RUN #%d---------------- \n", i)
			}
			F, S, T := YAMLSerialise(t, silenceInside)
			YAMLSize = F
			if !silence {
				fmt.Printf("YAML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d", F, S, T, S+T)
			}
			YAMLInTime += S
			YAMLOutTime += T
		}

		YAMLInTime /= num_runs
		YAMLOutTime /= num_runs

		YAML["size"] += YAMLSize
		YAML["inTime"] += YAMLInTime
		YAML["outTime"] += YAMLOutTime

		fmt.Printf("Avarage: ")
		fmt.Printf("YAML size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n", YAMLSize, YAMLInTime, YAMLOutTime, YAMLInTime+YAMLOutTime)
		println("\n\n")

		MSGpSize, MSGpInTime, MSGpOutTime := 0, 0, 0

		for i := 0; i < num_runs; i++ {
			if !silence || !silenceInside {
				fmt.Printf("------------MSGPack------RUN #%d---------------- \n", i)
			}
			F, S, T := MSGpSerialise(t, silenceInside)
			MSGpSize = F
			if !silence {
				fmt.Printf("MSGPack size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d", F, S, T, S+T)
			}
			MSGpInTime += S
			MSGpOutTime += T
		}

		MSGpInTime /= num_runs
		MSGpOutTime /= num_runs

		MSG["size"] += MSGpSize
		MSG["inTime"] += MSGpInTime
		MSG["outTime"] += MSGpOutTime

		fmt.Printf("Avarage: ")
		fmt.Printf("MSGp size:   %d bytes, Serialize: %d nanosec, Deserialize: %d nanosec\n Total time: %d\n\n", MSGpSize, MSGpInTime, MSGpOutTime, MSGpInTime+MSGpOutTime)

	}

	fmt.Printf("Overall Gob\n Sum: size: %d serializationTime: %d deserializationTime: %d \nSumTime: %d\n", Gob["size"], Gob["inTime"], Gob["outTime"], Gob["inTime"]+Gob["outTime"])

	fmt.Printf("Overall XML\n Sum: size: %d serializationTime: %d deserializationTime: %d \nSumTime: %d\n", XML["size"], XML["inTime"], XML["outTime"], XML["inTime"]+XML["outTime"])

	fmt.Printf("Overall JSON\n Sum: size: %d serializationTime: %d deserializationTime: %d \nSumTime: %d\n", JSON["size"], JSON["inTime"], JSON["outTime"], JSON["inTime"]+JSON["outTime"])

	fmt.Printf("Overall Proto\n Sum: size: %d serializationTime: %d deserializationTime: %d \nSumTime: %d\n", Proto["size"], Proto["inTime"], Proto["outTime"], Proto["inTime"]+Proto["outTime"])

	fmt.Printf("Overall Avro\n Sum: size: %d serializationTime: %d deserializationTime: %d \nSumTime: %d\n", Avro["size"], Avro["inTime"], Avro["outTime"], Avro["inTime"]+Avro["outTime"])

	fmt.Printf("Overall YAML\n Sum: size: %d serializationTime: %d deserializationTime: %d \nSumTime: %d\n", YAML["size"], YAML["inTime"], YAML["outTime"], YAML["inTime"]+YAML["outTime"])

	fmt.Printf("Overall MSG\n Sum: size: %d serializationTime: %d deserializationTime: %d \nSumTime: %d\n", MSG["size"], MSG["inTime"], MSG["outTime"], MSG["inTime"]+MSG["outTime"])

}
