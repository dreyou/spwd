package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	synca "sync/atomic"
	"syscall"
	"time"
)

import gcfg "code.google.com/p/gcfg"

const KERNEL_VERSION = 26

const NEED_DATA int32 = 1

const NO_DATA = 0

const WAIT_DATA_PERIOD = 5

const ERROR = 0

const INFO = 1

const DEBUG = 2

const TRACE = 3

var dataFlag int32

var sendFlag int32

var dataPeriod int32

var proc ProcAll

var procJson []byte

var mutex = &sync.Mutex{}

var webRoot = flag.String("webroot", "./web/", "Path to web root directory")

var level = flag.Int("level", 2, "Log level (0 -ERROR, 1 - INFO, 2 - DEBUG)")

var config = flag.String("config", "./spwd.gcfg", "Config file")

var pidFile = flag.String("pid", "/tmp/spwd.pid", "Pid file")

var Conf Config

type Config struct {
	Main struct {
		UpdateInterval time.Duration
		SendInterval   time.Duration
		Listen         string
		Allow          string
	}
	Js struct {
		MaxPoints     int
		TimeToRefresh int
		TimeToReload  int
	}
	Elasticsearch struct {
		HostId        string
		Send          bool
		SendProcesses bool
		Url           string
		LoadTreshold  float32
		MemTreshold   float32
	}
}

var defaultConf = `;Defaulf config
[Main]
;Update statictics data in miliiseconds
UpdateInterval = 5000
;Send statictics data in seconds
SendInterval = 60
Listen = localhost:4000
Allow = 127.0.0.1
[Js]
MaxPoints = 10
;Time to refresh page, min
TimeToRefresh = 10
;Time to reload data, sec
TimeToReload = 2
[Elasticsearch]
Url = http://localhost:9200
;Additional host id
HostId = default_host_id
;Send main data to elasticsearch index spwd with type proc
Send = false
;Send processes data to elasticsearch index spwd with type processes
SendProcesses = false
;Process data will be sended to elastiicsearch when process processor load (%) > LoadTreshold
LoadTreshold = 0.1
;or process memory usage (%) > MemTreshold
MemTreshold = 0.1
`

var jsConfig = `//Javascrit constants
var maxPoints = %v;
var timeToRefresh = %v;
var timeToReload = %v;
`
var senders = []Sender{}

type Sender func(proc ProcAll, sendProcesses bool)

func initConf() {
	if _, err := os.Stat(*config); os.IsNotExist(err) {
		log.Printf("no such file or directory: %s, creating\n", *config)
		ioutil.WriteFile(*config, []byte(defaultConf), 0644)
	}
	err := gcfg.ReadFileInto(&Conf, *config)
	if err != nil {
		log.Fatal(err)
	}
}

func removePid() {
	if _, err := os.Stat(*pidFile); os.IsNotExist(err) {
		log.Printf("no such file or directory: %s\n", *pidFile)
		return
	}
	err := os.Remove(*pidFile)
	if err != nil {
		log.Fatal(err)
	}
}

func writePid(pid int) {
	err := ioutil.WriteFile(*pidFile, []byte(fmt.Sprintf("%v", pid)), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func logger(lvl int, run func()) {
	if lvl <= *level {
		run()
	}
}

type HttpFile struct {
	reqPath  string
	fileType string
	fileName string
}

func procUpdater() {
	if synca.LoadInt32(&dataFlag) == NEED_DATA || dataPeriod != 0 {
		logger(DEBUG, func() { log.Println("Read data") })
		if synca.LoadInt32(&dataFlag) == NEED_DATA {
			dataPeriod = WAIT_DATA_PERIOD + 1
		}
		mutex.Lock()
		proc.Update()
		procJson, _ = json.Marshal(proc)
		mutex.Unlock()
		synca.StoreInt32(&dataFlag, NO_DATA)
		dataPeriod--
	} else {
		logger(DEBUG, func() { log.Println("Skip data") })
	}
}

var sendChan = make(chan int)

func procSender() {
	for {
		<-sendChan
		localProc := ProcAll{}
		localProc.Init()
		localProc.HostId = Conf.Elasticsearch.HostId
		time.Sleep(time.Millisecond * Conf.Main.UpdateInterval)
		localProc.Update()
		for _, send := range senders {
			send(localProc, Conf.Elasticsearch.SendProcesses)
		}
	}
}

func (hf HttpFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if notAllow(w, r) {
		return
	}
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http file request: " + req) })
	headers := w.Header()
	headers["Content-Type"] = []string{hf.fileType}
	logger(DEBUG, func() { log.Println("Using header: " + hf.fileType) })
	var fileName string
	if hf.fileName != "" {
		fileName = hf.fileName
	} else {
		fileName = strings.Replace(req, "/", "", 1)
	}
	fileName = *webRoot + fileName
	logger(DEBUG, func() { log.Println("Using file: " + fileName) })
	file, _ := ioutil.ReadFile(fileName)
	w.Write(file)
}

func fileJsConfig(w http.ResponseWriter, r *http.Request) {
	if notAllow(w, r) {
		return
	}
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http js config: " + req) })
	js := fmt.Sprintf(jsConfig, Conf.Js.MaxPoints, (Conf.Js.TimeToRefresh * 60 * 1000), (Conf.Js.TimeToReload * 1000))
	headers := w.Header()
	headers["Content-Type"] = []string{"text/javascript"}
	w.Write([]byte(js))
}

func fileProc(w http.ResponseWriter, r *http.Request) {
	if notAllow(w, r) {
		return
	}
	synca.StoreInt32(&dataFlag, NEED_DATA)
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http proc request: " + req) })
	mutex.Lock()
	w.Write(procJson)
	mutex.Unlock()
}

func fileRoot(w http.ResponseWriter, r *http.Request) {
	if notAllow(w, r) {
		return
	}
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http requesti for root: " + req) })
	if req == "/favicon.ico" {
		logger(DEBUG, func() { log.Println("Returning favicon.ico") })
		fileIco, _ := ioutil.ReadFile(*webRoot + "favicon.ico")
		w.Write(fileIco)
		return
	} else {
		logger(DEBUG, func() { log.Println("Returning root.html") })
		file, _ := ioutil.ReadFile(*webRoot + "root.html")
		w.Write(file)
	}
}

func fileLog(w http.ResponseWriter, r *http.Request) {
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Logging http requests: " + req + " from: " + r.RemoteAddr) })
	if req == "/" {
		logger(DEBUG, func() { log.Println("Root request, redirecting to root.html") })
		fileRoot(w, r)
	}
}

func notAllow(w http.ResponseWriter, r *http.Request) bool {
	allowExp := strings.Replace(Conf.Main.Allow, `.`, `\.`, -1)
	allowExp = strings.Replace(allowExp, `*`, `.+`, -1)
	logger(TRACE, func() { log.Println("Allow regexp:" + allowExp) })
	addr := regexp.MustCompile(`:`).Split(r.RemoteAddr, -1)[0]
	res := !regexp.MustCompile(allowExp).MatchString(addr)
	if res {
		http.Error(w, "403 Forbidden : you can't access this resource.", 403)
	}
	return res
}

func main() {
	if !checkVersion(KERNEL_VERSION) {
		log.Fatal("Invalid kernel version!")
	}
	flag.Parse()
	writePid(os.Getpid())
	logger(INFO, func() { log.Println("Start") })
	initConf()
	senders = append(senders, elasticsearch)
	var handlers = []HttpFile{
		{"/jquery.min.js", "text/javascript", "jquery.min.js"},
		{"/jquery.jqplot.min.js", "text/javascript", "jquery.jqplot.min.js"},
		{"/plugins/", "text/javascript", ""},
		{"/jquery-ui.min.js", "text/javascript", "jquery-ui.min.js"},
		{"/jquery.dataTables.min.js", "text/javascript", "jquery.dataTables.min.js"},
		{"/jquery-ui.min.css", "text/css", "jquery-ui.min.css"},
		{"/jquery.jqplot.min.css", "text/css", "jquery.jqplot.min.css"},
		{"/jquery.dataTables.css", "text/css", "jquery.dataTables.css"},
		{"/root.css", "text/css", "root.css"},
		{"/images/", "image/png", ""},
		{"/root", "text/html", "root.html"},
		{"/favicon.ico", "image/x-icon", "favicon.ico"},
	}
	proc.Init()
	proc.HostId = Conf.Elasticsearch.HostId

	dataPeriod = WAIT_DATA_PERIOD

	synca.StoreInt32(&dataFlag, NEED_DATA)

	updateTicker := time.NewTicker(time.Millisecond * Conf.Main.UpdateInterval)
	go func() {
		for _ = range updateTicker.C {
			procUpdater()
		}
	}()

	go procSender()

	sendTicker := time.NewTicker(time.Second * Conf.Main.SendInterval)
	go func() {
		for _ = range sendTicker.C {
			sendChan <- 1
		}
	}()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		sig := <-signalChannel
		logger(INFO, func() { log.Printf("Exitting by signal %v\n", sig) })
		removePid()
		os.Exit(0)
	}()

	http.HandleFunc("/proc", fileProc)
	http.HandleFunc("/config.js", fileJsConfig)

	for _, handler := range handlers {
		http.Handle(handler.reqPath, handler)
	}

	http.HandleFunc("/", fileLog)

	err := http.ListenAndServe(Conf.Main.Listen, nil)
	if err != nil {
		log.Fatal(err)
	}
}
