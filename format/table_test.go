package format_test

import (
	"fmt"
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
)

var _ = Describe("Table", func() {
	var tab *format.Table

	Context("when the table has no body", func() {
		BeforeEach(func() {
			tab = &format.Table{}
			tab.Entitle([]string{"a", "b"})
		})

		It("should output only the title in bold", func() {
			bold := color.New(color.Bold).SprintfFunc()
			Expect(tab.String()).To(ContainSubstring(fmt.Sprintf("%s %s \n", bold("a"), bold("b"))))
		})

	})

	Context("when the table has a body", func() {
		BeforeEach(func() {
			tab = &format.Table{}
			tab.Entitle([]string{"a", "bb", "c"})
			tab.AddRow([]string{"aa", "b", "cc"})
		})

		It("should output the title and row in the correct colors", func() {
			bold := color.New(color.Bold).SprintfFunc()
			cyan := color.New(color.FgHiCyan).SprintfFunc()
			Expect(tab.String()).To(ContainSubstring(fmt.Sprintf("%s  %s %s  \n%s %s  %s \n",
				bold("a"), bold("bb"), bold("c"),
				cyan("aa"), "b", "cc")))
		})

	})

})
