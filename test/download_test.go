package test

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test tdl download", FlakeAttempts(3), func() {
	var (
		once        sync.Once
		fileHash    = make(map[string][16]byte)
		id          int64
		remoteFiles = make([]int64, 0)
	)

	BeforeEach(func() {
		once.Do(func() {
			By("collect local file hashes")
			Expect(filepath.WalkDir("testdata", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}

				bytes, err := os.ReadFile(path)
				Expect(err, Succeed())
				h := md5.Sum(bytes)
				fileHash[filepath.Base(path)] = h
				log.Println("path:", path, "md5:", h)

				return nil
			})).To(Succeed())

			By("upload files")
			exec(cmd, []string{"upload", "-p", "testdata"}, true)

			By("export uploaded files")
			exportFile := filepath.Join(GinkgoT().TempDir(), "export.json")
			exec(cmd, []string{"chat", "export", "-T", "last", "-i", strconv.Itoa(len(fileHash)), "-o", exportFile}, true)
			exportBytes, err := os.ReadFile(exportFile)
			Expect(err).To(Succeed())

			By("get chat id and remote file ids")
			id = gjson.GetBytes(exportBytes, "id").Int()
			Expect(id).NotTo(BeZero())

			gjson.GetBytes(exportBytes, "messages").ForEach(func(key, value gjson.Result) bool {
				remoteFiles = append(remoteFiles, value.Get("id").Int())
				return true
			})
			Expect(len(remoteFiles)).To(Equal(len(fileHash)))
		})
	})

	When("use url flag", func() {
		It("should success", func() {
			urls := make([]string, 0)
			for _, u := range remoteFiles {
				urls = append(urls, "-u", fmt.Sprintf("https://t.me/%d/%d", id, u))
			}

			dir := GinkgoT().TempDir()
			args := []string{"download", "-d", dir, "--template", "{{ .FileName }}"}
			args = append(args, urls...)
			exec(cmd, args, true)

			log.Println("check local files")
			Expect(filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				bytes, err := os.ReadFile(path)
				Expect(err, Succeed())
				h := md5.Sum(bytes)
				log.Println("path:", path, "md5:", h)

				Expect(h).To(Equal(fileHash[filepath.Base(path)]))
				return nil
			})).To(Succeed())
		})
	})
})
