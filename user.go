package main

import (
	"fmt"
	"time"

	"github.com/mattermost/focalboard/server/model"
	fb_model "github.com/mattermost/focalboard/server/model"
	fb_utils "github.com/mattermost/focalboard/server/utils"

	mm_model "github.com/mattermost/mattermost-server/v6/model"
)

type makerInfo struct {
	Cfg         *Config
	TeamID      string
	ChannelID   string
	User        *mm_model.User
	PropertyIds map[string][]string
}

func runUser(username string, ri *runInfo) (stats, error) {
	stats := stats{}

	// create user
	user, err := ri.admin.CreateUser(username, ri.cfg.teamID)
	if err != nil {
		return stats, fmt.Errorf("cannot create user: %w", err)
	}

	// add user to team
	_, _, err = ri.admin.client.AddTeamMember(ri.cfg.teamID, user.Id)
	if err != nil {
		return stats, fmt.Errorf("could not add user %s to team %s: %w", user.Username, ri.cfg.teamID, err)
	}

	password := "test-password-1" // reverseString(user.Username)

	client, err := NewClient(ri.cfg.SiteURL, user.Username, password)
	if err != nil {
		return stats, fmt.Errorf("cannot create client: %w", err)
	}
	stats.UserCount++

	// create channels, boards, cards, and content
	for i := 0; i < ri.cfg.ChannelsPerUser; i++ {
		channelName := fmt.Sprintf("%s-%d", user.Username, i+1)
		channel, err := client.CreateChannel(channelName, ri.cfg.TeamID())
		if err != nil {
			return stats, fmt.Errorf("cannot create channel %s: %w", channelName, err)
		}
		stats.ChannelCount++

		makerInfo := &makerInfo{
			Cfg:         ri.cfg,
			TeamID:      ri.cfg.teamID,
			ChannelID:   channel.Id,
			User:        user,
			PropertyIds: make(map[string][]string),
		}

		for i := 0; i < ri.cfg.BoardsPerChannel; i++ {
			boardNew := makeBoard(makerInfo)

			board, err := client.InsertBoard(boardNew)
			if err != nil {
				return stats, fmt.Errorf("cannot insert board %s: %w", boardNew.Title, err)
			}
			stats.BoardCount++
			ri.IncBlockCount(1)

			cards := make([]*fb_model.Block, 0, ri.cfg.CardsPerBoard)

			for j := 0; j < ri.cfg.CardsPerBoard; j++ {
				card := makeCard(makerInfo, board.ID)
				content := makeContent(makerInfo, ri.cfg.GetMaxParagraphs(), board.ID, card.ID)
				card.Fields = makeCardFields(makerInfo, content)

				cardBlocks := make([]*fb_model.Block, 0, len(content)+1)
				cardBlocks = append(cardBlocks, card)
				cardBlocks = append(cardBlocks, content...)

				_, err = insertBlocks(client, board.ID, cardBlocks)
				if err != nil {
					return stats, fmt.Errorf("cannot insert blocks for card %s: %w", card.ID, err)
				}

				cards = append(cards, card)

				stats.CardCount++
				stats.TextCount += len(content)

				ri.IncBlockCount(len(content) + 1)

				select {
				case <-ri.abort:
					return stats, nil
				case <-time.After(time.Millisecond * ri.cfg.CardDelay):
				default:
				}
			}

			views := makeViews(cards, makerInfo, board.ID)

			_, err = insertBlocks(client, board.ID, views)
			if err != nil {
				return stats, fmt.Errorf("cannot insert views for board %s: %w", board.ID, err)
			}
			stats.ViewCount += len(views)
			ri.IncBlockCount(len(views))

			if ri.cfg.BoardDelay != 0 {
				select {
				case <-ri.abort:
					return stats, fmt.Errorf("aborting user %s", user.Username)
				case <-time.After(time.Millisecond * ri.cfg.BoardDelay):
				}
			}
		}
	}
	return stats, nil
}

func insertBlocks(client *Client, boardID string, blocks []*model.Block) ([]model.Block, error) {
	insertBlocks := make([]model.Block, 0, len(blocks))
	for _, b := range blocks {
		insertBlocks = append(insertBlocks, *b)
	}
	return client.InsertBlocks(boardID, insertBlocks)
}

func makeBoard(info *makerInfo) *fb_model.Board {
	board := &fb_model.Board{
		TeamID:          info.TeamID,
		ChannelID:       info.ChannelID,
		CreatedBy:       info.User.Id,
		ModifiedBy:      info.User.Id,
		Type:            fb_model.BoardTypeOpen,
		Title:           fmt.Sprintf("board %d", pickRandomInt(1, 10000)),
		Icon:            randomIcon(),
		ShowDescription: false,
		IsTemplate:      false,
		TemplateVersion: 0,
		Properties:      nil,
		CardProperties:  makeBoardCardProperties(info),
		CreateAt:        mm_model.GetMillis(),
		UpdateAt:        mm_model.GetMillis(),
	}
	return board
}

func makeBoardFields(info *makerInfo) map[string]interface{} {
	fields := make(map[string]interface{})
	fields["cardProperties"] = makeBoardCardProperties(info)
	fields["columnCalculations"] = make([]interface{}, 0)
	fields["description"] = ""
	fields["icon"] = randomIcon()
	fields["isTemplate"] = false
	fields["showDescription"] = false
	fields["templateVer"] = 0
	return fields
}

func makeBoardCardProperties(info *makerInfo) []map[string]interface{} {
	optionIds := make([]string, 0)
	options := make([]map[string]string, 0)

	for _, val := range []string{"Good", "Bad", "Ugly"} {
		id := fb_utils.NewID(fb_utils.IDTypeNone)
		options = append(options, map[string]string{
			"color": pickRandomString([]string{
				"propColorGray", "propColorBrown", "propColorOrange", "propColorYellow",
				"propColorGreen", "propColorBlue", "propColorPurple", "propColorPink", "propColorRed",
			}),
			"id":    id,
			"value": val,
		})
		optionIds = append(optionIds, id)
	}

	propId := fb_utils.NewID(fb_utils.IDTypeNone)

	property := make(map[string]interface{})
	property["id"] = propId
	property["name"] = "Status"
	property["options"] = options
	property["type"] = "select"

	info.PropertyIds[propId] = optionIds
	return []map[string]interface{}{property}
}

func makeCard(info *makerInfo, boardID string) *fb_model.Block {
	card := &fb_model.Block{
		ID:          fb_utils.NewID(fb_utils.IDTypeCard),
		BoardID:     boardID,
		ParentID:    boardID,
		CreatedBy:   info.User.Id,
		ModifiedBy:  info.User.Id,
		Schema:      1,
		Type:        fb_model.TypeCard,
		Title:       fmt.Sprintf("card %d", pickRandomInt(1, 10000)),
		Fields:      make(map[string]interface{}),
		CreateAt:    mm_model.GetMillis(),
		UpdateAt:    mm_model.GetMillis(),
		WorkspaceID: info.TeamID,
	}
	return card
}

func makeCardFields(info *makerInfo, contentBlocks []*fb_model.Block) map[string]interface{} {
	fields := make(map[string]interface{})
	fields["icon"] = randomIcon()
	fields["isTemplate"] = false
	fields["properties"] = struct{}{}

	order := make([]string, 0, len(contentBlocks))
	for _, block := range contentBlocks {
		order = append(order, block.ID)
	}
	fields["contentOrder"] = order

	props := make(map[string]string)
	for propId, options := range info.PropertyIds {
		props[propId] = options[pickRandomInt(0, len(options))]
	}
	fields["properties"] = props

	return fields
}

func makeViews(cards []*fb_model.Block, info *makerInfo, boardID string) []*fb_model.Block {
	view := &fb_model.Block{
		ID:          fb_utils.NewID(fb_utils.IDTypeView),
		BoardID:     boardID,
		ParentID:    boardID,
		CreatedBy:   info.User.Id,
		ModifiedBy:  info.User.Id,
		Schema:      1,
		Type:        fb_model.TypeView,
		Title:       "Board view",
		Fields:      makeViewFields(cards),
		CreateAt:    mm_model.GetMillis(),
		UpdateAt:    mm_model.GetMillis(),
		WorkspaceID: info.TeamID,
	}
	return []*fb_model.Block{view}
}

func makeViewFields(cards []*fb_model.Block) map[string]interface{} {
	fields := make(map[string]interface{})
	fields["collaspedOptionIds"] = []interface{}{}
	fields["columnCalculations"] = struct{}{}
	fields["columnWidths"] = struct{}{}
	fields["defaultTemplateId"] = ""
	fields["viewType"] = "board"

	order := make([]string, 0, len(cards))
	for _, block := range cards {
		order = append(order, block.ID)
	}
	fields["cardOrder"] = order
	return fields
}

func makeContent(info *makerInfo, count int, boardID string, cardID string) []*fb_model.Block {
	blocks := make([]*fb_model.Block, 0, count)
	for i := 0; i < count; i++ {
		block := &fb_model.Block{
			ID:          fb_utils.NewID(fb_utils.IDTypeBlock),
			BoardID:     boardID,
			ParentID:    cardID,
			CreatedBy:   info.User.Id,
			ModifiedBy:  info.User.Id,
			Schema:      1,
			Type:        fb_model.TypeText,
			Title:       lorem(info.Cfg),
			Fields:      make(map[string]interface{}),
			CreateAt:    mm_model.GetMillis(),
			UpdateAt:    mm_model.GetMillis(),
			WorkspaceID: info.TeamID,
		}
		blocks = append(blocks, block)
	}
	return blocks
}
