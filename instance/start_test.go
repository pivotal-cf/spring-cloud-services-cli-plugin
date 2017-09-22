package instance_test

import (
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
)

var _ = DescribeCommandTest("Start", "start",
	func(fakeAuthClient *httpclientfakes.FakeAuthenticatedClient, serviceInstanceAdminURL string,
		accessToken string) (string, error) {

		return instance.Start(fakeAuthClient, serviceInstanceAdminURL, accessToken)
	})
