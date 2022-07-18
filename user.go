package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mattermost/logr/v2"

	fb_client "github.com/mattermost/focalboard/server/client"
	"github.com/mattermost/focalboard/server/model"
	fb_model "github.com/mattermost/focalboard/server/model"
	fb_utils "github.com/mattermost/focalboard/server/utils"

	mm_model "github.com/mattermost/mattermost-server/v6/model"
)

type runInfo struct {
	cfg        *Config
	logger     logr.Logger
	abort      chan struct{}
	admin      *AdminClient
	quiet      bool
	blockCount int64
	output     buffer
}

func (ri *runInfo) IncBlockCount(add int) {
	count := atomic.AddInt64(&ri.blockCount, int64(add))

	if !ri.quiet {
		const space = "                          "
		s := fmt.Sprintf("block count: %d%s", count, space)
		s = s[:30] + "\r"
		fmt.Print(s)
	}
}

type stats struct {
	ChannelCount int
	BoardCount   int
	CardCount    int
	TextCount    int
}

type makerInfo struct {
	Cfg         *Config
	WorkspaceID string
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
			WorkspaceID: channel.Id,
			User:        user,
			PropertyIds: make(map[string][]string),
		}

		for i := 0; i < ri.cfg.BoardsPerChannel; i++ {
			board := makeBoard(makerInfo)

			board, resp := insertBlock(client, channel.Id, board)
			if resp.Error != nil {
				return stats, fmt.Errorf("cannot insert blocks for board %s: %w", board.ID, resp.Error)
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

				_, resp = insertBlocks(client, channel.Id, cardBlocks)
				if resp.Error != nil {
					return stats, fmt.Errorf("cannot insert blocks for card %s: %w", card.ID, resp.Error)
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

			_, resp = insertBlocks(client, channel.Id, views)
			if resp.Error != nil {
				return stats, fmt.Errorf("cannot insert views for board %s: %w", board.ID, resp.Error)
			}
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

func insertBlock(client *Client, workspaceID string, block *model.Block) (*model.Block, *fb_client.Response) {
	blocks := []*fb_model.Block{block}
	b, _ := json.Marshal(blocks)
	r, err := client.FBclient.DoAPIPost(fmt.Sprintf("/workspaces/%s/blocks", workspaceID), string(b))
	if err != nil {
		return nil, fb_client.BuildErrorResponse(r, err)
	}
	defer closeBody(r)

	blocksNew := model.BlocksFromJSON(r.Body)
	return &blocksNew[0], fb_client.BuildResponse(r)
}

func insertBlocks(client *Client, workspaceID string, blocks []*model.Block) ([]model.Block, *fb_client.Response) {
	b, _ := json.Marshal(blocks)
	r, err := client.FBclient.DoAPIPost(fmt.Sprintf("/workspaces/%s/blocks", workspaceID), string(b))
	if err != nil {
		return nil, fb_client.BuildErrorResponse(r, err)
	}
	defer closeBody(r)

	return model.BlocksFromJSON(r.Body), fb_client.BuildResponse(r)
}

func closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}
}

func makeBoard(info *makerInfo) *fb_model.Block {
	id := fb_utils.NewID(fb_utils.IDTypeBoard)
	board := &fb_model.Block{
		ID:          id,
		RootID:      id,
		CreatedBy:   info.User.Id,
		ModifiedBy:  info.User.Id,
		Schema:      1,
		Type:        fb_model.TypeBoard,
		Title:       fmt.Sprintf("board %d", pickRandomInt(1, 10000)),
		Fields:      makeBoardFields(info),
		CreateAt:    mm_model.GetMillis(),
		UpdateAt:    mm_model.GetMillis(),
		WorkspaceID: info.WorkspaceID,
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

func makeBoardCardProperties(info *makerInfo) []interface{} {
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

	props := make(map[string]interface{})
	props["id"] = propId
	props["name"] = "Status"
	props["options"] = options
	props["type"] = "select"

	info.PropertyIds[propId] = optionIds
	return []interface{}{props}
}

func makeCard(info *makerInfo, boardID string) *fb_model.Block {
	card := &fb_model.Block{
		ID:          fb_utils.NewID(fb_utils.IDTypeCard),
		RootID:      boardID,
		ParentID:    boardID,
		CreatedBy:   info.User.Id,
		ModifiedBy:  info.User.Id,
		Schema:      1,
		Type:        fb_model.TypeCard,
		Title:       fmt.Sprintf("card %d", pickRandomInt(1, 10000)),
		Fields:      make(map[string]interface{}),
		CreateAt:    mm_model.GetMillis(),
		UpdateAt:    mm_model.GetMillis(),
		WorkspaceID: info.WorkspaceID,
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
		RootID:      boardID,
		ParentID:    boardID,
		CreatedBy:   info.User.Id,
		ModifiedBy:  info.User.Id,
		Schema:      1,
		Type:        fb_model.TypeView,
		Title:       "Board view",
		Fields:      makeViewFields(cards),
		CreateAt:    mm_model.GetMillis(),
		UpdateAt:    mm_model.GetMillis(),
		WorkspaceID: info.WorkspaceID,
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
			RootID:      boardID,
			ParentID:    cardID,
			CreatedBy:   info.User.Id,
			ModifiedBy:  info.User.Id,
			Schema:      1,
			Type:        fb_model.TypeText,
			Title:       lorem(info.Cfg),
			Fields:      make(map[string]interface{}),
			CreateAt:    mm_model.GetMillis(),
			UpdateAt:    mm_model.GetMillis(),
			WorkspaceID: info.WorkspaceID,
		}
		blocks = append(blocks, block)
	}
	return blocks
}
