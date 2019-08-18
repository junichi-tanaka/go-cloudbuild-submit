package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"google.golang.org/api/cloudbuild/v1"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		projectId string
		repoName string
		branchName string
		topicName string
	)
	flag.StringVar(&projectId, "P", "", "your gcp `project_id`")
	flag.StringVar(&repoName, "R", "", "`repository_name`")
	flag.StringVar(&branchName, "B", "", "`branch_name`")
	flag.StringVar(&topicName, "T", "", "`topic_name`")
	flag.Parse()

	if projectId == "" || repoName == "" || branchName == "" || topicName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.Background()
	cloudbuildService, err := cloudbuild.NewService(ctx)
	if err != nil {
		panic(err)
	}

	buildsService := cloudbuild.NewProjectsBuildsService(cloudbuildService)

	file, err := os.Open("cloudbuild.yaml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	build := &cloudbuild.Build{}
	err = yaml.Unmarshal(buffer, build)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", build.Steps[0])

	substitutions := map[string]string{}
	substitutions["_FUNCTIONS_TOPIC"] = topicName
	substitutions["_PROJECT_ID"] = projectId
	substitutions["_REPO_NAME"] = repoName
	substitutions["_BRANCH_NAME"] = branchName
	build.Substitutions = substitutions

	op, err := buildsService.Create(projectId, build).Context(ctx).Do()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", op.Name)

	// Wait until done
	for !op.Done {
		time.Sleep(10 * time.Second)

		opService := cloudbuild.NewOperationsService(cloudbuildService)
		op, err = opService.Get(op.Name).Context(ctx).Do()
		if err != nil {
			panic(err)
		}
	}

	// TODO: format the result
	fmt.Printf("%v\n", string(op.Response))
}
