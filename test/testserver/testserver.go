package testserver

import (
	"context"
	_ "embed"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/storage"
	tclientcore "github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/tclient"
)

//go:embed public_key.pem
var publicKeyData []byte

var (
	dc     = 1
	dcList = dcs.List{
		Options: []tg.DCOption{
			{
				ID:        1,
				IPAddress: "127.0.0.1",
				Port:      10443,
			},
		},
		Domains: nil,
		Test:    false,
	}
	publicKeys []exchange.PublicKey
	phone      = "+86 13858528382"
)

func init() {
	keys, _ := crypto.ParseRSAPublicKeys(publicKeyData)
	for _, k := range keys {
		publicKeys = append(publicKeys, exchange.PublicKey{RSA: k})
	}
}

// Setup creates test user and returns account and session file path. Namespace is the value of account.
func Setup(ctx context.Context, rnd rand.Source) (account string, sessionFile string, _ error) {
	tclientcore.DC = dc
	tclientcore.DCList = dcList
	tclientcore.PublicKeys = publicKeys

	dcpool.EnableTestMode()

	account = strconv.FormatInt(rand.Int63(), 10)
	sessionFile = filepath.Join(os.TempDir(), "tdl", account)

	return account, sessionFile, setupTestUser(ctx, rand.New(rnd), account, sessionFile)
}

func setupTestUser(ctx context.Context, rnd *rand.Rand, account, sessionFile string) error {
	kvd, err := kv.New(kv.DriverFile, map[string]any{
		"path": sessionFile,
	})
	if err != nil {
		return errors.Wrapf(err, "create kv storage: %s", sessionFile)
	}
	log.Printf("session file: %s", sessionFile)

	stg, err := kvd.Open(account)
	if err != nil {
		return errors.Wrap(err, "open test namespace")
	}

	sess := storage.NewSession(stg, true)

	opts := telegram.Options{
		DC:             dc,
		DCList:         dcList,
		PublicKeys:     publicKeys,
		SessionStorage: sess,
	}

	app := tclient.Apps[tclient.AppDesktop]
	c := telegram.NewClient(app.AppID, app.AppHash, opts)

	if err = c.Run(ctx, func(ctx context.Context) error {
		if err = c.Ping(ctx); err != nil {
			return err
		}

		authClient := auth.NewClient(c.API(), rnd, app.AppID, app.AppHash)

		if err = auth.NewFlow(
			testAuth{phone: phone},
			auth.SendCodeOptions{},
		).Run(ctx, authClient); err != nil {
			return errors.Wrap(err, "register test user")
		}

		user, err := c.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "get self")
		}

		log.Printf("user: %v, %v, %v", user.ID, user.FirstName, user.LastName)
		return nil
	}); err != nil {
		return errors.Wrap(err, "run auth")
	}

	return nil
}

type testAuth struct {
	phone string
}

func (t testAuth) Phone(_ context.Context) (string, error)    { return t.phone, nil }
func (t testAuth) Password(_ context.Context) (string, error) { return "", auth.ErrPasswordNotProvided }
func (t testAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return "12345", nil
}

func (t testAuth) AcceptTermsOfService(_ context.Context, _ tg.HelpTermsOfService) error {
	return nil
}

func (t testAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{
		FirstName: "Test",
		LastName:  "User",
	}, nil
}
