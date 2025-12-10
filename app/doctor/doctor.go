package doctor

import (
	"context"

	"github.com/fatih/color"
	"github.com/gotd/td/telegram"

	"github.com/iyear/tdl/pkg/kv"
)

const (
	CheckNameTimeSync      = "Checking time synchronization"
	CheckNameConnectivity  = "Checking Telegram server connectivity"
	CheckNameDatabaseInteg = "Checking database integrity"
	CheckNameLoginStatus   = "Checking login status"
)

// init registers all checks in order
func init() {
	Register(newCheck(CheckNameTimeSync, false, checkNTPTime))
	Register(newCheck(CheckNameConnectivity, true, checkConnectivity))
	Register(newCheck(CheckNameDatabaseInteg, false, checkDatabaseIntegrity))
	Register(newCheck(CheckNameLoginStatus, true, checkLoginStatus))
}

type Options struct {
	KV     kv.Storage
	Client *telegram.Client
}

type Checker interface {
	Name() string
	NeedClient() bool
	Run(ctx context.Context, opts Options)
}

var checks = make([]Checker, 0)

func Register(checker Checker) {
	checks = append(checks, checker)
}

func Run(ctx context.Context, opts Options) error {
	color.Blue("=== TDL Doctor ===\n")

	// Separate checks into client-dependent and client-independent
	var clientIndependent []Checker
	var clientDependent []Checker

	for _, check := range checks {
		if check.NeedClient() {
			clientDependent = append(clientDependent, check)
		} else {
			clientIndependent = append(clientIndependent, check)
		}
	}

	// Run client-independent checks first
	total := len(checks)
	currentIndex := 0
	for _, check := range clientIndependent {
		currentIndex++
		color.Cyan("\n[%d/%d] %s...", currentIndex, total, check.Name())
		check.Run(ctx, opts)
	}

	// Run client-dependent checks within a single client.Run()
	if len(clientDependent) > 0 && opts.Client != nil {
		err := opts.Client.Run(ctx, func(ctx context.Context) error {
			for _, check := range clientDependent {
				currentIndex++
				color.Cyan("\n[%d/%d] %s...", currentIndex, total, check.Name())
				check.Run(ctx, opts)
			}
			return nil
		})
		if err != nil {
			color.Red("\n[FAIL] Client error: %v", err)
		}
	} else {
		// Run checks without client
		for _, check := range clientDependent {
			currentIndex++
			color.Cyan("\n[%d/%d] %s...", currentIndex, total, check.Name())
			check.Run(ctx, opts)
		}
	}

	color.Blue("\n=== Diagnosis Complete ===")
	return nil
}

type checkImpl struct {
	name       string
	needClient bool
	runFunc    func(ctx context.Context, opts Options)
}

func (c *checkImpl) Name() string {
	return c.name
}

func (c *checkImpl) NeedClient() bool {
	return c.needClient
}

func (c *checkImpl) Run(ctx context.Context, opts Options) {
	c.runFunc(ctx, opts)
}

func newCheck(name string, needClient bool, runFunc func(ctx context.Context, opts Options)) Checker {
	return &checkImpl{
		name:       name,
		needClient: needClient,
		runFunc:    runFunc,
	}
}
