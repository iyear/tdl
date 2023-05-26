package test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"math/rand"
	"os"
	"path/filepath"
)

var _ = Describe("Test tdl upload", FlakeAttempts(3), func() {
	BeforeEach(func() {
		args = []string{"upload"}
	})

	var (
		dir   string
		files []string
	)

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		// create files
		files = make([]string, 0)
		for i := 0; i < rand.Intn(3)+3; i++ {
			file := filepath.Join(dir, uuid.New().String())

			// generate random file with size between 1MB and 2MB
			f, err := os.Create(file)
			Expect(err).To(Succeed())
			Expect(f.Truncate(int64(rand.Intn(1e5)) + 1e5)).To(Succeed())
			Expect(f.Close()).To(Succeed())

			files = append(files, file)
		}
	})

	When("use path flag", func() {
		BeforeEach(func() {
			args = append(args, "-p")
		})

		It("should success", func() {
			args = append(args, dir)
			exec(cmd, args, true)
		})
	})
})
