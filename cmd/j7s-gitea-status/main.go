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
	"code.gitea.io/sdk/gitea"
	"flag"
	"log"
	"os"
)

func main() {
	url := os.Getenv("GIT_URL")
	token := os.Getenv("GIT_TOKEN")
	owner := os.Getenv("GIT_OWNER")
	repo := os.Getenv("GIT_REPO")

	if url == "" || token == "" || owner == "" || repo == "" {
		log.Fatal("Must set env variables GIT_URL, GIT_TOKEN, GIT_OWNER, and GIT_REPO.")
	}

	sha := flag.String("sha", "", "Commit to mark with status.")
	context := flag.String("name", "", "gitea context for this status, e.g. test/mytest")
	statusFlag := flag.String("status", "", "status for the commit.")
	description := flag.String("description", "", "Text description for the gitea user.")
	targetUrl := flag.String("url", "", "URL to show the gitea user.")
	flag.Parse()

	if *sha == "" {
		log.Fatal("Must provide a commit.")
	}
	if *context == "" {
		log.Fatal("Must provide a name.")
	}
	status, ok := convertToStatus(*statusFlag)
	if !ok {
		log.Fatal("Status must be a valid string.")
	}

	clientOpt := gitea.SetToken(token)
	client, err := gitea.NewClient(url, clientOpt)
	if err != nil {
		log.Fatal(err)
	}

	statusOption := gitea.CreateStatusOption{State: status, TargetURL: *targetUrl, Description: *description, Context: *context}
	_, _, err = client.CreateStatus(owner, repo, *sha, statusOption)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Set status for of", status, "for commit", *sha, ".")
}

func convertToStatus(inStatus string) (gitea.StatusState, bool) {
	switch inStatus {
	// First two cases are valid from Tekton as a pipeline run status.
	case "Succeeded", "Completed":
		return gitea.StatusSuccess, true
	case "Failed", "None":
		return gitea.StatusFailure, true
	// Not valid from Tekton as a pipeline run, but still useful.
	case "Pending":
		return gitea.StatusPending, true
	case "Warning":
		return gitea.StatusWarning, true
	case "Error":
		return gitea.StatusError, true
	default:
		return gitea.StatusError, false
	}
}
