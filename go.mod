module github.com/iyear/tdl

go 1.21

replace (
	github.com/iyear/tdl/core => ./core
	github.com/iyear/tdl/extension => ./extension
)

require (
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/bcicen/jstream v1.0.1
	github.com/beevik/ntp v1.4.3
	github.com/expr-lang/expr v1.16.9
	github.com/fatih/color v1.18.0
	github.com/flytam/filenamify v1.2.0
	github.com/gabriel-vasile/mimetype v1.4.8
	github.com/go-faster/errors v0.7.1
	github.com/go-faster/jx v1.1.0
	github.com/go-playground/validator/v10 v10.24.0
	github.com/google/go-github/v62 v62.0.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gotd/contrib v0.20.0
	github.com/gotd/td v0.110.1
	github.com/iancoleman/strcase v0.3.0
	github.com/ivanpirog/coloredcobra v1.0.1
	github.com/iyear/tdl/core v0.18.5
	github.com/iyear/tdl/extension v0.18.5
	github.com/jedib0t/go-pretty/v6 v6.5.0
	github.com/klauspost/compress v1.17.11
	github.com/kopoli/go-terminal-size v0.0.0-20170219200355-5c97524c8b54
	github.com/mattn/go-runewidth v0.0.16
	github.com/mitchellh/mapstructure v1.5.0
	github.com/onsi/ginkgo/v2 v2.20.1
	github.com/onsi/gomega v1.34.1
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/spf13/cast v1.7.1
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.10.0
	github.com/tidwall/gjson v1.18.0
	go.etcd.io/bbolt v1.3.10
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.34.0
	golang.org/x/time v0.9.0
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-faster/xor v1.0.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/pprof v0.0.0-20240727154555-813a5fbdbec8 // indirect
	github.com/gotd/ige v0.2.2 // indirect
	github.com/gotd/neo v0.1.5 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/iyear/connectproxy v0.1.1 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20230326075908-cb1d2100619a // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yapingcat/gomedia v0.0.0-20240601043430-920523f8e5c7 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.11 // indirect
	rsc.io/qr v0.2.0 // indirect
)
