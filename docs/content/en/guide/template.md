---
title: "Template Guide"
bookHidden: true
bookToC: false
---

# Template Guide

This guide is intended to introduce variables and functions that are available in the tdl template.

Template syntax is based on [Go's text/template](https://golang.org/pkg/text/template/) package.

## Download

### Variables (beta)

|      Var       |                 Desc                 |
|:--------------:|:------------------------------------:|
|   `DialogID`   |          Telegram dialog id          |
|  `MessageID`   |         Telegram message id          |
| `MessageDate`  |   Telegram message date(timestamp)   |
|   `FileName`   |          Telegram file name          |
|   `FileSize`   | Human-readable file size, like `1GB` |
| `DownloadDate` |       Download date(timestamp)       |

### Functions (beta)

|     Func     |                                                           Desc                                                           |                                             Usage                                              |                                        Example                                        |
|:------------:|:------------------------------------------------------------------------------------------------------------------------:|:----------------------------------------------------------------------------------------------:|:-------------------------------------------------------------------------------------:|
|   `repeat`   |                                                Repeat `STRING` `N` times                                                 |                                       `repeat STRING N`                                        |                                `{{ repeat "test" 3 }}`                                |
|  `replace`   |                                       Perform replacement on `STRING` with `PAIRS`                                       |                                   `replace STRING PAIRS...`                                    |                        `{{ replace "Test" "t" "T" "e" "E" }}`                         |
|   `upper`    |                                              Convert `STRING` to uppercase                                               |                                         `upper STRING`                                         |                                 `{{ upper "Test" }}`                                  |
|   `lower`    |                                              Convert `STRING` to lowercase                                               |                                         `lower STRING`                                         |                                 `{{ lower "Test" }}`                                  |
| `snakecase`  |                                              Convert `STRING` to snake_case                                              |                                       `snakecase STRING`                                       |                               `{{ snakecase "Test" }}`                                |
| `camelcase`  |                                              Convert `STRING` to camelCase                                               |                                       `camelcase STRING`                                       |                               `{{ camelcase "Test" }}`                                |
| `kebabcase`  |                                              Convert `STRING` to kebab-case                                              |                                       `kebabcase STRING`                                       |                               `{{ kebabcase "Test" }}`                                |
|    `rand`    |                                      Generate random number in range `MIN` to `MAX`                                      |                                         `rand MIN MAX`                                         |                                   `{{ rand 1 10 }}`                                   |
|    `now`     |                                                  Get current timestamp                                                   |                                             `now`                                              |                                      `{{ now }}`                                      |
| `formatDate` | Format `TIMESTAMP` with [format](https://golang.cafe/blog/golang-time-format-example.html)<br/>Default: `20060102150405` | `formatDate TIMESTAMP` <br/> `formatDate TIMESTAMP "format"` | `{{ formatDate 1600000000 }}`<br/> `{{ formatDate 1600000000 "2006-01-02-15-04-05"}}` |

### Examples:

```gotemplate
{{ .DialogID }}_{{ .MessageID }}_{{ replace .FileName `/` `_` `\` `_` `:` `_` `*` `_` `?` `_` `<` `_` `>` `_` `|` `_` ` ` `_`  }}

{{ .FileName }}_{{ formatDate .DownloadDate }}_{{ .FileSize }}

{{ .FileName }}_{{ formatDate .DownloadDate "2006-01-02-15-04-05"}}_{{ .FileSize }}

{{ lower (replace .FileName ` ` `_`) }}

{{ formatDate (now) }}
```

### Default:

```gotemplate
{{ .DialogID }}_{{ .MessageID }}_{{ replace .FileName `/` `_` `\` `_` `:` `_` `*` `_` `?` `_` `<` `_` `>` `_` `|` `_` ` ` `_`  }}
```
