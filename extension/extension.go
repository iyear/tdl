package extension

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/go-faster/errors"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/core/util/logutil"
)

const EnvKey = "TDL_EXTENSION"

type Env struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	AppID     int    `json:"app_id"`
	AppHash   string `json:"app_hash"`
	Session   []byte `json:"session"`
	DataDir   string `json:"data_dir"`
	NTP       string `json:"ntp"`
	Proxy     string `json:"proxy"`
	Pool      int64  `json:"pool"`
	Debug     bool   `json:"debug"`
}

type Options struct {
	// UpdateHandler will be passed to telegram.Client Options.
	UpdateHandler telegram.UpdateHandler
	// Middlewares will be passed to telegram.Client Options,
	// and recovery,retry,flood-wait will be used if nil.
	Middlewares []telegram.Middleware
	// Logger will be used as extension logger,
	// and default logger(write to extension data dir) will be used if nil.
	Logger *zap.Logger
}

type Extension struct {
	name   string           // extension name
	client *telegram.Client // telegram client
	log    *zap.Logger      // logger
	config *Config          // extension config
}

type Config struct {
	Namespace string // tdl namespace
	DataDir   string // data directory for extension
	Proxy     string // proxy URL
	Pool      int64  // pool size
	Debug     bool   // debug mode enabled
}

func (e *Extension) Name() string {
	return e.name
}

func (e *Extension) Client() *telegram.Client {
	return e.client
}

func (e *Extension) Log() *zap.Logger {
	return e.log
}

func (e *Extension) Config() *Config {
	return e.config
}

type Handler func(ctx context.Context, e *Extension) error

func New(o Options) func(h Handler) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	ext, client, err := buildExtension(ctx, o)
	assert(err)

	return func(h Handler) {
		defer cancel()

		assert(tclient.RunWithAuth(ctx, client, func(ctx context.Context) error {
			if err := h(ctx, ext); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}

			return nil
		}))
	}
}

func buildExtension(ctx context.Context, o Options) (*Extension, *telegram.Client, error) {
	envFile := os.Getenv(EnvKey)
	if envFile == "" {
		return nil, nil, errors.New("please launch extension with `tdl EXTENSION_NAME`")
	}

	extEnv, err := os.ReadFile(envFile)
	if err != nil {
		return nil, nil, errors.Wrap(err, "read env file")
	}

	env := &Env{}
	if err = json.Unmarshal(extEnv, env); err != nil {
		return nil, nil, errors.Wrap(err, "unmarshal extension environment")
	}

	if o.Logger == nil {
		level := zap.InfoLevel
		if env.Debug {
			level = zap.DebugLevel
		}
		o.Logger = logutil.New(level, filepath.Join(env.DataDir, "log", "latest.log"))
	}

	// save logger to context
	ctx = logctx.With(ctx, o.Logger)

	if o.Middlewares == nil {
		o.Middlewares = tclient.NewDefaultMiddlewares(ctx, 0)
	}

	client, err := buildClient(ctx, env, o)
	if err != nil {
		return nil, nil, errors.Wrap(err, "build client")
	}

	return &Extension{
		name:   env.Name,
		client: client,
		log:    o.Logger,
		config: &Config{
			Namespace: env.Namespace,
			DataDir:   env.DataDir,
			Proxy:     env.Proxy,
			Pool:      env.Pool,
			Debug:     env.Debug,
		},
	}, client, nil
}

func buildClient(ctx context.Context, env *Env, o Options) (*telegram.Client, error) {
	storage := &session.StorageMemory{}
	if err := storage.StoreSession(ctx, env.Session); err != nil {
		return nil, errors.Wrap(err, "store session")
	}

	return tclient.New(ctx, tclient.Options{
		AppID:            env.AppID,
		AppHash:          env.AppHash,
		Session:          storage,
		Middlewares:      o.Middlewares,
		Proxy:            env.Proxy,
		NTP:              env.NTP,
		ReconnectTimeout: 0, // no timeout
		UpdateHandler:    o.UpdateHandler,
	})
}

func assert(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
