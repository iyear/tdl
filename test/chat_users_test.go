package test

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test tdl chat users", FlakeAttempts(3), func() {
	var (
		once sync.Once
		skip bool
		id   int64
	)

	BeforeEach(func() {
		Skip("no better test method, so we skip now")

		if skip {
			Skip("skip because of no channel/group")
		}

		args = []string{"chat", "users"}

		once.Do(func() {
			By("list all channels/groups to iter")
			args := []string{"chat", "ls", "-f", `Type in ["channel","group"]`, "-o", "json"}

			exec(cmd, args, true)

			r := gjson.Parse(output)
			if len(r.Array()) == 0 {
				skip = true
				return
			}

			id = r.Get("0.id").Int()
			Expect(id).NotTo(BeEmpty())
		})
	})

	When("use chat flag", func() {
		It("should success", func() {
			args = append(args, "-c", strconv.FormatInt(id, 10))

			exec(cmd, args, true)

			j, err := os.ReadFile("tdl-users.json")
			Expect(err).To(Succeed())

			r := gjson.ParseBytes(j)

			Expect(r.Get("id").Int()).To(BeNumerically("==", id))
			Expect(len(r.Get("users").Array())).To(BeNumerically(">", 0))
			Expect(r.Get("kicked").Exists()).To(BeTrue())
			Expect(r.Get("banned").Exists()).To(BeTrue())
			Expect(r.Get("admins").Exists()).To(BeTrue())
			Expect(r.Get("bots").Exists()).To(BeTrue())

			log.Println("users:", len(r.Get("users").Array()))
		})

		It("with invalid chat", func() {
			args = append(args, "-c", "invalid_username")

			exec(cmd, args, false)
		})
	})
})
