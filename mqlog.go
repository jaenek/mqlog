package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type Page struct {
	Topics []string
}

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

func (p* Page) datahandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/mqlog/"):]
	if filename == "" {
		p.rendertemplate(w, r, "topics.html")
		return
	}
	servefile(w, r, filename)
}

func (p* Page) filehandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/mqlog/"):]
	if filename == "" || filename == "sp.html" {
		p.rendertemplate(w, r, "public/sp.html")
		return
	} else if filename == "topics.html" {
		p.rendertemplate(w, r, "public/"+filename)
		return
	}
	servefile(w, r, "public/"+filename)
}

func servefile(w http.ResponseWriter, r *http.Request, filename string) {
	log.WithFields(log.Fields{
		"file": filename,
	}).Info("Serving file.")
	_, err := os.Open(filename)
	if os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"file": filename,
		}).Error(err.Error())
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filename)
}

func (p* Page) rendertemplate(w http.ResponseWriter, r *http.Request, tmpl string) {
	log.WithFields(log.Fields{
		"file": tmpl,
	}).Info("Rendering template.")

	t, err := template.ParseFiles(tmpl)
	if err != nil {
		log.WithFields(log.Fields{
			"file": tmpl,
		}).Error(err.Error())
		http.NotFound(w, r)
		return
	}

	err = t.Execute(w, p)
	if err != nil {
		log.WithFields(log.Fields{
			"file": tmpl,
		}).Error(err.Error())
		http.NotFound(w, r)
		return
	}
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

	p := &Page{
		Topics: strings.Split(*topics, ","),
	}

	c := mqttinit(*mqtthost, *mqttport, p.Topics)

	http.HandleFunc("/mqlog/topics/", p.datahandler)
	http.HandleFunc("/mqlog/", p.filehandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))

	c.Disconnect(250)
}
