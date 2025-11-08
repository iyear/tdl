package doctor

import (
	"context"

	"github.com/fatih/color"
	"github.com/gotd/td/telegram"

	"github.com/iyear/tdl/pkg/kv"
)

type Options struct {
	KV     kv.Storage
	Client *telegram.Client
}

// CheckFunc represents a diagnostic check function
type CheckFunc func(ctx context.Context, opts Options)

// Check represents a registered diagnostic check
type Check struct {
	Name string
	Fn   CheckFunc
}

var (
	checks = make([]Check, 0)
)

// Register registers a new diagnostic check
func Register(name string, fn CheckFunc) {
	checks = append(checks, Check{
		Name: name,
		Fn:   fn,
	})
}

// Run executes all registered diagnostic checks
func Run(ctx context.Context, opts Options) error {
	color.Blue("=== TDL Doctor ===\n")

	// Separate checks into client-dependent and client-independent
	var clientIndependent []Check
	var clientDependent []Check

	for _, check := range checks {
		if check.Name == "Checking database integrity" || check.Name == "Checking time synchronization" {
			clientIndependent = append(clientIndependent, check)
		} else {
			clientDependent = append(clientDependent, check)
		}
	}

	// Run client-independent checks first
	total := len(checks)
	currentIndex := 0
	for _, check := range clientIndependent {
		currentIndex++
		color.Cyan("\n[%d/%d] %s...", currentIndex, total, check.Name)
		check.Fn(ctx, opts)
	}

	// Run client-dependent checks within a single client.Run()
	if len(clientDependent) > 0 && opts.Client != nil {
		err := opts.Client.Run(ctx, func(ctx context.Context) error {
			for _, check := range clientDependent {
				currentIndex++
				color.Cyan("\n[%d/%d] %s...", currentIndex, total, check.Name)
				check.Fn(ctx, opts)
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
			color.Cyan("\n[%d/%d] %s...", currentIndex, total, check.Name)
			check.Fn(ctx, opts)
		}
	}

	color.Green("\n=== Diagnosis Complete ===")
	return nil
}

// init registers all checks in order
func init() {
	Register("Checking database integrity", checkDatabaseIntegrity)
	Register("Checking Telegram server connectivity", checkConnectivity)
	Register("Checking time synchronization", checkNTPTime)
	Register("Checking login status", checkLoginStatus)
}
