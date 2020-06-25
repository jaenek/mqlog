package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type Page struct {
	Topics []string
}

type Mqlog struct {
	client mqtt.Client
	page Page
}

func (m* Mqlog) init(hostname, port string, topics []string) mqtt.Client {
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

	lines := []byte(strconv.Itoa(bytes.Count(data, []byte("\n"))))
	line := bytes.Join([][]byte{lines, []byte(";"), msg.Payload(), []byte("\n")}, []byte(""))
	data = append(data, line...)
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"file": filename,
		}).Error(err.Error())
	}
}

func (m* Mqlog) datahandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/mqlog/"):]
	if filename == "" {
		m.rendertemplate(w, r, "topics.html")
		return
	}
	servefile(w, r, filename)
}

func (m* Mqlog) filehandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/mqlog/"):]
	if filename == "" || filename == "sp.html" {
		m.rendertemplate(w, r, "public/sp.html")
		return
	} else if filename == "topics.html" {
		m.rendertemplate(w, r, "public/"+filename)
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

func (m* Mqlog) rendertemplate(w http.ResponseWriter, r *http.Request, tmpl string) {
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

	err = t.Execute(w, m.page)
	if err != nil {
		log.WithFields(log.Fields{
			"file": tmpl,
		}).Error(err.Error())
		http.NotFound(w, r)
		return
	}
}

func (m* Mqlog) addhandler(w http.ResponseWriter, r *http.Request) {
	topic := r.FormValue("newtopic")
	log.WithFields(log.Fields{
		"topic": topic,
	}).Info("Add new topic.")

	if m.page.Topics[0] == "" {
		m.page.Topics = []string{topic}
	} else {
		m.page.Topics = append(m.page.Topics, topic)
	}

	m.client.Subscribe(topic, 0, callback)

	http.Redirect(w, r, "/mqlog/", http.StatusFound)
}

func main() {
	port := flag.String("p", "8000", "port to serve on")
	mqtthost := flag.String("h", "localhost", "hostname for mqtt broker")
	mqttport := flag.String("m", "1883", "port for mqtt broker")
	directory := flag.String("d", "./public", "the directory of static file to host")
	topics := flag.String("t", "", "topic to subscribe to")
	flag.Parse()

	log.WithFields(log.Fields{
		"port": *port,
		"mqtthost": *mqtthost,
		"mqttport": *mqttport,
		"directory": *directory,
		"topics": *topics,
	}).Info("File server started.")

	m := &Mqlog{}

	m.page = Page{
		Topics: strings.Split(*topics, ","),
	}

	m.client = m.init(*mqtthost, *mqttport, m.page.Topics)

	http.HandleFunc("/mqlog/topics/", m.datahandler)
	http.HandleFunc("/mqlog/add", m.addhandler)
	http.HandleFunc("/mqlog/", m.filehandler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))

	m.client.Disconnect(250)
}
