---
title: "Troubleshooting"
weight: 40
---

# Troubleshooting

## Best Practices

How to minimize the risk of blocking?

- Login with the official client session.
- Use the default download and upload options as possible. Do not set too large `threads` and `size`.
- Do not use the same account to login on multiple devices at the same time.
- Don't download or upload too many files at once.
- Become a Telegram premium user. 😅

## FAQ

#### Q: Why no response after entering the command? And why there is `msg_id too high` in the log?

**A:** Check if you need to use a proxy (use `proxy` flag); Check if your system's local time is correct (use `ntp` flag
or calibrate system time)

If that doesn't work, run again with `--debug` flag. Then file a new issue and paste your log in the issue.

#### Q: Desktop client stop working after using tdl?

**A:** If your desktop client can't receive messages, load chats, or send messages, you may encounter session conflicts.

You can try re-login with `tdl login` and **select YES for logout**, which will delete the session files to separate
sessions.

#### Q: How to migrate session to another device?

**A:** You can use the `tdl backup` and `tdl recover` commands to export and import sessions.
See [Migration](/guide/migration) for more details.

#### Q: Is this a form of abuse?

**A:** No. The download and upload speed is limited by the server side. Since the speed of official clients usually does
not
reach the account limit, this tool was developed to download files at the highest possible speed.

#### Q: Will this result in a ban?

**A:** I am not sure. All operations do not involve dangerous actions such as actively sending messages to other people.
But
it's safer to use a long-term account.
