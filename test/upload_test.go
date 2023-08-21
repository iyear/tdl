package test

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	checkFiles := func(chat string, n int, expected []string) {
		By("check if files are uploaded")
		exportFile := filepath.Join(dir, "export.json")
		exec(cmd, []string{"chat", "-c", chat, "export", "-T", "last", "-i", strconv.Itoa(n), "-o", exportFile}, true)

		exportBytes, err := os.ReadFile(exportFile)
		Expect(err).To(Succeed())

		actualFiles := make([]string, 0)
		gjson.GetBytes(exportBytes, "messages").ForEach(func(key, value gjson.Result) bool {
			actualFiles = append(actualFiles, filepath.Join(dir, value.Get("file").String()))
			return true
		})
		log.Printf("actual files on server: %v", actualFiles)

		Expect(actualFiles).To(ConsistOf(expected))
	}

	When("use path flag", func() {
		It("should success", func() {
			args = append(args, "-p", dir)
			exec(cmd, args, true)

			checkFiles("", len(files), files)
		})

		It("should fail with invalid path", func() {
			args = append(args, "-p", "foo")
			exec(cmd, args, false)
		})

		It("should fail with invalid file", func() {
			args = append(args, "-p", "foo.bar")
			exec(cmd, args, false)
		})
	})

	When("use rm flag", func() {
		It("should success", func() {
			args = append(args, "-p", dir, "--rm")
			exec(cmd, args, true)

			checkFiles("", len(files), files)

			By("check if files are removed")
			for _, file := range files {
				_, err := os.Stat(file)
				Expect(os.IsNotExist(err)).To(BeTrue())
			}
		})
	})

	When("use chat flag", func() {
		It("should success", func() {
			By("get a private chat id")
			exec(cmd, []string{"chat", "ls", "-o", "json", "-f", "Type contains 'private'"}, true)
			chat := gjson.Get(output, "0.id").String()
			Expect(chat).NotTo(BeEmpty())

			args = append(args, "-p", dir, "-c", chat)
			exec(cmd, args, true)

			checkFiles(chat, len(files), files)
		})

		It("should fail with invalid chat domain", func() {
			args = append(args, "-p", dir, "-c", "foo")
			exec(cmd, args, false)
		})

		It("should fail with invalid chat id", func() {
			args = append(args, "-p", dir, "-c", "-100")
			exec(cmd, args, false)
		})
	})

	When("use exclude flag", func() {
		It("should success", func() {
			By("modify files' extension")
			modify, remain := files[:len(files)/2], files[len(files)/2:]
			log.Printf("modify files: %v", modify)
			log.Printf("remain files: %v", remain)

			for _, file := range modify {
				Expect(os.Rename(file, file+".foo")).To(Succeed())
			}

			args = append(args, "-p", dir, "-e", ".foo")
			exec(cmd, args, true)

			checkFiles("", len(remain), remain)
		})
	})
})
