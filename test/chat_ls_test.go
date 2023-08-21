package test

import (
	"encoding/json"
	"strings"

	"github.com/iyear/tdl/app/chat"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test tdl chat ls", FlakeAttempts(3), func() {
	BeforeEach(func() {
		args = []string{"chat", "ls"}
	})

	expectTable := func() {
		lines := strings.Split(output, "\n")
		By("check header")
		Expect(len(lines)).To(BeNumerically(">=", 2))
		Expect(lines[0]).To(MatchRegexp("ID\\s+Type\\s+VisibleName\\s+Username\\s+Topics"))
	}

	When("use output flag", func() {
		It("with default", func() {
			exec(cmd, args, true)

			expectTable()
		})

		It("with table", func() {
			exec(cmd, append(args, "--output", "table"), true)

			expectTable()
		})

		It("with json", func() {
			exec(cmd, append(args, "--output", "json"), true)

			Expect(json.Valid([]byte(output))).To(BeTrue())
		})
	})

	When("use filter flag", func() {
		BeforeEach(func() {
			args = append(args, "--output", "json", "--filter")
		})

		readDialogs := func() []*chat.Dialog {
			dialogs := make([]*chat.Dialog, 0)
			Expect(json.Unmarshal([]byte(output), &dialogs)).To(Succeed())
			return dialogs
		}

		It("to display available fields", func() {
			exec(cmd, append(args, "-"), true)

			Expect(len(strings.Split(output, "\n"))).To(BeNumerically(">=", 2))
		})

		It("to filter id", func() {
			exec(cmd, append(args, "ID>2200000000"), true)

			for _, dialog := range readDialogs() {
				Expect(dialog.ID).To(BeNumerically(">", 2200000000))
			}
		})

		It("to filter type", func() {
			exec(cmd, append(args, "Type contains 'private'"), true)

			for _, dialog := range readDialogs() {
				Expect(dialog.Type).To(Equal("private"))
			}
		})

		It("to filter visible name", func() {
			exec(cmd, append(args, "VisibleName contains 'Telegram'"), true)

			for _, dialog := range readDialogs() {
				Expect(dialog.VisibleName).To(ContainSubstring("Telegram"))
			}
		})

		It("to filter username", func() {
			exec(cmd, append(args, "Username contains 'telegram'"), true)

			for _, dialog := range readDialogs() {
				Expect(dialog.Username).To(ContainSubstring("telegram"))
			}
		})

		It("to filter topics", func() {
			exec(cmd, append(args, "len(Topics)>0"), true)

			for _, dialog := range readDialogs() {
				Expect(len(dialog.Topics)).To(BeNumerically(">", 0))
			}
		})

		It("to filter multiple fields", func() {
			exec(cmd, append(args, "ID>2200000000 && len(Topics)>0"), true)

			for _, dialog := range readDialogs() {
				Expect(dialog.ID).To(BeNumerically(">", 2200000000))
				Expect(len(dialog.Topics)).To(BeNumerically(">", 0))
			}
		})
	})
})
