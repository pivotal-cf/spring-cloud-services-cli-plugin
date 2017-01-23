package cli_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cli"
)

var _ = Describe("Flags", func() {

	var (
		args           = []string{"cf", "srd", "provision-service-registry", "provision-sr-1", "-i", "1", "-i", "2", "-i", "3"}
		instanceIdx    *int
		sslNoVerify    bool
		positionalArgs []string
		err            error
	)

	JustBeforeEach(func() {
		sslNoVerify, instanceIdx, positionalArgs, err = cli.ParseFlags(args)
	})

	Context("when duplicate flags are passed on command line", func() {

		It("the final of the duplicated flags will be chosen", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(*instanceIdx).To(Equal(3))
		})
	})

	Context("when an unexpected flag is received", func() {

		BeforeEach(func() {
			args = []string{"cf", "srd", "provision-service-registry", "provision-sr-1", "-i", "1", "-z"}
		})

		It("should raise a suitable error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Error parsing arguments: Invalid flag: -z"))
		})
	})

	Context("when the skip ssl validation flag is set", func() {

		BeforeEach(func() {
			args = []string{"cf", "srd", "provision-service-registry", "provision-sr-1", "--skip-ssl-validation"}
		})

		It("should capture the flags value", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(sslNoVerify).To(BeTrue())
		})

		Context("when positional arguments are used by the command", func() {

			It("should capture an array of positional arguments", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(positionalArgs)).To(Equal(4))
			})
		})

	})

	Context("when no flag is passed for instance index", func() {

		BeforeEach(func() {
			args = []string{"cf", "srd", "provision-service-registry", "provision-sr-1", "--skip-ssl-validation"}
		})

		It("should be parsed as nil", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(instanceIdx).To(BeNil())
		})
	})

	Context("when a string value is passed for instance index", func() {

		BeforeEach(func() {
			args = []string{"cf", "srd", "provision-service-registry", "provision-sr-1", "--skip-ssl-validation", "-i", "one"}
		})

		It("should be parsed as nil", func() {
			Expect(err).To(MatchError("Error parsing arguments: Value for flag 'cf-instance-index' must be an integer"))
		})
	})
})
