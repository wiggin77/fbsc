package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mattermost/logr"
)

const (
	DefaultTeamName                 = "TestTeam"
	DefaultUserCount                = 5
	DefaultAvgPostDelay             = 15000
	DefaultProbReact                = 0.25
	DefaultProbReply                = 0.10
	DefaultMaxWordsPerSentence      = 100
	DefaultMaxSentencesPerParagraph = 20
	DefaultMaxParagraphsPerComment  = 2

	FilePerms = 0664
)

func main() {
	var exitCode int
	var configFile string
	var logConfigFile string
	var createConfig bool
	var help bool
	flag.StringVar(&configFile, "f", "", "config file")
	flag.BoolVar(&createConfig, "c", false, "creates a default config file")
	flag.StringVar(&logConfigFile, "log", "", "specifies a custom logr config")

	flag.BoolVar(&help, "h", false, "displays this help text")
	flag.Parse()

	defer func(code *int) { os.Exit(*code) }(&exitCode)

	lgr := &logr.Logr{}
	logger := lgr.NewLogger()
	if err := initLogging(lgr, logConfigFile); err != nil {
		exitCode = 1
		return
	}

	defer func(l *logr.Logr) {
		if lgr.IsShutdown() {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		if err := lgr.ShutdownWithTimeout(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}(lgr)

	if help {
		flag.PrintDefaults()
		return
	}

	if configFile == "" {
		logger.Error("You must specify a config file")
		flag.PrintDefaults()
		exitCode = 1
		return
	}

	if createConfig {
		if err := createDefaultConfig(configFile); err != nil {
			logger.Error(err)
			exitCode = 1
		}
		return
	}

	cfg, err := loadConfig(configFile)
	if err != nil {
		logger.Error(err)
		exitCode = 1
		return
	}

	done := make(chan struct{})

	setUpInterruptHandler(func() {
		close(done)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_ = lgr.ShutdownWithTimeout(ctx)
		// give the workers a chance to shut down gracefully (some may not, that's ok).
		time.Sleep(time.Second * 2)
	})

	admin, err := NewAdminClient(cfg)
	if err != nil {
		logger.Error(err)
		exitCode = 3
		return
	}

	if err := setUpServer(admin, cfg); err != nil {
		logger.Error(err)
		exitCode = 4
		return
	}

	ri := runInfo{
		cfg:    cfg,
		logger: logger,
		done:   done,
		admin:  admin,
	}
	run(ri)
}

func run(ri runInfo) {
	var wg sync.WaitGroup

	usernames := make([]string, 0, len(ri.cfg.Usernames)+ri.cfg.UserCount)
	usernames = append(usernames, ri.cfg.Usernames...)

	for i := 0; i < ri.cfg.UserCount; i++ {
		usernames = append(usernames, makeName("."))
	}

	for _, username := range usernames {
		wg.Add(1)

		go func(u string) {
			defer wg.Done()
			if err := runUser(u, ri); err != nil {
				ri.logger.Error(err)
			}
		}(username)
	}

	wg.Wait()
}

func setUpInterruptHandler(cleanUp func()) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Println("  user abort; exiting...")

		if cleanUp != nil {
			cleanUp()
		}
		os.Exit(0)
	}()
}

// setUpServer creates the workspaces and boards listed in config.
func setUpServer(admin *AdminClient, cfg *Config) error {
	/*
		if len(cfg.Workspaces) == 0 {
			return errors.New("At least one channel name must specified in config.")
		}

		team, err := admin.CreateTeam(cfg.TeamName, true)
		if err != nil {
			return err
		}
		cfg.TeamId = team.Id

		for _, channelName := range cfg.Workspaces {
			channel, err := admin.CreateChannel(channelName, team.Id)
			if err != nil {
				return err
			}
			cfg.ChannelIds = append(cfg.ChannelIds, channel.Id)
		}
	*/
	return nil
}
