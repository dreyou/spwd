package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	synca "sync/atomic"
	"time"
)

import gcfg "code.google.com/p/gcfg"

const NEED_DATA int32 = 1

const NO_DATA = 0

const WAIT_DATA_PERIOD = 5

const ERROR = 0

const INFO = 1

const DEBUG = 2

var dataFlag int32

var dataPeriod int32

var proc ProcAll

var mutex = &sync.Mutex{}

var webRoot = flag.String("webroot", "./web/", "Path to web root directory")

var level = flag.Int("level", 2, "Log level (0 -ERROR, 1 - INFO, 2 - DEBUG)")

var config = flag.String("config", "./spwd.gcfg", "Config file")

var Conf Config

type Config struct {
	Main struct {
		UpdateInterval time.Duration
		Listen         string
		Allow          string
	}
	Js struct {
		MaxPoints     int
		TimeToRefresh int
		TimeToReload  int
	}
}

var defaultConf = `;Defaulf config
[Main]
;Update statictics data in miliiseconds
UpdateInterval = 5000
Listen = localhost:4000
Allow = 127.0.0.1
[Js]
MaxPoints = 10
;Time to refresh page, min
TimeToRefresh = 10
;Time to reload data, sec
TimeToReload = 2
`

var jsConfig = `//Javascrit constants
var maxPoints = %v;
var timeToRefresh = %v;
var timeToReload = %v;
`

func initConf() {
	if _, err := os.Stat(*config); os.IsNotExist(err) {
		log.Printf("no such file or directory: %s, creatina\n", *config)
		ioutil.WriteFile(*config, []byte(defaultConf), 0644)
	}
	err := gcfg.ReadFileInto(&Conf, *config)
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
		mutex.Unlock()
		synca.StoreInt32(&dataFlag, NO_DATA)
		dataPeriod--
	} else {
		logger(DEBUG, func() { log.Println("Skip data") })
	}
}

func (hf HttpFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if notAllow(r) {
		http.Error(w, "403 Forbidden : you can't access this resource.", 403)
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
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http js config: " + req) })
	js := fmt.Sprintf(jsConfig, Conf.Js.MaxPoints, (Conf.Js.TimeToRefresh * 60 * 1000), (Conf.Js.TimeToReload * 1000))
	headers := w.Header()
	headers["Content-Type"] = []string{"text/javascript"}
	w.Write([]byte(js))
}

func fileStat(w http.ResponseWriter, r *http.Request) {
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http proc stat request: " + req) })
	b, _ := json.Marshal(proc.Stat)
	w.Write(b)
}

func fileMeminfo(w http.ResponseWriter, r *http.Request) {
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http proc meminfo request: " + req) })
	b, _ := json.Marshal(proc.Meminfo)
	w.Write(b)
}

func fileProc(w http.ResponseWriter, r *http.Request) {
	synca.StoreInt32(&dataFlag, NEED_DATA)
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http proc request: " + req) })
	mutex.Lock()
	b, _ := json.Marshal(proc)
	mutex.Unlock()
	w.Write(b)
}

func fileProcesses(w http.ResponseWriter, r *http.Request) {
	req := r.URL.RequestURI()
	logger(DEBUG, func() { log.Println("Servig http proc request: " + req) })
	b, _ := json.Marshal(proc.Processes.All)
	w.Write(b)
}

func fileRoot(w http.ResponseWriter, r *http.Request) {
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

func notAllow(r *http.Request) bool {
	addr := regexp.MustCompile(`:`).Split(r.RemoteAddr, -1)[0]
	return !regexp.MustCompile(Conf.Main.Allow).MatchString(addr)
}

func main() {
	flag.Parse()
	logger(INFO, func() { log.Println("Start") })
	initConf()
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

	dataPeriod = WAIT_DATA_PERIOD

	synca.StoreInt32(&dataFlag, NEED_DATA)

	ticker := time.NewTicker(time.Millisecond * Conf.Main.UpdateInterval)
	go func() {
		for _ = range ticker.C {
			procUpdater()
		}
	}()

	http.HandleFunc("/ps", fileProcesses)
	http.HandleFunc("/proc", fileProc)
	http.HandleFunc("/stat", fileStat)
	http.HandleFunc("/meminfo", fileMeminfo)
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
