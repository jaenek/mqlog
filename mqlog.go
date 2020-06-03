package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

func mqttinit(hostname, port string, topics []string) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(hostname+":"+port)
	opts.SetClientID("mqlog")

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for _, topic := range topics {
		c.Subscribe(topic, 0, callback)
	}

	return c
}

func callback(client mqtt.Client, msg mqtt.Message) {
	log.WithFields(log.Fields{
		"topic": msg.Topic(),
		"message": msg.Payload(),
	}).Info("Message Received.")

	filename := "topics/"+msg.Topic()

	data, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		os.MkdirAll(filename[:len(filename)-len(filepath.Base(filename))], 0755)
	}

	data = append(data, append(msg.Payload(), '\n')...)
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"file": filename,
		}).Error(err.Error())
	}
}

func datahandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/mqlog/"):]
	servefile(w, r, filename)
}

func filehandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/mqlog/"):]
	servefile(w, r, "public/"+filename)
}

func servefile(w http.ResponseWriter, r *http.Request, filename string) {
	_, err := os.Open(filename)
	if os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"file": filename,
		}).Error(err.Error())
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filename)
	log.WithFields(log.Fields{
		"file": filename,
	}).Info("Serving file.")
}

func main() {
	port := flag.String("p", "8000", "port to serve on")
	mqtthost := flag.String("h", "localhost", "hostname for mqtt broker")
	mqttport := flag.String("m", "1883", "port for mqtt broker")
	directory := flag.String("d", "./public", "the directory of static file to host")
	topics := flag.String("t", "test/data", "topic to subscribe to")
	flag.Parse()

	log.WithFields(log.Fields{
		"port": *port,
		"mqtthost": *mqtthost,
		"mqttport": *mqttport,
		"directory": *directory,
		"topics": *topics,
	}).Info("File server started.")

	c := mqttinit(*mqtthost, *mqttport, strings.Split(*topics, ","))

	http.HandleFunc("/mqlog/topics/", datahandler)
	http.HandleFunc("/mqlog/", filehandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))

	c.Disconnect(250)
}
