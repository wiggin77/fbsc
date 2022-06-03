package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
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

func wait(delay int64, done chan struct{}) bool {
	select {
	case <-done:
		return true
	case <-time.After(time.Millisecond * time.Duration(delay)):
	}
	return false
}

func reverseString(s string) string {
	chars := []rune(s)
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}
