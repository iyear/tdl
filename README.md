# tdl

<img align="right" src="docs/assets/img/logo.png" height="280" alt="">

> 📥 Telegram Downloader, but more than a downloader

English | <a href="README_zh.md">简体中文</a>

<p>
<img src="https://img.shields.io/github/go-mod/go-version/iyear/tdl?style=flat-square" alt="">
<img src="https://img.shields.io/github/license/iyear/tdl?style=flat-square" alt="">
<img src="https://img.shields.io/github/actions/workflow/status/iyear/tdl/master.yml?branch=master&amp;style=flat-square" alt="">
<img src="https://img.shields.io/github/v/release/iyear/tdl?color=red&amp;style=flat-square" alt="">
<img src="https://img.shields.io/github/downloads/iyear/tdl/total?style=flat-square" alt="">
</p>

#### Features:
- Single file start-up
- Low resource usage
- Take up all your bandwidth
- Faster than official clients
- Download files from (protected) chats
- Forward messages with automatic fallback and message routing
- Upload files to Telegram, including streamable video and grouped albums
- Export messages/members/subscribers to JSON

## Preview

It reaches my proxy's speed limit, and the **speed depends on whether you are a premium**

![](docs/assets/img/preview.gif)

## Uploading media and albums

The `upload` command has two extended pipelines beyond plain document upload:

- `--video`: send `video/*` files as streamable inline video, extracting duration, resolution, and a thumbnail via `ffprobe`/`ffmpeg`. Combine with `--detect-video` to force `mp4`/`m4v`/`mov` files through this path regardless of MIME sniffing.
- `--album`: group items whose base name matches `<model>_<YYYY-MM-DD>_<HH-MM-SS>` and send them as Telegram grouped media (up to 10 items per album). Non-matching files remain individual uploads. Files inside an album upload in parallel; each item is committed via `messages.uploadMedia` before `messages.sendMultiMedia` posts the group.

Requirements and behavior:

- `ffprobe` and `ffmpeg` on `PATH` when `--video` or `--detect-video` is active.
- Probed metadata and generated thumbnails are cached under `.cache/metadata/` and `.cache/thumbs/`, keyed by the source file's SHA-256. A sidecar thumbnail (`<video>.jpg`, `.jpeg`, or `.png` next to the video) is picked up automatically; otherwise a 320×320 JPEG ≤ 200 KB is generated.
- Captions follow the existing `--caption` expression. Inside an album only the first item carries the caption, matching Telegram grouped-media semantics.

## Documentation

Please refer to the [documentation](https://docs.iyear.me/tdl/).

## Sponsors

![](https://raw.githubusercontent.com/iyear/sponsor/master/sponsors.svg)

## Contributors
<a href="https://github.com/iyear/tdl/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=iyear/tdl&max=750&columns=20" alt="contributors"/>
</a>

## LICENSE

AGPL-3.0 License
