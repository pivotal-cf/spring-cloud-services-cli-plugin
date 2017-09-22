package instance_test

import (
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
)

var _ = DescribeCommandTest("Restage", "restage",
	func(fakeAuthClient *httpclientfakes.FakeAuthenticatedClient, serviceInstanceAdminURL string,
		accessToken string) (string, error) {

		return instance.Restage(fakeAuthClient, serviceInstanceAdminURL, accessToken)
	})
