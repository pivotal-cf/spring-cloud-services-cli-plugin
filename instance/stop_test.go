package instance_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
)

var _ = Describe("Stop", operationTest("stop", instance.NewStopOperation))
