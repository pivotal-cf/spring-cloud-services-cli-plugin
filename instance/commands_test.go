package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

func CommandTestBody(command string,
	commandClosure func(
		fakeAuthClient *httpclientfakes.FakeAuthenticatedClient,
		serviceInstanceAdminURL string,
		accessToken string) (string, error)) func() {

	return func() {
		const testAccessToken = "someaccesstoken"

		var (
			fakeAuthClient          *httpclientfakes.FakeAuthenticatedClient
			serviceInstanceAdminURL string

			output string
			err    error
		)

		BeforeEach(func() {
			fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
			serviceInstanceAdminURL = "http://some.host/x/y/cli/instances/someguid"

		})

		JustBeforeEach(func() {
			output, err = commandClosure(fakeAuthClient, serviceInstanceAdminURL, testAccessToken)
		})

		It("should issue a PUT with the correct parameters", func() {
			Expect(fakeAuthClient.DoAuthenticatedPutCallCount()).To(Equal(1))
			url, accessToken := fakeAuthClient.DoAuthenticatedPutArgsForCall(0)
			Expect(url).To(Equal(fmt.Sprintf("http://some.host/x/y/cli/instances/someguid/command?%s", command)))
			Expect(accessToken).To(Equal(testAccessToken))
			Expect(output).To(Equal(""))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when PUT return an error", func() {
			var testError error

			BeforeEach(func() {
				testError = errors.New("failure is not an option")
				fakeAuthClient.DoAuthenticatedPutReturns(99, testError)
			})

			It("should return the error", func() {
				Expect(output).To(Equal(""))
				Expect(err).To(Equal(testError))
			})
		})
	}
}
