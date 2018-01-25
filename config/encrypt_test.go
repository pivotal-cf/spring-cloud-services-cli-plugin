package config_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"io"

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/config"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("Encrypt", func() {

	const (
		serviceRegistryInstance = "some-service-registry"
		plainText               = "plain-text"
		errorText               = "to err is human"
		accessToken             = "access-token"
		bearerAccessToken       = "bearer " + accessToken
		serviceURI              = "service-uri/"
		encryptURI              = "service-uri/encrypt"
		cipherText              = "cipher-text"
	)

	var (
		fakeCliConnection     *pluginfakes.FakeCliConnection
		fakeAuthClient        *httpclientfakes.FakeAuthenticatedClient
		fakeResolver          func(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error)
		resolverAccessToken   string
		testError             error
		postResponse          io.ReadCloser
		postStatusCode        int
		postErr               error
		output                string
		err                   error
		accessTokenURI        string
		fakeResolverCallCount int
		fakeResolverError     error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		testError = errors.New(errorText)
		postResponse, postStatusCode, postErr = ioutil.NopCloser(bytes.NewBufferString(cipherText)), http.StatusOK, nil
		accessTokenURI = "access-token-uri"
		fakeCliConnection.AccessTokenReturns(bearerAccessToken, nil)
		fakeResolverCallCount = 0
		fakeResolver = func(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
			fakeResolverCallCount++
			resolverAccessToken = accessToken
			return serviceURI, fakeResolverError
		}
		fakeResolverError = nil
	})

	JustBeforeEach(func() {
		fakeAuthClient.DoAuthenticatedPostReturns(postResponse, postStatusCode, postErr)
		output, err = config.EncryptWithResolver(fakeCliConnection, serviceRegistryInstance, plainText, "", fakeAuthClient, fakeResolver)
	})

	It("should create an access token", func() {
		Expect(fakeCliConnection.AccessTokenCallCount()).To(Equal(1))
	})

	Context("when the access token is not available", func() {
		BeforeEach(func() {
			fakeCliConnection.AccessTokenReturns("", errors.New("some access token error"))
		})

		It("should return a suitable error", func() {
			Expect(fakeCliConnection.AccessTokenCallCount()).To(Equal(1))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Access token not available: some access token error"))
		})
	})

	It("should call the service instance resolver", func() {
		Expect(fakeResolverCallCount).To(Equal(1))
	})

	It("should pass the access token to the resolver", func() {
		Expect(resolverAccessToken).To(Equal(accessToken))
	})

	Context("when the service instance resolver fails", func() {
		BeforeEach(func() {
			fakeResolverError = testError
		})

		It("should propagate the error", func() {
			Expect(err).To(MatchError("Error obtaining config server URL: " + errorText))
		})
	})

	It("should call the config server's /encrypt endpoint with the correct parameters", func() {
		Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(1))
		url, bodyType, body, token := fakeAuthClient.DoAuthenticatedPostArgsForCall(0)
		Expect(url).To(Equal(encryptURI))
		Expect(bodyType).To(Equal("text/plain"))
		Expect(body).To(Equal(plainText))
		Expect(token).To(Equal(accessToken))
	})

	Context("when the config server's /encrypt endpoint fails", func() {
		BeforeEach(func() {
			postResponse, postStatusCode, postErr = nil, 0, testError
		})

		It("should propagate the error", func() {
			Expect(err.Error()).To(ContainSubstring(errorText))
		})
	})

	Context("when the config server's encrypt endpoint returns a response body which cannot be read", func() {
		BeforeEach(func() {
			postResponse, postStatusCode, postErr = &badReader{readErr: testError}, http.StatusOK, nil
		})

		It("should propagate the error", func() {
			Expect(err.Error()).To(ContainSubstring(errorText))
		})
	})

	Context("when the config server's encrypt endpoint returns a response body which cannot be closed", func() {
		BeforeEach(func() {
			postResponse, postStatusCode, postErr = &badCloser{reader: bytes.NewBufferString(cipherText), closeErr: testError}, http.StatusOK, nil
		})

		It("should ignore the error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	It("should return the output from the config server's /encrypt endpoint", func() {
		Expect(output).To(Equal(cipherText))
		Expect(err).NotTo(HaveOccurred())
	})
})

type badReader struct {
	readErr error
}

func (br *badReader) Read(p []byte) (n int, err error) {
	return 0, br.readErr
}

func (*badReader) Close() error {
	return nil
}

type badCloser struct {
	reader   io.Reader
	closeErr error
}

func (bc *badCloser) Read(p []byte) (n int, err error) {
	return bc.reader.Read(p)
}

func (brc *badCloser) Close() error {
	return brc.closeErr
}
