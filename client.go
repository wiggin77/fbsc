package main

import (
	"fmt"
	"net/url"
	"path"
	"strings"

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

	user, _, err := mmclient.Login(username, password)
	if err != nil {
		return nil, err
	}

	fbclient := fb_client.NewClient(siteURL, mmclient.AuthToken)
	u, err := url.Parse(siteURL)
	if err != nil {
		return nil, fmt.Errorf("Invalid site URL: %w", err)
	}
	u.Path = path.Join("/plugins/focalboard/", fb_client.APIURLSuffix)
	fbclient.APIURL = u.String()

	me2, resp := fbclient.GetMe()
	if resp.Error != nil {
		return nil, fmt.Errorf("cannot fetch FocalBoard user %s: %w", username, resp.Error)
	}

	if me2.ID != user.Id {
		return nil, fmt.Errorf("user ids don't match %s != %s: %w", user.Id, me2.ID, err)
	}

	return &Client{
		FBclient: fbclient,
		MMclient: mmclient,
		user:     user,
	}, nil
}

func (c *Client) InsertBoard(board *fb_model.Board) (*fb_model.Board, error) {
	boardNew, resp := c.FBclient.CreateBoard(board)
	return boardNew, resp.Error
}

func (c *Client) InsertBlocks(boardID string, blocks []fb_model.Block) ([]fb_model.Block, error) {
	blocks, resp := c.FBclient.InsertBlocks(boardID, blocks, true)
	return blocks, resp.Error
}

// CreateChannel creates a new channel in an idempotent manner.
func (c *Client) CreateChannel(channelName string, teamId string) (*mm_model.Channel, error) {
	displayName := channelName

	channelName = strings.ToLower(channelName)
	channelName = strings.ReplaceAll(channelName, ".", "-")
	channelName = strings.ReplaceAll(channelName, " ", "")

	channel, _, _ := c.MMclient.GetChannelByName(channelName, teamId, "")
	if channel != nil {
		return channel, nil
	}

	channelNew := &mm_model.Channel{
		TeamId:      teamId,
		Type:        mm_model.ChannelTypeOpen,
		Name:        channelName,
		DisplayName: displayName,
		Header:      "A channel created by FBSC.",
		CreatorId:   c.user.Id,
	}

	channel, _, err := c.MMclient.CreateChannel(channelNew)
	if err != nil {
		return nil, fmt.Errorf("cannot create channel %s: %w", channelName, err)
	}
	return channel, nil
}
