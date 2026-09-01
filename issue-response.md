Thanks for the detailed report!

I found and fixed two bugs causing messages to be missed during export.

### Bug 1: Premature pagination stop in time-based export (main issue)

When using `-T time`, `OffsetDate` was set for the Telegram API while `OffsetID` defaulted to 0. The gotd iterator's `lastBatch` heuristic (`len(msgs) < limit`) assumes that if a batch is smaller than the limit, there are no more messages. However, the Telegram server may apply `offset_date` as a post-filter **after** collecting each batch. If the filter removes messages, the batch becomes smaller than the limit, causing pagination to stop prematurely — even though thousands of messages remain.

**Fix**: Remove `OffsetDate` from time-based export entirely. Use `OffsetID` for server-side pagination (same approach as the working ID-based export) and apply date filtering client-side with upper and lower bound checks. This trades a small efficiency cost for correct results.

### Bug 2: Integer overflow in no-range mode

When no `-i` flag is specified, `Input[1]` defaults to `math.MaxInt`. The expression `Input[1] + 1` overflows to `math.MinInt` (-9223372036854775808), sending a negative offset to the Telegram API.

**Fix**: Skip setting the offset when `Input[1]` equals `math.MaxInt`.

### About `-T id -i 0,6979` returning 0

This test did not include `--all`, while the working test (`-T id -i 0,5979`) did. Without `--all`, only media messages are exported. Adding `--all` should resolve this.

### How to test

```bash
git fetch https://github.com/lenny-ts/tdl.git fix/export-overflow-offset
git checkout FETCH_HEAD
go build -o tdl.exe .
```

The fix will be included in the next release once merged.
