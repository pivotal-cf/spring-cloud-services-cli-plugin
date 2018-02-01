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

var _ = Describe("Encrypt file", func() {

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
		config.DefaultResolver = fakeResolver
		fakeResolverError = nil
	})

	JustBeforeEach(func() {
		fakeAuthClient.DoAuthenticatedPostReturns(postResponse, postStatusCode, postErr)
	})

	It("should call the config server's /encrypt endpoint with content from a file", func() {
		testDir := config.CreateTempDir()
		testFile := config.CreateFile(testDir, "file-to-encrypt.txt")
		output, err = config.Encrypt(fakeCliConnection, serviceRegistryInstance, "", testFile, fakeAuthClient)

		Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(1))
		url, bodyType, body, token := fakeAuthClient.DoAuthenticatedPostArgsForCall(0)
		Expect(url).To(Equal(encryptURI))
		Expect(bodyType).To(Equal("text/plain"))
		Expect(body).To(Equal("Hello\nWorld\n"))
		Expect(token).To(Equal(accessToken))
	})

	It("should fail when given a non-existent file", func() {
		output, err = config.Encrypt(fakeCliConnection, serviceRegistryInstance, "", "bogus.txt", fakeAuthClient)

		Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(0))
		Expect(err.Error()).To(Equal("Error opening file at path bogus.txt : open bogus.txt: no such file or directory"))
	})

})
