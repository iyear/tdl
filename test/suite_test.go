package test

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	tcmd "github.com/iyear/tdl/cmd"
	"github.com/iyear/tdl/test/testserver"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test tdl")
}

var (
	cmd         *cobra.Command
	args        []string
	output      string
	testAccount string
	sessionFile string
)

var _ = BeforeSuite(func(ctx context.Context) {
	var err error
	testAccount, sessionFile, err = testserver.Setup(ctx, rand.NewSource(GinkgoRandomSeed()))
	Expect(err).To(Succeed())

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
