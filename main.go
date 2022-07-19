package main

import (
	"context"
	"fbsc/termprinter"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/mattermost/logr/v2"
)

const (
	DefaultConcurrentUsers  = 3
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
	var createConfig bool
	var quiet bool
	var help bool
	flag.StringVar(&configFile, "f", "", "config file")
	flag.BoolVar(&createConfig, "c", false, "creates a default config file")
	flag.BoolVar(&quiet, "q", false, "suppress output")

	flag.BoolVar(&help, "h", false, "displays this help text")
	flag.Parse()

	defer func(code *int) { os.Exit(*code) }(&exitCode)

	printer := termprinter.NewPrinter()

	lgr, _ := logr.New()
	logger := lgr.NewLogger()
	if err := initLogging(lgr, printer); err != nil {
		exitCode = 1
		return
	}

	defer func(l *logr.Logr) {
		if l.IsShutdown() {
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

	abort := make(chan struct{})
	workersExited := make(chan struct{})

	setUpInterruptHandler(func() {
		close(abort)

		// give the workers a chance to shut down gracefully (some may not, that's ok).
		select {
		case <-workersExited:
		case <-time.After(time.Second * 15):
		}

		// shutdown the logger
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_ = lgr.ShutdownWithTimeout(ctx)
	})

	ri := &runInfo{
		cfg:     cfg,
		printer: printer,
		logger:  logger,
		abort:   abort,
		admin:   admin,
		quiet:   quiet,
		stats:   &stats{},
	}
	start := time.Now()

	run(ri, workersExited)

	blockCount := atomic.LoadInt64(&ri.blockCount)
	duration := time.Since(start)

	printer.Printf("Duration: %s\n", duration.Round(time.Millisecond))

	blocksPerSecond := float64(blockCount) / duration.Seconds()
	printer.Printf("Blocks Per Second: %.2f\n", blocksPerSecond)
}

func run(ri *runInfo, workersExited chan struct{}) {
	defer close(workersExited)
	var wg sync.WaitGroup

	var usersLeft int32 = int32(ri.cfg.UserCount)
	concurrency := ri.cfg.ConcurrentUsers

	if !ri.quiet {
		ri.printer.Printf("Creating %d users with %d concurrent threads.\n\n", usersLeft, concurrency)
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runConcurrentUsers(ri, &usersLeft)
		}()
	}

	wg.Wait()
}

func runConcurrentUsers(ri *runInfo, usersLeft *int32) {
	for {
		select {
		case <-ri.abort:
			return
		default:
		}

		if atomic.AddInt32(usersLeft, -1) <= 0 {
			return
		}

		username := strings.ToLower(makeName("."))

		stats, err := runUser(username, ri)
		if err != nil {
			ri.logger.Error("Cannot simulate user", logr.String("username", username), logr.Err(err))
		}

		ri.AddStats(stats)
	}
}

func setUpInterruptHandler(cleanUp func()) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Println("\n  user abort; exiting...")

		if cleanUp != nil {
			cleanUp()
		}
		os.Exit(1)
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
