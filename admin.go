package main

import (
	"fmt"
	"sync"

	"github.com/mattermost/mattermost-server/model"
)

type AdminClient struct {
	mux    sync.Mutex
	client *model.Client4
}

// NewAdminClient creates a new admin client that is logged into a Mattermost server.
func NewAdminClient(cfg *Config) (*AdminClient, error) {
	client := model.NewAPIv4Client(cfg.SiteURL)
	if _, resp := client.Login(cfg.AdminUsername, cfg.AdminPassword); !isSuccess(resp) {
		return nil, resp.Error
	}

	admin := &AdminClient{
		client: client,
	}
	return admin, nil
}

// CreateTeam creates a new team in idempotent manner.
func (ac *AdminClient) CreateTeam(name string, open bool) (*model.Team, error) {
	ac.mux.Lock()
	defer ac.mux.Unlock()

	team, resp := ac.client.GetTeamByName(name, "")
	if isSuccess(resp) {
		return team, nil
	}

	teamType := model.TEAM_OPEN
	if !open {
		teamType = model.TEAM_INVITE
	}

	teamNew := &model.Team{
		Name:            name,
		DisplayName:     name,
		Description:     "Team created by MMSC",
		Type:            teamType,
		AllowOpenInvite: open,
	}

	team, resp = ac.client.CreateTeam(teamNew)
	if !isSuccess(resp) {
		return nil, fmt.Errorf("cannot create team %s: %w", name, resp.Error)
	}
	return team, nil
}

// CreateChannel creates a new channel in a idempotent manner.
func (ac *AdminClient) CreateChannel(channelName string, teamId string) (*model.Channel, error) {
	ac.mux.Lock()
	defer ac.mux.Unlock()

	channel, resp := ac.client.GetChannelByName(channelName, teamId, "")
	if isSuccess(resp) {
		return channel, nil
	}

	me, resp := ac.client.GetMe("")
	if !isSuccess(resp) {
		return nil, resp.Error
	}

	channelNew := &model.Channel{
		TeamId:      teamId,
		Type:        model.CHANNEL_OPEN,
		Name:        channelName,
		DisplayName: channelName,
		Header:      "A channel created by MMSC.",
		CreatorId:   me.Id,
	}

	channel, resp = ac.client.CreateChannel(channelNew)
	if !isSuccess(resp) {
		return nil, fmt.Errorf("cannot create channel %s: %w", channelName, resp.Error)
	}
	return channel, nil
}

// CreateUser creates a new user in a idempotent manner.
func (ac *AdminClient) CreateUser(username string) (*model.User, error) {
	ac.mux.Lock()
	defer ac.mux.Unlock()

	user, resp := ac.client.GetUserByUsername(username, "")
	if isSuccess(resp) {
		return user, nil
	}

	userNew := &model.User{
		Username: username,
		Password: username,
		Email:    fmt.Sprintf("%s@example.com", username),
	}

	user, resp = ac.client.CreateUser(userNew)
	if !isSuccess(resp) {
		return nil, fmt.Errorf("cannot create user %s: %w", username, resp.Error)
	}
	return user, nil
}

// AddUserToTeam adds a user to a team in idempotent manner.
func (ac *AdminClient) AddUserToTeam(userId string, teamId string) error {
	ac.mux.Lock()
	defer ac.mux.Unlock()

	if _, resp := ac.client.AddTeamMember(teamId, userId); !isSuccess(resp) {
		return fmt.Errorf("cannot add user %s to team %s: %w", userId, teamId, resp.Error)
	}
	return nil
}

// AddUserToChannel adds a user to a channel in idempotent manner.
func (ac *AdminClient) AddUserToChannel(userId string, channelId string) error {
	ac.mux.Lock()
	defer ac.mux.Unlock()

	if _, resp := ac.client.AddChannelMember(channelId, userId); !isSuccess(resp) {
		return fmt.Errorf("cannot add user %s to channel %s: %w", userId, channelId, resp.Error)
	}
	return nil
}
