package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	SiteURL string `json:"site_url"`

	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`

	TeamName string `json:"team_name"`
	TeamId   string `json:"-"`

	Usernames []string `json:"user_names"` // used when using specific usernames
	UserCount int      `json:"user_count"` // used when creating random users

	ChannelNames []string `json:"channel_names"` // channels to create and/or post to
	ChannelIds   []string `json:"-"`

	AvgPostDelay  int64   `json:"avg_post_delay"` // average milliseconds between posting
	DelayVariance float32 `json:"delay_variance"` // how much the actual delay can randomly vary from averge (0.0 - 1.0)

	ProbReact  float32 `json:"prob_react"`  // probability a recieved post will be reacted to (0.0 - 1.0)
	ProbReply  float32 `json:"prob_reply"`  // probability a recieved post will be replied to (0.0 - 1.0)
	ProbEdit   float32 `json:"prob_edit"`   // probability a user will edit their own post (0.0 - 1.0)
	ProbDelete float32 `json:"prob_delete"` // probability a user will delete their own post (0.0 - 1.0)

	MaxWordsPerSentence      int `json:"max_words_per_sentence"`
	MaxSentencesPerParagraph int `json:"max_sentences_per_paragraph"`
	MaxParagraphsPerPost     int `json:"max_paragraphs_per_post"`
}

func createDefaultConfig(filename string) error {
	cfg := Config{
		TeamName:                 DefaultTeamName,
		Usernames:                make([]string, 0),
		UserCount:                DefaultUserCount,
		AvgPostDelay:             DefaultAvgPostDelay,
		ProbReact:                DefaultProbReact,
		ProbReply:                DefaultProbReply,
		MaxWordsPerSentence:      DefaultMaxWordsPerSentence,
		MaxSentencesPerParagraph: DefaultMaxSentencesPerParagraph,
		MaxParagraphsPerPost:     DefaultMaxParagraphsPerPost,
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
	return c.MaxParagraphsPerPost
}
