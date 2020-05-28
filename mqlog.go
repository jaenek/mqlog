package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

func mqttinit() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("localhost:1883")
	opts.SetClientID("mqlog")

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	c.Subscribe("test/data", 0, callback)

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
		os.MkdirAll(filename[:len(filename)-len("/data")], 0755)
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
	// TODO add topic menu
	filename := "topics/test/data"
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
	port := flag.String("p", "8080", "port to serve on")
	directory := flag.String("d", "./public", "the directory of static file to host")
	flag.Parse()

	c := mqttinit()

	log.WithFields(log.Fields{
		"directory": *directory,
		"port": *port,
	}).Info("File server started.")

	http.HandleFunc("/sp.data", datahandler)
	http.Handle("/", http.FileServer(http.Dir(*directory)))
	log.Fatal(http.ListenAndServe(":"+*port, nil))

	c.Disconnect(250)
}
