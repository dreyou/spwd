package main

import (
	"bytes"
	json "encoding/json"
	"log"
	"net/http"
)

func elasticsearch(proc ProcAll, sendProcesses bool) {
	if sendProcesses {
		for _, p := range proc.Processes.All {
			p.TimeStamp = proc.TimeStamp
			p.Hostname = proc.Kernel.Hostname
			p.HostId = proc.HostId
			j, _ := json.Marshal(p)
			sendJsonToElasticsearch(j, "process")
		}
	}
	proc.Processes = Processes{}
	j, _ := json.Marshal(proc)
	sendJsonToElasticsearch(j, "proc")
}

func sendJsonToElasticsearch(j []byte, typeName string) {
	postUrl := Conf.Elasticsearch.Url + "/spwd/" + typeName + "/?pretty"
	resp, err := http.Post(postUrl, "application/json", bytes.NewBuffer(j))
	if err == nil {
		logger(TRACE, func() { log.Printf("Data sended to: %v, %v\n ", postUrl, resp) })
	} else {
		logger(ERROR, func() { log.Println("Data NOT sended to: %v, %v, %v\n", postUrl, resp, err) })
	}
}
