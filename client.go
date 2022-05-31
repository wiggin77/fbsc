package main

import (
	"fmt"

	fb_client "github.com/mattermost/focalboard/server/client"

	fb_model "github.com/mattermost/focalboard/server/model"

	mm_model "github.com/mattermost/mattermost-server/v6/model"
)

type Client struct {
	FBclient *fb_client.Client
	MMclient *mm_model.Client4

	user *mm_model.User
}

func NewClient(siteURL string, username string, password string) (*Client, error) {
	mmclient := mm_model.NewAPIv4Client(siteURL)

	_, _, err := mmclient.Login(username, password)
	if err != nil {
		return nil, err
	}

	fbclient := fb_client.NewClient(siteURL, mmclient.AuthToken)

	me, _, err := mmclient.GetMe("")
	if err != nil {
		return nil, fmt.Errorf("cannot fetch user %s: %w", username, err)
	}

	return &Client{
		FBclient: fbclient,
		MMclient: mmclient,
		user:     me,
	}, nil
}

func (c *Client) InsertBlocks(blocks []fb_model.Block) ([]fb_model.Block, error) {
	blocks, resp := c.FBclient.InsertBlocks(blocks)
	return blocks, resp.Error
}

// CreateChannel creates a new channel in an idempotent manner.
func (c *Client) CreateChannel(channelName string, teamId string) (*mm_model.Channel, error) {
	channel, _, _ := c.MMclient.GetChannelByName(channelName, teamId, "")
	if channel != nil {
		return channel, nil
	}

	channelNew := &mm_model.Channel{
		TeamId:      teamId,
		Type:        mm_model.ChannelTypeOpen,
		Name:        channelName,
		DisplayName: channelName,
		Header:      "A channel created by FBSC.",
		CreatorId:   c.user.Id,
	}

	channel, _, err := c.MMclient.CreateChannel(channelNew)
	if err != nil {
		return nil, fmt.Errorf("cannot create channel %s: %w", channelName, err)
	}
	return channel, nil
}
