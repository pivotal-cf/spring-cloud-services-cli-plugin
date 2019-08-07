package instance_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
)

var _ = Describe("Start", operationTest("start", instance.NewStartOperation))
