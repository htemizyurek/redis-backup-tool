package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const address string = "127.0.0.1:6379"
const network string = "tcp"

func main() {
	log.Println("Starting...")

	var fileName = GetFileName()
	database := flag.String("database", "0", "Redis database.")
	file := flag.String("file", fileName, "File address.")
	flag.Parse()
	redis := RedisConnect{}
	redis.Connect()
	selectedDatabase := redis.Exec("SELECT " + *database)
	log.Printf("Selected database: %v - status %v", string(*database), selectedDatabase)
	dbsize := redis.Exec("DBSIZE")
	log.Printf("DB Size: %v", dbsize)
	allKeys := redis.Exec("keys *")
	patternForAllKeys := regexp.MustCompile(`\r\n`)
	keys := patternForAllKeys.Split(allKeys, -1)
	findKeys := make([]string, len(keys))
	for _, v := range keys {
		if v != "" && v[:1] != "$" && v[:1] != "*" {
			findKeys = append(findKeys, strings.Trim(v, ""))
		}

	}

	var result map[string]string
	result = make(map[string]string)

	for _, i := range findKeys {
		if i != "" {
			test := redis.Get(string(i))
			result[i] = test
		}

	}

	jsonString, _ := json.Marshal(result)
	WriteFile(string(jsonString), *file)
}

/*RedisConnect struct*/
type RedisConnect struct {
	Connection net.Conn
}

/*Connect to redis*/
func (r *RedisConnect) Connect() {
	connection, err := net.Dial(network, address)
	if nil != err {
		log.Panicf("Could not open TCP connection %v", err)
	}

	log.Println("Redis connection successful")
	r.Connection = connection
}

/*Exec query for redis */
func (r *RedisConnect) Exec(query string) string {
	r.Connection.Write([]byte(query + "\r\n"))
	data := r.ReadData()
	return data
}

/*ReadData read on redis*/
func (r *RedisConnect) ReadData() string {
	command := make([]byte, 1024)
	n, err := r.Connection.Read(command)
	if nil != err {
		log.Panicf("Could not read data %v", err)
	}
	return string(command[:n])
}

/*Get redis get query */
func (r *RedisConnect) Get(query string) string {
	r.Connection.Write([]byte("get " + query + " \r\n"))
	command2 := make([]byte, 1024)
	_, _ = r.Connection.Read(command2)
	reply, _ := doBulkReply(command2[1:])
	str := fmt.Sprintf("%v", reply)
	return str
}

func doBulkReply(reply []byte) (interface{}, error) {
	pos := getFlagPos('\r', reply)
	pstart := 0
	if reply[:pos][0] == '$' {
		pstart = 1
	}

	vlen, err := strconv.Atoi(string(reply[pstart:pos]))
	if err != nil {
		return nil, err
	}
	if vlen == -1 {
		return nil, nil
	}

	start := pos + 2
	end := start + vlen
	return string(reply[start:end]), nil
}

func getFlagPos(flag byte, reply []byte) int {
	pos := 0
	for _, v := range reply {
		if v == flag {
			break
		}
		pos++
	}

	return pos
}

/*GetFileName create a file name. */
func GetFileName() string {
	var currentTime = time.Now()
	var currentMonth int = int(currentTime.Month())
	var fileName string = "./" + strconv.Itoa(currentTime.Year()) + "-" + strconv.Itoa(currentMonth) + "-" + strconv.Itoa(currentTime.Day())
	return fileName + ".json"
}

/*WriteFile this method write a file to json data. */
func WriteFile(data string, file string) {
	byteData := []byte(data)
	err := ioutil.WriteFile(file, byteData, 0644)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Operation successful. File : ", file)
}
