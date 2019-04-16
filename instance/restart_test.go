package instance_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
)

var _ = Describe("Restart", CommandTestBody("restart",
	func(fakeAuthClient *httpclientfakes.FakeAuthenticatedClient, serviceInstanceAdminURL string,
		accessToken string) (string, error) {

		return instance.NewRestartOperation().Run(fakeAuthClient, serviceInstanceAdminURL, accessToken)
	}))
