package cli_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cli"
)

var _ = Describe("ArgConsumer", func() {

	var argConsumer *cli.ArgConsumer

	Context("when there are no arguments (which should never happen in practice)", func() {
		It("should panic", func() {
			Expect(func() {
				cli.NewArgConsumer([]string{}, func(message string, command string) {})
			}).To(Panic())
		})
	})

	Context("when there is at least one argument", func() {
		var (
			args               []string
			diagnose           cli.DiagnosticFunc
			diagnoseCallCount  int
			diagnoseMessageArg string
			diagnoseCommandArg string
		)

		BeforeEach(func() {
			diagnoseCallCount = 0
			diagnoseMessageArg = ""
			diagnoseCommandArg = ""
			diagnose = func(message string, command string) {
				diagnoseCallCount++
				diagnoseMessageArg = message
				diagnoseCommandArg = command
			}
		})

		JustBeforeEach(func() {
			argConsumer.CheckAllConsumed()
		})

		Context("when there is one argument", func() {
			BeforeEach(func() {
				args = []string{"command"}
				argConsumer = cli.NewArgConsumer(args, diagnose)
			})

			It("should not fail", func() {
				Expect(diagnoseCallCount).To(Equal(0))
			})

			Context("when an attempt is made to consume a second argument", func() {
				BeforeEach(func() {
					argConsumer.Consume(2, "second argument")
				})

				It("should diagnose the problem", func() {
					Expect(diagnoseCallCount).To(Equal(1))
					Expect(diagnoseCommandArg).To(Equal("command"))
					Expect(diagnoseMessageArg).To(Equal("Incorrect usage: second argument not specified."))
				})
			})
		})

		Context("when there are two arguments", func() {
			BeforeEach(func() {
				args = []string{"command", "arg2"}
				argConsumer = cli.NewArgConsumer(args, diagnose)
			})

			Context("when the second argument is not consumed", func() {
				It("should fail", func() {
					Expect(diagnoseCallCount).To(Equal(1))
					Expect(diagnoseCommandArg).To(Equal("command"))
					Expect(diagnoseMessageArg).To(Equal("Incorrect usage: invalid argument 'arg2'."))
				})
			})

			Context("when the second argument is consumed", func() {
				BeforeEach(func() {
					argConsumer.Consume(1, "second argument")
				})

				It("should not fail", func() {
					Expect(diagnoseCallCount).To(Equal(0))
				})
			})

			Context("when an attempt is made to consume a third argument", func() {
				BeforeEach(func() {
					argConsumer.Consume(1, "second argument")
					argConsumer.Consume(2, "third argument")
				})

				It("should diagnose the problem", func() {
					Expect(diagnoseCallCount).To(Equal(1))
					Expect(diagnoseCommandArg).To(Equal("command"))
					Expect(diagnoseMessageArg).To(Equal("Incorrect usage: third argument not specified."))
				})
			})
		})

		Context("when there are three arguments", func() {
			BeforeEach(func() {
				args = []string{"command", "arg2", "arg3"}
				argConsumer = cli.NewArgConsumer(args, diagnose)
			})

			Context("when more than one argument is not consumed", func() {
				It("should fail with the correct plural 'arguments' and a list of the invalid arguments", func() {
					Expect(diagnoseCallCount).To(Equal(1))
					Expect(diagnoseCommandArg).To(Equal("command"))
					Expect(diagnoseMessageArg).To(Equal("Incorrect usage: invalid arguments 'arg2 arg3'."))
				})
			})
		})
	})
})
