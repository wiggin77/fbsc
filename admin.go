package main

import (
	"fmt"
	"sync"

	mm_model "github.com/mattermost/mattermost-server/v6/model"
)

type AdminClient struct {
	mux    sync.Mutex
	client *mm_model.Client4
}

// NewAdminClient creates a new admin client that is logged into a Mattermost server.
func NewAdminClient(cfg *Config) (*AdminClient, error) {
	client := mm_model.NewAPIv4Client(cfg.SiteURL)
	if _, _, err := client.Login(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		return nil, err
	}

	admin := &AdminClient{
		client: client,
	}
	return admin, nil
}

// CreateTeam creates a new team in idempotent manner.
func (ac *AdminClient) CreateTeam(name string, open bool) (*mm_model.Team, error) {
	ac.mux.Lock()
	defer ac.mux.Unlock()

	team, _, err := ac.client.GetTeamByName(name, "")
	if err == nil {
		return team, nil
	}

	teamType := mm_model.TeamOpen
	if !open {
		teamType = mm_model.TeamInvite
	}

	teamNew := &mm_model.Team{
		Name:            name,
		DisplayName:     name,
		Description:     "Team created by MMSC",
		Type:            teamType,
		AllowOpenInvite: open,
	}

	team, _, err = ac.client.CreateTeam(teamNew)
	if err != nil {
		return nil, fmt.Errorf("cannot create team %s: %w", name, err)
	}
	return team, nil
}

// CreateChannel creates a new channel in a idempotent manner.
func (ac *AdminClient) CreateChannel(channelName string, teamId string) (*mm_model.Channel, error) {
	ac.mux.Lock()
	defer ac.mux.Unlock()

	channel, _, err := ac.client.GetChannelByName(channelName, teamId, "")
	if err != nil {
		return channel, nil
	}

	me, _, err := ac.client.GetMe("")
	if err != nil {
		return nil, err
	}

	channelNew := &mm_model.Channel{
		TeamId:      teamId,
		Type:        mm_model.ChannelTypeOpen,
		Name:        channelName,
		DisplayName: channelName,
		Header:      "A channel created by FBSC.",
		CreatorId:   me.Id,
	}

	channel, _, err = ac.client.CreateChannel(channelNew)
	if err != nil {
		return nil, fmt.Errorf("cannot create channel %s: %w", channelName, err)
	}
	return channel, nil
}

// CreateUser creates a new user in a idempotent manner.
func (ac *AdminClient) CreateUser(username string, teamID string) (*mm_model.User, error) {
	ac.mux.Lock()
	defer ac.mux.Unlock()

	user, _, err := ac.client.GetUserByUsername(username, "")
	if err == nil {
		return user, nil
	}

	userNew := &mm_model.User{
		Username:      username,
		Password:      "test-password-1", //reverseString(username),
		Email:         fmt.Sprintf("%s@example.com", username),
		EmailVerified: true,
	}

	user, _, err = ac.client.CreateUser(userNew)
	if err != nil {
		return nil, fmt.Errorf("cannot create user %s: %w", username, err)
	}
	return user, nil
}
