// Copyright 2023 James Pace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"bytes"
	api "code.gitea.io/gitea/modules/structs"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	postUrl := flag.String("postUrl", "", "URL for Texton Trigger.")
	flag.Parse()
	context := NewHandlerContext(*postUrl)
	http.HandleFunc("/hook", context.hook)
	log.Println("Listening on :8080/hook")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (ctx *HandlerContext) hook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Make sure we're dealing with the right type of event.
	event := r.Header.Get("X-Gitea-Event")
	if event != "push" {
		log.Println("Got unsupported event of type", event)
		http.Error(w, "Unsupported event", 400)
		return
	}
	// Now unmarshall.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read body.")
		http.Error(w, "Failed to read body", 400)
		return
	}
	var inPayload api.PushPayload
	err = json.Unmarshal(body, &inPayload)
	if err != nil {
		log.Printf("Failed to parse body.")
		http.Error(w, "Failed to parse payload.", 400)
		return
	}

	// Make map of stuff to send to the trigger.
	outMap := make(map[string]string)
	outMap["sha"] = inPayload.HeadCommit.ID
	outMap["repo_owner"] = inPayload.Repo.Owner.UserName
	outMap["repo_name"] = inPayload.Repo.Name
	outJson, err := json.Marshal(outMap)
	if err != nil {
		log.Println("Failed to make json response.")
		http.Error(w, "Failed to make json response.", 400)
		return
	}

	// Now send it.
	request, err := http.NewRequest("POST", ctx.postUrl, bytes.NewBuffer(outJson))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Failed to make request. err:", err)
		http.Error(w, "Failed to pass on request.", 400)
		return
	}
	defer response.Body.Close()

	// It worked!
	log.Println("Success!")
	return
}

type HandlerContext struct {
	postUrl string
}

func NewHandlerContext(postUrl string) *HandlerContext {
	return &HandlerContext{postUrl: postUrl}
}
