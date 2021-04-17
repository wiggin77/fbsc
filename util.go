package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-server/model"
)

var rnd *rand.Rand

func init() {
	s1 := rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(s1)
}

func pickRandomString(arr []string) string {
	return arr[rnd.Intn(len(arr))]
}

func pickRandomInt(min int, max int) int {
	if min > max {
		return max
	}
	return rnd.Intn(max-min) + min
}

func shouldDoIt(probability float32) bool {
	return rnd.Float32() <= probability
}

func randomDuration(avgDurationMillis int64, variance float32) int64 {
	if variance < 0 {
		variance = 0
	}
	if variance >= 1.0 {
		variance = 0.99
	}

	if avgDurationMillis <= 0 {
		avgDurationMillis = 1000
	}

	if avgDurationMillis < 100 {
		return avgDurationMillis
	}

	delta := int64(float32(avgDurationMillis) * variance)
	return avgDurationMillis + rnd.Int63n(delta) - rnd.Int63n(delta)
}

func isSuccess(resp *model.Response) bool {
	if resp == nil {
		return false
	}
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return true
	}
	return false
}

func getPostsForChannelAroundLastUnread(client *model.Client4, userId string, channelId string) (*model.PostList, error) {
	path := fmt.Sprintf("/users/%s/channels/%s/posts/unread", userId, channelId)
	query := fmt.Sprintf("?limit_before=%d&limit_after=%d", 1, 60)

	r, err := client.DoApiGet(path+query, "")

	if err != nil {
		return nil, err
	}
	defer closeBody(r)

	return model.PostListFromJson(r.Body), nil
}

func closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
	}
}
