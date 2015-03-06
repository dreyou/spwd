package main

import (
	"bytes"
	json "encoding/json"
	"log"
	"net/http"
)

func elasticsearch(proc ProcAll, jsonProc []byte) {
	if jsonProc == nil {
		j, _ := json.Marshal(proc)
		sendJsonToElasticsearch(j)
	} else {
		sendJsonToElasticsearch(jsonProc)
	}
}

func sendJsonToElasticsearch(j []byte) {
	postUrl := Conf.Elasticsearch.Url + "/spwd/proc/?pretty"
	resp, err := http.Post(postUrl, "application/json", bytes.NewBuffer(j))
	if err == nil {
		logger(DEBUG, func() { log.Printf("Data sended to: %v, %v\n ", postUrl, resp) })
	} else {
		logger(ERROR, func() { log.Println("Data NOT sended to: %v, %v, %v\n", postUrl, resp, err) })
	}
}
