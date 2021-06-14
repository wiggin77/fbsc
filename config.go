package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	SiteURL string `json:"site_url"`

	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`

	Usernames []string `json:"user_names"` // used when using specific usernames
	UserCount int      `json:"user_count"` // used when creating random users

	Workspaces []string `json:"workspaces"` // workspaces to create and/or use

	BoardCount int `json:"board_count"`
	CardCount  int `json:"card_count"`

	AvgActionDelay int64   `json:"avg_action_delay"` // average milliseconds between actions
	DelayVariance  float32 `json:"delay_variance"`   // how much the actual delay can randomly vary from averge (0.0 - 1.0)

	ProbComment  float32 `json:"prob_comment"`  // probability a user will comment on a card (0.0 - 1.0)
	ProbProperty float32 `json:"prob_property"` // probability a user will add/modify a card property (0.0 - 1.0)
	ProbEdit     float32 `json:"prob_edit"`     // probability a user will edit their own card (0.0 - 1.0)
	ProbDelete   float32 `json:"prob_delete"`   // probability a user will delete their own card (0.0 - 1.0)

	MaxWordsPerSentence      int `json:"max_words_per_sentence"`
	MaxSentencesPerParagraph int `json:"max_sentences_per_paragraph"`
	MaxParagraphsPerComment  int `json:"max_paragraphs_per_comment"`
}

func createDefaultConfig(filename string) error {
	cfg := Config{
		Usernames:                make([]string, 0),
		UserCount:                DefaultUserCount,
		AvgActionDelay:           DefaultAvgActionDelay,
		ProbComment:              DefaultProbComment,
		ProbProperty:             DefaultProbProperty,
		MaxWordsPerSentence:      DefaultMaxWordsPerSentence,
		MaxSentencesPerParagraph: DefaultMaxSentencesPerParagraph,
		MaxParagraphsPerComment:  DefaultMaxParagraphsPerComment,
	}

	b, err := json.MarshalIndent(&cfg, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, FilePerms)
}

func loadConfig(filename string) (*Config, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) GetMaxWordsPerSentence() int {
	return c.MaxWordsPerSentence
}

func (c *Config) GetMaxSentencesPerParagraph() int {
	return c.MaxSentencesPerParagraph
}

func (c *Config) GetMaxParagraphs() int {
	return c.MaxParagraphsPerComment
}
