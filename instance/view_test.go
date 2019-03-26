package instance_test

import (
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"

	"errors"
	"net/http"

	"bytes"
	"io/ioutil"

	"fmt"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("View", func() {

	const testAccessToken = "someaccesstoken"

	var (
		fakeAuthClient          *httpclientfakes.FakeAuthenticatedClient
		serviceInstanceAdminURL string

		output string
		err    error
	)

	BeforeEach(func() {
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		serviceInstanceAdminURL = "https://some.host/x/y/cli/instances/someguid"

	})

	JustBeforeEach(func() {
		output, err = instance.View(fakeAuthClient, serviceInstanceAdminURL, testAccessToken)
	})

	It("should issue a GET with the correct parameters", func() {
		Expect(fakeAuthClient.DoAuthenticatedGetCallCount()).To(Equal(1))
		url, accessToken := fakeAuthClient.DoAuthenticatedGetArgsForCall(0)
		Expect(url).To(Equal("https://some.host/x/y/cli/instances/someguid"))
		Expect(accessToken).To(Equal(testAccessToken))
	})

	Context("when GET return an error", func() {
		var testError error

		BeforeEach(func() {
			testError = errors.New("failure is not an option")
			fakeAuthClient.DoAuthenticatedGetReturns(nil, 99, testError)
		})

		It("should return the error", func() {
			Expect(output).To(Equal(""))
			Expect(err).To(Equal(testError))
		})
	})

	Context("when GET returns a bad status code", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(nil, http.StatusNotFound, nil)
		})

		It("should return the error", func() {
			Expect(err).To(MatchError("Service broker view instance failed: 404"))
		})
	})

	Context("when GET returns no response reader", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(nil, http.StatusOK, nil)
		})

		It("should return a suitable error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Service broker view instance response body missing"))
		})
	})

	Context("when GET returns response reader which cannot be read", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(badReader{}), http.StatusOK, nil)
		})

		It("should return a suitable error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Cannot read service broker view instance response body: read error"))
		})
	})

	Context("when GET returns invalid JSON", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{`)), http.StatusOK, nil)
		})

		It("should return a suitable error", func() {
			Expect(output).To(Equal(""))
			Expect(err).To(MatchError("Invalid service broker view instance response JSON: unexpected end of JSON input, response body: '{'"))
		})
	})

	Context("when GET returns valid JSON", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`
						{"backing_apps":[
			               {"name":"eureka-4bf65bd3-b587-4eec-94a0-7023cb94dff8",
			                "buildpack":"container-certificate-trust-store=2.0.0_RELEASE java-buildpack=v3.13-offline-https://github.com/cloudfoundry/java-buildpack.git#03b493f java-main open-jdk-like-jre=1.8.0_121 open-jdk-like-memory-calculator=2.0.2_RELEASE spring-auto-reconfiguration=1.10...",
			                "stack":"cflinuxfs2",
			                "memory":1024,
			                "routes":["eureka-4bf65bd3-b587-4eec-94a0-7023cb94dff8.olive.springapps.io"],
			                "instances":[
			                    {"index":0,
			                     "state":"RUNNING",
			                     "since":1500682722000,
			                     "cpu":0.002385765443113775,
			                     "details":"DETAILS",
			                     "memory_usage":1039134720,
			                     "memory_quota":1073741824,
			                     "disk_usage":195514368,
			                     "disk_quota":1073741824},
			                    {"index":1,
			                     "state":"STARTING",
			                     "since":2500682722000,
			                     "cpu":1.0,
			                     "details":"DETAILS",
			                     "memory_usage":2039134720,
			                     "memory_quota":2073741824,
			                     "disk_usage":295514368,
			                     "disk_quota":2073741824}
			                ],
			                "last_uploaded":1498494177000,
			                "num_instances":1,
			                "running_instances":1,
			                "requested_state":"STARTED"}
			            ]}`)), http.StatusOK, nil)
		})

		It("should print the expected output", func() {
			Expect(output).To(Equal(fmt.Sprintf(`
backing app name: eureka-4bf65bd3-b587-4eec-94a0-7023cb94dff8
requested state:  started
instances:        1/1
usage:            1G x 1 instances
routes:           eureka-4bf65bd3-b587-4eec-94a0-7023cb94dff8.olive.springapps.io
last uploaded:    %s
stack:            cflinuxfs2
buildpack:        container-certificate-trust-store=2.0.0_RELEASE java-buildpack=v3.13-offline-https://github.com/cloudfoundry/java-buildpack.git#03b493f java-main open
                  -jdk-like-jre=1.8.0_121 open-jdk-like-memory-calculator=2.0.2_RELEASE spring-auto-reconfiguration=1.10...

     state     since                  cpu    memory       disk           details
#0   running   2017-07-22T00:18:42Z   0.2%%   991M of 1G   186.5M of 1G   DETAILS
#1   starting  2049-03-30T02:05:22Z   100.0%% 1.9G of 1.9G 281.8M of 1.9G DETAILS
`, time.Unix(1498494177, 0).Local().Format("Mon 02 Jan 15:04:05 MST 2006"))))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

type badReader struct{}

func (b badReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
