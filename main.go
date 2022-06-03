package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/mattermost/logr/v2"
)

const (
	DefaultUserCount        = 5
	DefaultChannelsPerUser  = 3
	DefaultBoardsPerChannel = 5
	DefaultCardsPerBoard    = 20

	DefaultMaxWordsPerSentence      = 30
	DefaultMaxSentencesPerParagraph = 5
	DefaultMaxParagraphsPerComment  = 2

	DefaultBoardDelay = 10
	DefaultCardDelay  = 10

	FilePerms = 0664
)

func main() {
	var exitCode int
	var configFile string
	var logConfigFile string
	var createConfig bool
	var quiet bool
	var help bool
	flag.StringVar(&configFile, "f", "", "config file")
	flag.BoolVar(&createConfig, "c", false, "creates a default config file")
	flag.StringVar(&logConfigFile, "log", "", "specifies a custom logr config")
	flag.BoolVar(&quiet, "q", false, "suppress output")

	flag.BoolVar(&help, "h", false, "displays this help text")
	flag.Parse()

	defer func(code *int) { os.Exit(*code) }(&exitCode)

	lgr, _ := logr.New()
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
			logger.Error("Cannot create default config", logr.Err(err))
			exitCode = 1
		}
		return
	}

	cfg, err := loadConfig(configFile)
	if err != nil {
		logger.Error("Cannot load config", logr.Err(err))
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
		logger.Error("Cannot create admin client", logr.Err(err))
		exitCode = 3
		return
	}

	if err := setUpServer(admin, cfg); err != nil {
		logger.Error("Cannot setup server", logr.Err(err))
		exitCode = 4
		return
	}

	ri := &runInfo{
		cfg:    cfg,
		logger: logger,
		done:   done,
		admin:  admin,
		quiet:  quiet,
	}

	start := time.Now()

	run(ri)

	blockCount := atomic.LoadInt64(&ri.blockCount)
	duration := time.Since(start)

	fmt.Print("\n" + ri.output.String())
	fmt.Printf("Duration: %s\n", duration.Round(time.Millisecond))

	blocksPerSecond := float64(blockCount) / duration.Seconds()
	fmt.Printf("Blocks Per Second: %.2f\n", blocksPerSecond)
}

func run(ri *runInfo) {
	var wg sync.WaitGroup
	for i := 0; i < ri.cfg.UserCount; i++ {
		wg.Add(1)

		username := makeName(".")

		go func(u string) {
			defer wg.Done()
			stats, err := runUser(u, ri)
			if err != nil {
				ri.logger.Error("Cannot simulate user", logr.Err(err))
			}

			if !ri.quiet {
				s := fmt.Sprintf("%s: channels=%d  boards=%d  cards=%d  text=%d\n",
					username, stats.ChannelCount, stats.BoardCount, stats.CardCount, stats.TextCount)

				ri.output.Write(s)
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

// setUpServer creates the team.
func setUpServer(admin *AdminClient, cfg *Config) error {
	team, err := admin.CreateTeam(cfg.TeamName, true)
	if err != nil {
		return err
	}
	cfg.setTeamID(team.Id)
	return nil
}
