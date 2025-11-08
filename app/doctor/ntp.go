package doctor

import (
	"context"
	"time"

	"github.com/beevik/ntp"
	"github.com/fatih/color"
	"github.com/spf13/viper"

	"github.com/iyear/tdl/pkg/consts"
)

func checkNTPTime(ctx context.Context, opts Options) {
	ntpServer := viper.GetString(consts.FlagNTP)
	if ntpServer == "" {
		ntpServer = "pool.ntp.org"
	}

	// Get NTP time
	resp, err := ntp.Query(ntpServer)
	if err != nil {
		color.Red("  [FAIL] Failed to query NTP server (%s): %v", ntpServer, err)
		return
	}

	offset := resp.ClockOffset
	localTime := time.Now()
	ntpTime := localTime.Add(offset)

	color.White("  Local time: %s", localTime.Format("2006-01-02 15:04:05 MST"))
	color.White("  NTP time:   %s", ntpTime.Format("2006-01-02 15:04:05 MST"))
	color.White("  Offset:     %v", offset)

	// Warning if offset is too large
	absOffset := offset
	if absOffset < 0 {
		absOffset = -absOffset
	}

	if absOffset > 30*time.Second {
		color.Yellow("  [WARN] Time offset is greater than 30 seconds. This may cause issues with Telegram.")
		color.Yellow("  [WARN] Consider synchronizing your system time or using --ntp flag.")
	} else {
		color.Green("  [OK] Time synchronization is acceptable")
	}
}
