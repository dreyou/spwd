package main

import (
	"bytes"
	json "encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func elasticsearch(inProc ProcAll, sendProcesses bool) {
	logger(DEBUG, func() { log.Println("Send data to elasticsearch") })
	if sendProcesses {
		num := 0
		for _, p := range inProc.Processes.All {
			if p.ProcLoad > Conf.Elasticsearch.LoadTreshold || p.MemLoad > Conf.Elasticsearch.MemTreshold {
				p.TimeStamp = inProc.TimeStamp
				p.Hostname = inProc.Kernel.Hostname
				p.HostId = inProc.HostId
				j, err := json.Marshal(p)
				if err == nil {
					sendJsonToElasticsearch(j, "process")
					num++
				} else {
					logger(ERROR, func() { log.Printf("JSON error in process%v\n", err) })
				}
			}
		}
		logger(DEBUG, func() { log.Printf("Sended info about %v processes\n", num) })
	}
	inProc.Processes.All = nil
	j, err := json.Marshal(inProc)
	if err == nil {
		sendJsonToElasticsearch(j, "proc")
	} else {
		logger(ERROR, func() { log.Printf("JSON error in proc %v\n", err) })
	}
}

func sendJsonToElasticsearch(j []byte, typeName string) {
	postUrl := Conf.Elasticsearch.Url + "/spwd/" + typeName + "/?pretty"
	client := &http.Client{}
	resp, err := client.Post(postUrl, "application/json", bytes.NewBuffer(j))
	if err == nil {
		logger(TRACE, func() { log.Printf("Data sended to: %v, %v\n ", postUrl, resp) })
	} else {
		logger(ERROR, func() { log.Printf("Data NOT sended to: %v\n", err) })
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}
