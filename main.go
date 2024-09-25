package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chainguard-dev/terraform-infra-common/modules/github-bots/sdk"
	"github.com/google/go-github/v61/github"
	"sigs.k8s.io/yaml"
)

func main() {
	yamlFile, err := os.ReadFile("label-syncer.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	var labelSyncer Labels
	err = yaml.Unmarshal(yamlFile, &labelSyncer)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	policyName := os.Getenv("POLICY_NAME")
	if policyName == "" {
		log.Fatalf("POLICY_NAME environment variable not set")
	}
	ctx := context.Background()
	client := sdk.NewGitHubClient(ctx, labelSyncer.Org, labelSyncer.PolicyRepo, policyName)
	defer client.Close(ctx)

	opts := &github.RepositoryListByOrgOptions{
		Type: "all",
	}
	repos := []*github.Repository{}
	for {
		moreRepos, resp, err := client.Client().Repositories.ListByOrg(context.Background(), labelSyncer.Org, opts)
		if err != nil {
			log.Fatalf("Error client.Repositories.List: %v", err)
		}

		repos = append(repos, moreRepos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	for _, repo := range repos {
		if repo.GetFork() {
			fmt.Println("Skipping forked repo: ", repo.GetName())
			continue
		}

		ignoreFlag := false
		for _, ignore := range labelSyncer.IgnoreRepos {
			if ignore.RepoName == repo.GetName() {
				log.Printf("Skipping repo %s it is in the ignore list\n", repo.GetName())
				ignoreFlag = true
				break
			}
		}
		if ignoreFlag {
			continue
		}

		log.Printf("Will process labels for %s\n", repo.GetName())
		for _, label := range labelSyncer.General {
			_, _, err := client.Client().Issues.CreateLabel(context.Background(), labelSyncer.Org, repo.GetName(), &github.Label{
				Name:        &label.Name,
				Color:       &label.Color,
				Description: &label.Description,
			})
			if err != nil && err.(*github.ErrorResponse).Response.StatusCode != 422 {
				log.Fatalf("Error client.Issues.CreateLabel: %v", err)
			}
		}

		for _, labelRepo := range labelSyncer.Repos {
			log.Printf("Applying repo specific labels for %s\n", repo.GetName())
			if labelRepo.RepoName == repo.GetName() {
				_, _, err := client.Client().Issues.CreateLabel(context.Background(), labelSyncer.Org, repo.GetName(), &github.Label{
					Name:        &labelRepo.Labels.Name,
					Color:       &labelRepo.Labels.Color,
					Description: &labelRepo.Labels.Description,
				})
				if err != nil && err.(*github.ErrorResponse).Response.StatusCode != 422 {
					log.Fatalf("Error client.Issues.CreateLabel: %v", err)
				}
			}
		}
	}

	log.Printf("Done\n")
}
