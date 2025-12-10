package doctor

import (
	"context"
	"runtime"
	"time"

	"github.com/beevik/ntp"
	"github.com/fatih/color"
	"github.com/spf13/viper"

	"github.com/iyear/tdl/pkg/consts"
)

func getDefaultNTPServers() []string {
	switch runtime.GOOS {
	case "darwin": // macOS
		return []string{
			"time.apple.com",
			"time.cloudflare.com",
			"time.windows.com",
			"time.google.com",
			"pool.ntp.org",
		}
	case "windows":
		return []string{
			"time.windows.com",
			"time.cloudflare.com",
			"time.apple.com",
			"time.google.com",
			"pool.ntp.org",
		}
	case "linux":
		return []string{
			"ntp.ubuntu.com",
			"time.cloudflare.com",
			"time.google.com",
			"pool.ntp.org",
		}
	default:
		return []string{
			"time.cloudflare.com",
			"time.google.com",
			"pool.ntp.org",
		}
	}
}

func checkNTPTime(ctx context.Context, opts Options) {
	var ntpServers []string

	ntpServer := viper.GetString(consts.FlagNTP)
	if ntpServer != "" {
		ntpServers = []string{ntpServer}
	} else {
		ntpServers = getDefaultNTPServers()
	}

	var resp *ntp.Response
	var err error
	var successServer string

	for _, server := range ntpServers {
		color.White("  Querying NTP server: %s", server)
		resp, err = ntp.Query(server)
		if err != nil {
			color.Yellow("[WARN] Failed to query %s: %v", server, err)
			continue
		}
		successServer = server
		break
	}

	if err != nil {
		color.Red("[FAIL] Failed to query all NTP servers")
		return
	}

	color.White("  Successfully connected to: %s", successServer)

	offset := resp.ClockOffset
	localTime := time.Now()
	ntpTime := localTime.Add(offset)

	color.White("  Local time: %s", localTime.Format("2006-01-02 15:04:05.000000000 MST"))
	color.White("  NTP time:   %s", ntpTime.Format("2006-01-02 15:04:05.000000000 MST"))
	color.White("  Offset:     %v", offset)

	absOffset := offset
	if absOffset < 0 {
		absOffset = -absOffset
	}

	if absOffset > time.Second && absOffset < 10*time.Second {
		color.Yellow("[WARN] Time offset is between 1 and 10 seconds")
		color.White("  Your system time is slightly off, but should work normally")
		color.White("  Consider synchronizing your system time for better accuracy")
	} else if absOffset > 10*time.Second {
		color.Yellow("[WARN] Time offset is greater than 10 seconds")
		color.Yellow("  This may cause issues with Telegram authentication")
		color.Yellow("  Consider synchronizing your system time or using --ntp flag")
	} else {
		color.Green("[OK] Time synchronization is acceptable")
	}
}
