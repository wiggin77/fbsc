package main

import (
	"time"

	"github.com/mattermost/logr"
	"github.com/mattermost/mattermost-server/model"
)

type postActionFunc func(post *model.Post, channelId string) error

type Action struct {
	f             postActionFunc
	name          string
	actOnOwnPosts bool
}

type runInfo struct {
	cfg    *Config
	logger logr.Logger
	done   chan struct{}
	admin  *AdminClient
}

func runUser(username string, ri runInfo) error {
	sim, err := NewUserSim(username, ri)
	if err != nil {
		return err
	}

	avgDelay := sim.ri.cfg.AvgActionDelay
	variance := sim.ri.cfg.DelayVariance

	var actions = []Action{
		{f: sim.Reply, name: "reply", actOnOwnPosts: false},
		{f: sim.React, name: "react", actOnOwnPosts: false},
		{f: sim.Edit, name: "edit", actOnOwnPosts: true},
		{f: sim.Delete, name: "delete", actOnOwnPosts: true},
	}

	for {
		delay := randomDuration(avgDelay, variance)

		if wait(delay, ri.done) {
			return nil
		}

		sim.Post(lorem(sim.ri.cfg), pickRandomString(sim.ri.cfg.ChannelIds))

		for _, channelId := range sim.ri.cfg.ChannelIds {
			postList, err := getPostsForChannelAroundLastUnread(sim.client, sim.userId, channelId)
			if err != nil {
				sim.ri.logger.Errorf("Cannot getPostsForChannelAroundLastUnread for user %s: %v", sim.username, err)
				continue
			}

			for _, postId := range postList.Order {
				post, resp := sim.client.GetPost(postId, "")
				if !isSuccess(resp) {
					sim.ri.logger.Errorf("Cannot get post %s for user %s: %v", postId, sim.username, err)
					continue
				}

				for _, action := range actions {
					if (action.actOnOwnPosts && post.UserId != sim.userId) || (!action.actOnOwnPosts && post.UserId == sim.userId) {
						continue
					}

					if wait(10, ri.done) {
						return nil
					}
					if err := action.f(post, channelId); err != nil {
						sim.ri.logger.Errorf("Cannot %s for post %s in channel %s for user %s: %v",
							action.name, post.Id, channelId, sim.username, resp.Error)
					}
				}
			}
		}
	}
}

func wait(delay int64, done chan struct{}) bool {
	select {
	case <-done:
		return true
	case <-time.After(time.Millisecond * time.Duration(delay)):
	}
	return false
}

type UserSim struct {
	username string
	userId   string
	client   *Client
	ri       runInfo
}

func NewUserSim(username string, ri runInfo) (*UserSim, error) {
	client := NewClient(ri.cfg.SiteURL)

	if _, err := ri.admin.CreateUser(username); err != nil {
		return nil, err
	}

	user, resp := client.Login(username, username)
	if !isSuccess(resp) {
		return nil, resp.Error
	}

	userSim := &UserSim{
		username: username,
		userId:   user.Id,
		client:   client,
		ri:       ri,
	}
	return userSim, nil
}

// Post creates a new post for the specified channel.
func (sim *UserSim) Post(text string, channelId string) {
	post := &model.Post{
		UserId:    sim.userId,
		ChannelId: channelId,
		Message:   text,
		Type:      model.POST_DEFAULT,
	}

	if _, resp := sim.client.CreatePost(post); !isSuccess(resp) {
		sim.ri.logger.Errorf("Cannot post to channel %s for user %s:", channelId, sim.username, resp.Error)
	}
}

func (sim *UserSim) Reply(post *model.Post, channelId string) error {
	// don't reply to a reply
	if post.RootId != "" {
		return nil
	}

	if !shouldDoIt(sim.ri.cfg.ProbProperty) {
		return nil
	}

	reply := &model.Post{
		UserId:    sim.userId,
		ChannelId: channelId,
		RootId:    post.Id,
		Message:   lorem(sim.ri.cfg),
	}
	if _, resp := sim.client.CreatePost(reply); !isSuccess(resp) {
		return resp.Error
	}
	return nil
}

func (sim *UserSim) React(post *model.Post, channelId string) error {
	if !shouldDoIt(sim.ri.cfg.ProbComment) {
		return nil
	}

	reaction := &model.Reaction{
		UserId:    sim.userId,
		PostId:    post.Id,
		EmojiName: randomEmoji(),
		CreateAt:  model.GetMillis(),
	}
	if _, resp := sim.client.SaveReaction(reaction); !isSuccess(resp) {
		return resp.Error
	}
	return nil
}

func (sim *UserSim) Edit(post *model.Post, channelId string) error {
	if !shouldDoIt(sim.ri.cfg.ProbEdit) {
		return nil
	}

	text := lorem(sim.ri.cfg)

	patch := &model.PostPatch{
		Message: &text,
	}

	if _, resp := sim.client.PatchPost(post.Id, patch); !isSuccess(resp) {
		return resp.Error
	}
	return nil
}

func (sim *UserSim) Delete(post *model.Post, channelId string) error {
	if !shouldDoIt(sim.ri.cfg.ProbDelete) {
		return nil
	}

	if _, resp := sim.client.DeletePost(post.Id); !isSuccess(resp) {
		return resp.Error
	}
	return nil
}
