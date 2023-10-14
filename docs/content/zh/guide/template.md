---
title: "模板指南"
bookHidden: true
bookToC: false
---

# 模板指南

本指南将介绍可用于 tdl 模板中的变量和函数。

模板语法基于[Go text/template](https://golang.org/pkg/text/template/)。

## 下载

### 变量 (Beta)

|       变量       |         描述         |
|:--------------:|:------------------:|
|   `DialogID`   |   Telegram 对话ID    |
|  `MessageID`   |   Telegram 消息ID    |
| `MessageDate`  | Telegram 消息日期（时间戳） |
|   `FileName`   |    Telegram 文件名    |
|   `FileSize`   |  可读的文件大小，例如 `1GB`  |
| `DownloadDate` |     下载日期（时间戳）      |

### 函数 (Beta)

|      函数      |                                             描述                                             |                              用法                              |                                          示例                                           |
|:------------:|:------------------------------------------------------------------------------------------:|:------------------------------------------------------------:|:-------------------------------------------------------------------------------------:|
|   `repeat`   |                                     重复 `STRING` `N` 次                                      |                      `repeat STRING N`                       |                                `{{ repeat "test" 3 }}`                                |
|  `replace`   |                                  对 `STRING` 执行 `PAIRS` 替换                                  |                  `replace STRING PAIRS...`                   |                        `{{ replace "Test" "t" "T" "e" "E" }}`                         |
|   `upper`    |                                      将 `STRING` 转换为大写                                      |                        `upper STRING`                        |                                 `{{ upper "Test" }}`                                  |
|   `lower`    |                                      将 `STRING` 转换为小写                                      |                        `lower STRING`                        |                                 `{{ lower "Test" }}`                                  |
| `snakecase`  |                                 将 `STRING` 转换为 snake_case                                  |                      `snakecase STRING`                      |                               `{{ snakecase "Test" }}`                                |
| `camelcase`  |                                  将 `STRING` 转换为 camelCase                                  |                      `camelcase STRING`                      |                               `{{ camelcase "Test" }}`                                |
| `kebabcase`  |                                 将 `STRING` 转换为 kebab-case                                  |                      `kebabcase STRING`                      |                               `{{ kebabcase "Test" }}`                                |
|    `rand`    |                                  在范围 `MIN` 到 `MAX` 生成随机数                                   |                        `rand MIN MAX`                        |                                   `{{ rand 1 10 }}`                                   |
|    `now`     |                                          获取当前时间戳                                           |                            `now`                             |                                      `{{ now }}`                                      |
| `formatDate` | [格式化](https://zhuanlan.zhihu.com/p/145009400) `TIMESTAMP` 时间戳<br/>(默认格式: `20060102150405`) | `formatDate TIMESTAMP` <br/> `formatDate TIMESTAMP "format"` | `{{ formatDate 1600000000 }}`<br/> `{{ formatDate 1600000000 "2006-01-02-15-04-05"}}` |

### 示例：

```gotemplate
{{ .DialogID }}_{{ .MessageID }}_{{ replace .FileName `/` `_` `\` `_` `:` `_` `*` `_` `?` `_` `<` `_` `>` `_` `|` `_` ` ` `_`  }}

{{ .FileName }}_{{ formatDate .DownloadDate }}_{{ .FileSize }}

{{ .FileName }}_{{ formatDate .DownloadDate "2006-01-02-15-04-05"}}_{{ .FileSize }}

{{ lower (replace .FileName ` ` `_`) }}

{{ formatDate (now) }}
```

### 默认：

```gotemplate
{{ .DialogID }}_{{ .MessageID }}_{{ replace .FileName `/` `_` `\` `_` `:` `_` `*` `_` `?` `_` `<` `_` `>` `_` `|` `_` ` ` `_`  }}
```
