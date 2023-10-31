package instance_test

import (
	"errors"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

type operationFactory func(client httpclient.AuthenticatedClient) instance.Operation

func operationTest(command string, operationFactory operationFactory) func() {
	return func() {
		const (
			applicationJson     = "application/json"
			testAccessToken     = "someaccesstoken"
			managementUrl       = "https://some.host/x/y/cli/instances/someguid"
			serviceOfferingName = "p.config-server"
			planName            = "standard"
		)

		var (
			fakeAuthClient       *httpclientfakes.FakeAuthenticatedClient
			managementParameters serviceutil.ManagementParameters

			output string
			err    error
		)

		BeforeEach(func() {
			fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
			managementParameters = serviceutil.ManagementParameters{
				Url:                 managementUrl,
				ServiceOfferingName: serviceOfferingName,
				ServicePlanName:     planName,
			}
		})

		JustBeforeEach(func() {
			output, err = operationFactory(fakeAuthClient).Run(managementParameters, testAccessToken)
		})

		It("should issue a PUT with the correct parameters", func() {
			Expect(fakeAuthClient.DoAuthenticatedPutCallCount()).To(Equal(1))
			url, bodyType, body, accessToken := fakeAuthClient.DoAuthenticatedPutArgsForCall(0)
			Expect(url).To(Equal(fmt.Sprintf("%s/command?%s=", managementUrl, command)))
			Expect(accessToken).To(Equal(testAccessToken))
			Expect(bodyType).To(Equal(applicationJson))
			Expect(body).To(MatchJSON(fmt.Sprintf(`{"serviceOfferingName": "%s", "planName": "%s"}`, serviceOfferingName, planName)))
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
