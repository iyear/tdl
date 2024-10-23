package test

import (
	"context"
	"crypto/rand"
	_ "embed"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/core/dcpool"
	tclientcore "github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/tclient"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	tcmd "github.com/iyear/tdl/cmd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed public_key.pem
var publicKeyData []byte

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test tdl")
}

func TestSetupTestUser(t *testing.T) {
	dcpool.TestMode = true

	testAccount = strconv.FormatInt(time.Now().UnixNano(), 10)
	sessionFile = filepath.Join(os.TempDir(), "tdl", testAccount)

	ctx := context.Background()
	err := setupTestUser(ctx)
	if err != nil {
		t.Fatalf("setupTestUser failed: %v", err)
	}
}

var (
	cmd         *cobra.Command
	args        []string
	output      string
	testAccount string
	sessionFile string
)

var _ = BeforeSuite(func() {
	testAccount = strconv.FormatInt(time.Now().UnixNano(), 10)
	sessionFile = filepath.Join(os.TempDir(), "tdl", testAccount)
	tclientcore.DCList = dcs.List{
		Options: dcOptions,
		Domains: nil,
		Test:    false,
	}
	tclientcore.DC = 1

	dcpool.TestMode = true

	publicKey, err := crypto.ParseRSAPublicKeys(publicKeyData)
	Expect(err).To(Succeed())
	tclientcore.PublicKeys = []exchange.PublicKey{
		{RSA: publicKey[0]},
	}

	ctx := context.Background()
	Expect(setupTestUser(ctx)).To(Succeed())

	log.SetOutput(GinkgoWriter)
})

var _ = BeforeEach(func() {
	cmd = tcmd.New()
})

func exec(cmd *cobra.Command, args []string, success bool) {
	r, w, err := os.Pipe()
	Expect(err).To(Succeed())
	os.Stdout = w
	color.Output = w

	log.Printf("args: %s\n", args)
	cmd.SetArgs(append([]string{
		"-n", testAccount,
		"--storage", fmt.Sprintf("type=file,path=%s", sessionFile),
	}, args...))
	if err = cmd.Execute(); success {
		Expect(err).To(Succeed())
	} else {
		Expect(err).ToNot(Succeed())
	}

	Expect(w.Close()).To(Succeed())

	o, err := io.ReadAll(r)
	Expect(err).To(Succeed())
	output = string(o)
}

var dcOptions = []tg.DCOption{
	{
		ID:        1,
		IPAddress: "43.155.11.190",
		Port:      10443,
	},
}

func setupTestUser(ctx context.Context) error {
	kvd, err := kv.New(kv.DriverFile, map[string]any{
		"path": sessionFile,
	})
	if err != nil {
		return errors.Wrapf(err, "create kv storage: %s", sessionFile)
	}
	log.Printf("session file: %s", sessionFile)

	stg, err := kvd.Open(testAccount)
	if err != nil {
		return errors.Wrap(err, "open test namespace")
	}

	sess := storage.NewSession(stg, true)

	publicKey, err := crypto.ParseRSAPublicKeys(publicKeyData)
	if err != nil {
		return errors.Wrap(err, "parse public key")
	}

	opts := telegram.Options{
		ReconnectionBackoff: func() backoff.BackOff {
			b := backoff.NewExponentialBackOff()

			b.Multiplier = 1.1
			b.MaxElapsedTime = 0
			b.MaxInterval = 200 * time.Millisecond

			return b
		},
		DC: 1,
		DCList: dcs.List{
			Options: dcOptions,
			Domains: map[int]string{},
			Test:    false,
		},
		PublicKeys: []exchange.PublicKey{
			{RSA: publicKey[0]},
		},
		Device:         tutil.Device,
		SessionStorage: sess,
		RetryInterval:  5 * time.Second,
		MaxRetries:     2,
		DialTimeout:    10 * time.Second,
		Middlewares:    append(tclientcore.NewDefaultMiddlewares(ctx, 0)),
	}

	appId, appHash := tclient.Apps[tclient.AppDesktop].AppID, tclient.Apps[tclient.AppDesktop].AppHash
	c := telegram.NewClient(appId, appHash, opts)

	if err = c.Run(ctx, func(ctx context.Context) error {
		if err = c.Ping(ctx); err != nil {
			return err
		}

		authClient := auth.NewClient(c.API(), rand.Reader, appId, appHash)

		if err = auth.NewFlow(
			testAuth{phone: "+86 13858528382"},
			auth.SendCodeOptions{},
		).Run(ctx, authClient); err != nil {
			return errors.Wrap(err, "register test user")
		}

		user, err := c.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "get self")
		}

		// dc := dcpool.NewPool(c, 1)
		// result, err := dc.Default(ctx).ContactsResolveUsername(ctx, "hshshsh")
		// if err != nil {
		// 	return errors.Wrap(err, "resolve username")
		// }
		// log.Printf("resolve username: %v", result)

		log.Printf("user: %v, %v, %v, %v", user.ID, user.Username, user.FirstName, user.LastName)
		return nil
	}); err != nil {
		return errors.Wrap(err, "run auth")
	}

	// stg.Set(key.App(), []byte(tclient.AppDesktop))

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
