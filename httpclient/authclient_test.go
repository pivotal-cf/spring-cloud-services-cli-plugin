package httpclient_test

import (
	"io/ioutil"
	"net/http"
	"strings"

	"errors"

	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

type badReader struct{}

func (b badReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

var _ = Describe("Authclient", func() {
	const testAccessToken = "securetoken"
	const testUrl = "https://eureka.pivotal.io/auth/request"
	var (
		fakeClient  *httpclientfakes.FakeClient
		authClient  httpclient.AuthenticatedClient
		url         string
		accessToken string
		buf         *bytes.Buffer
		err         error
	)

	BeforeEach(func() {
		fakeClient = &httpclientfakes.FakeClient{}
		url = testUrl
		accessToken = testAccessToken
	})

	JustBeforeEach(func() {
		authClient = httpclient.NewAuthenticatedClient(fakeClient)
		buf, err = authClient.DoAuthenticatedGet(url, accessToken)
	})

	Context("when the underlying request cannot be created", func() {
		BeforeEach(func() {
			url = ":"
		})

		It("should return a suitable error if the request cannot be created", func() {
			Expect(buf).To(BeNil())
			Expect(err).To(MatchError("Request creation error: parse :: missing protocol scheme"))
		})
	})

	Context("when the underlying request can be created", func() {
		Context("and an authenticated request is made", func() {

			Context("but the request fails", func() {
				BeforeEach(func() {
					fakeClient.DoReturns(nil, errors.New(`request failed`))
					authClient = httpclient.NewAuthenticatedClient(fakeClient)
				})

				It("should produce an error", func() {
					Expect(buf).To(BeNil())
					Expect(err).To(MatchError("authenticated get of 'https://eureka.pivotal.io/auth/request' failed: request failed"))
				})
			})

			Context("and the request succeeds", func() {
				Context("but the response body is nil", func() {
					BeforeEach(func() {
						resp := &http.Response{}
						//resp.Body = nil
						fakeClient.DoReturns(resp, nil)
					})

					It("should produce an error", func() {
						Expect(buf).To(BeNil())
						Expect(err).To(MatchError("authenticated get of 'https://eureka.pivotal.io/auth/request' failed: nil response body"))
					})
				})

				Context("but the response body cannot be read", func() {
					BeforeEach(func() {
						resp := &http.Response{}

						resp.Body = ioutil.NopCloser(badReader{})
						fakeClient.DoReturns(resp, nil)
					})

					It("should produce an error", func() {
						Expect(buf).To(BeNil())
						Expect(err).To(MatchError("authenticated get of 'https://eureka.pivotal.io/auth/request' failed: body cannot be read"))
					})
				})

				Context("and the response body can be read", func() {
					BeforeEach(func() {
						resp := &http.Response{}
						resp.Body = ioutil.NopCloser(strings.NewReader("payload"))
						fakeClient.DoReturns(resp, nil)
					})

					It("should have sent a request with the correct accept header", func() {
						Expect(fakeClient.DoCallCount()).To(Equal(1))
						req := fakeClient.DoArgsForCall(0)
						Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					})

					It("should have sent a request with the correct authorization header", func() {
						Expect(fakeClient.DoCallCount()).To(Equal(1))
						req := fakeClient.DoArgsForCall(0)
						Expect(req.Header.Get("Authorization")).To(Equal("securetoken"))
					})

					It("should produce a non-empty response body", func() {
						Expect(buf.String()).Should(Equal("payload"))
						Expect(err).To(BeNil())
					})
				})
			})
		})
	})
})
