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

	statusFlag := flag.String("status", "", "Status for the commit.")
	title := flag.String("title", "", "Title for issue.")
	body := flag.String("body", "", "Body for issue.")
	flag.Parse()

	if *title == "" {
		log.Fatal("Must provide a title.")
	}

	isIssue, ok := shouldIssueIssue(*statusFlag)
	if !ok {
		log.Fatal("Status must be a valid string.")
	}

	if !isIssue {
		log.Println("Status is fine, don't need to issue an issue.")
		os.Exit(0)
	}

	clientOpt := gitea.SetToken(token)
	client, err := gitea.NewClient(url, clientOpt)
	if err != nil {
		log.Fatal(err)
	}

	issueOption := gitea.CreateIssueOption{Title: *title, Body: *body}
	issue, _, err := client.CreateIssue(owner, repo, issueOption)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Filed issue at", issue.HTMLURL)
}

func shouldIssueIssue(inStatus string) (bool, bool) {
	switch inStatus {
	// First two cases are valid from Tekton as a pipeline run status.
	case "Succeeded", "Completed":
		return false, true
	case "Failed", "None":
		return true, true
	default:
		return false, false
	}
}
