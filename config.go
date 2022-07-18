package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type Config struct {
	SiteURL string `json:"site_url"`

	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`

	TeamName string `json:"team_name"` // optional: if empty then a new team will be created
	teamID   string

	ConcurrentUsers int `json:"concurrent_users"` // number of users to simulate concurrently
	UserCount       int `json:"user_count"`       // number of users to create

	ChannelsPerUser  int `json:"channels_per_user"`
	BoardsPerChannel int `json:"boards_per_channel"`
	CardsPerBoard    int `json:"cards_per_board"`

	MaxWordsPerSentence      int `json:"max_words_per_sentence"`
	MaxSentencesPerParagraph int `json:"max_sentences_per_paragraph"`
	MaxParagraphsPerComment  int `json:"max_paragraphs_per_comment"`

	BoardDelay time.Duration `json:"board_delay_ms"`
	CardDelay  time.Duration `json:"card_delay_ms"`
}

func createDefaultConfig(filename string) error {
	cfg := Config{
		SiteURL:                  "",
		AdminUsername:            "",
		AdminPassword:            "",
		TeamName:                 "",
		ConcurrentUsers:          DefaultConcurrentUsers,
		UserCount:                DefaultUserCount,
		ChannelsPerUser:          DefaultChannelsPerUser,
		BoardsPerChannel:         DefaultBoardsPerChannel,
		CardsPerBoard:            DefaultCardsPerBoard,
		MaxWordsPerSentence:      DefaultMaxWordsPerSentence,
		MaxSentencesPerParagraph: DefaultMaxSentencesPerParagraph,
		MaxParagraphsPerComment:  DefaultMaxParagraphsPerComment,
		BoardDelay:               DefaultBoardDelay,
		CardDelay:                DefaultCardDelay,
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

func (c *Config) TeamID() string {
	return c.teamID
}

func (c *Config) setTeamID(teamID string) {
	c.teamID = teamID
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
