package main

import (
	"fmt"

	"github.com/mattermost/logr"

	fb_model "github.com/mattermost/focalboard/server/model"
	fb_utils "github.com/mattermost/focalboard/server/utils"

	mm_model "github.com/mattermost/mattermost-server/v6/model"
)

type runInfo struct {
	cfg    *Config
	logger logr.Logger
	done   chan struct{}
	admin  *AdminClient
}

type stats struct {
	ChannelCount int
	BoardCount   int
	CardCount    int
	TextCount    int
}

func (s stats) add(s2 stats) stats {
	return stats{
		ChannelCount: s.ChannelCount + s2.ChannelCount,
		BoardCount:   s.BoardCount + s2.BoardCount,
		CardCount:    s.CardCount + s2.CardCount,
		TextCount:    s.TextCount + s2.CardCount,
	}
}

func runUser(username string, ri runInfo) (stats, error) {
	stats := stats{}

	// create user
	user, err := ri.admin.CreateUser(username)
	if err != nil {
		return stats, err
	}

	client, err := NewClient(ri.cfg.SiteURL, user.Username, user.Username)
	if err != nil {
		return stats, err
	}

	// create channels, boards, cards, and content
	for i := 0; i < ri.cfg.ChannelsPerUser; i++ {
		channelName := fmt.Sprintf("%s-%d", user.Username, i+1)
		channel, err := client.CreateChannel(channelName, ri.cfg.TeamID())
		if err != nil {
			return stats, fmt.Errorf("cannot create channel %s: %w", channelName, err)
		}
		stats.ChannelCount++

		boards := makeBoards(ri.cfg.BoardsPerChannel, channel.Id, user)

		for _, board := range boards {
			blocks := make([]fb_model.Block, 0, ri.cfg.CardsPerBoard*7+1)
			blocks = append(blocks, board)
			var content []fb_model.Block

			cards := makeCards(ri.cfg.CardsPerBoard, channel.Id, board.ID, user)
			for _, card := range cards {
				content = makeContent(ri.cfg, pickRandomInt(1, 7), channel.Id, board.ID, card.ID, user)
				blocks = append(blocks, content...)

				select {
				case <-ri.done:
					return stats, fmt.Errorf("aborting user %s", user.Username)
				default:
				}
			}

			_, resp := client.FBclient.InsertBlocks(blocks)
			if resp.Error != nil {
				return stats, fmt.Errorf("cannot insert blocks for board %s: %w", board.ID, resp.Error)
			}
			stats.BoardCount++
			stats.CardCount += len(cards)
			stats.TextCount += len(content)
		}
	}
	return stats, nil
}

func makeBoards(count int, workspaceID string, creator *mm_model.User) []fb_model.Block {
	blocks := make([]fb_model.Block, 0, count)
	for i := 0; i < count; i++ {
		id := fb_utils.NewID(fb_utils.IDTypeBoard)
		board := fb_model.Block{
			ID:          id,
			RootID:      id,
			CreatedBy:   creator.Id,
			ModifiedBy:  creator.Id,
			Schema:      1,
			Type:        fb_model.TypeBoard,
			Title:       fmt.Sprintf("board %d", pickRandomInt(1, 10000)),
			WorkspaceID: workspaceID,
		}
		blocks = append(blocks, board)
	}
	return blocks
}

func makeCards(count int, workspaceID string, boardID string, creator *mm_model.User) []fb_model.Block {
	blocks := make([]fb_model.Block, 0, count)
	for i := 0; i < count; i++ {
		card := fb_model.Block{
			ID:          fb_utils.NewID(fb_utils.IDTypeCard),
			RootID:      boardID,
			ParentID:    boardID,
			CreatedBy:   creator.Id,
			ModifiedBy:  creator.Id,
			Schema:      1,
			Type:        fb_model.TypeCard,
			Title:       fmt.Sprintf("card %d", pickRandomInt(1, 10000)),
			WorkspaceID: workspaceID,
		}
		blocks = append(blocks, card)
	}
	return blocks
}

func makeContent(cfg *Config, count int, workspaceID string, boardID string, cardID string, creator *mm_model.User) []fb_model.Block {
	blocks := make([]fb_model.Block, 0, count)
	for i := 0; i < count; i++ {
		block := fb_model.Block{
			ID:          fb_utils.NewID(fb_utils.IDTypeBlock),
			RootID:      boardID,
			ParentID:    cardID,
			CreatedBy:   creator.Id,
			ModifiedBy:  creator.Id,
			Schema:      1,
			Type:        fb_model.TypeText,
			Title:       lorem(cfg),
			WorkspaceID: workspaceID,
		}
		blocks = append(blocks, block)
	}
	return blocks
}
