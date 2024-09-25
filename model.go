package main

type Labels struct {
	Org        string `yaml:"org"`
	PolicyRepo string `yaml:"policyRepo"`
	General    []struct {
		Name        string `yaml:"name"`
		Color       string `yaml:"color"`
		Description string `yaml:"description"`
	} `yaml:"general"`
	IgnoreRepos []struct {
		RepoName string `yaml:"repoName"`
	} `yaml:"ignoreRepos"`
	Repos []struct {
		RepoName string `yaml:"repoName"`
		Labels   struct {
			Name        string `yaml:"name"`
			Color       string `yaml:"color"`
			Description string `yaml:"description"`
		} `yaml:"labels"`
	} `yaml:"repos"`
}
