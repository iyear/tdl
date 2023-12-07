package test

import (
	"fmt"
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

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test tdl")
}

var (
	cmd         *cobra.Command
	args        []string
	output      string
	testAccount string
)

func storageValue(test string) string {
	return fmt.Sprintf("type=file,path=%s",
		filepath.Join(os.TempDir(), "tdl", testAccount))
}

var _ = BeforeSuite(func() {
	testAccount = strconv.FormatInt(time.Now().UnixNano(), 10)

	exec(tcmd.New(), []string{"login", "--code"}, true)

	log.SetOutput(GinkgoWriter)
})

var _ = BeforeEach(func() {
	cmd = tcmd.New()

	// wait before each test to avoid rate limit
	time.Sleep(10 * time.Second)
})

func exec(cmd *cobra.Command, args []string, success bool) {
	r, w, err := os.Pipe()
	Expect(err).To(Succeed())
	os.Stdout = w
	color.Output = w

	log.Printf("args: %s\n", args)
	cmd.SetArgs(append([]string{
		"--test", testAccount,
		"-n", testAccount,
		"--storage", storageValue(testAccount),
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
