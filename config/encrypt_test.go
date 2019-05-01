package config_test

import (
	"bytes"
	"errors"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil/serviceutilfakes"
	"io/ioutil"
	"net/http"

	"io"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/config"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"strings"
)

var _ = Describe("Encrypter", func() {

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
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient
		fakeResolver      *serviceutilfakes.FakeServiceInstanceUrlResolver
		testError         error
		//postResponse      io.ReadCloser
		//postStatusCode    int
		//postErr           error
		output    string
		err       error
		encrypter config.Encrypter
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		fakeResolver = &serviceutilfakes.FakeServiceInstanceUrlResolver{}
		testError = errors.New(errorText)

		fakeCliConnection.AccessTokenReturns(bearerAccessToken, nil)
		fakeAuthClient.DoAuthenticatedPostReturns(ioutil.NopCloser(bytes.NewBufferString(cipherText)), http.StatusOK, nil)
		fakeResolver.GetServiceInstanceUrlReturns(serviceURI, nil)
	})

	JustBeforeEach(func() {
		encrypter = config.NewEncrypter(fakeCliConnection, fakeAuthClient, fakeResolver)
	})

	itAuthenticatesAndResolvesTheServiceInstanceUrl := func() {
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
			Expect(fakeResolver.GetServiceInstanceUrlCallCount()).To(Equal(1))
			serviceInstanceName, resolverAccessToken := fakeResolver.GetServiceInstanceUrlArgsForCall(0)
			Expect(serviceInstanceName).To(Equal(serviceInstanceName))
			Expect(resolverAccessToken).To(Equal(accessToken))
		})

		Context("when the service instance resolver fails", func() {
			BeforeEach(func() {
				fakeResolver.GetServiceInstanceUrlReturns("", testError)
			})

			It("should propagate the error", func() {
				Expect(err).To(MatchError("Error obtaining config server URL: " + errorText))
			})
		})
	}

	Describe("Encrypt", func() {
		JustBeforeEach(func() {
			output, err = encrypter.EncryptString(serviceRegistryInstance, plainText)
		})

		itAuthenticatesAndResolvesTheServiceInstanceUrl()

		It("should call the config server's /encrypt endpoint with the correct parameters", func() {
			Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(1))
			url, bodyType, body, token := fakeAuthClient.DoAuthenticatedPostArgsForCall(0)
			Expect(url).To(Equal(encryptURI))
			Expect(bodyType).To(Equal("text/plain"))
			Expect(body).To(Equal(plainText))
			Expect(token).To(Equal(accessToken))
		})

		Context("when the config server's /encrypt endpoint returns a non-200 status", func() {
			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(ioutil.NopCloser(strings.NewReader("")), http.StatusNotFound, nil)
			})

			It("reports that encryption failed or is not supported", func() {
				Expect(err).To(MatchError("Encryption failed or is not supported by this config server"))
			})
		})

		Context("when the config server's /encrypt endpoint fails", func() {
			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(nil, 0, testError)
			})

			It("should propagate the error", func() {
				Expect(err.Error()).To(ContainSubstring(errorText))
			})
		})

		Context("when the config server's /encrypt endpoint fails but returns a body containing error details", func() {
			var errorBody = ioutil.NopCloser(strings.NewReader("{error details}"))
			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(ioutil.NopCloser(errorBody), 0, testError)
			})

			It("should propagate an error containing error body content", func() {
				Expect(err.Error()).To(ContainSubstring("{error details}"))
			})
		})

		Context("when the config server's encrypt endpoint returns a response body which cannot be read", func() {
			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(&badReader{readErr: testError}, http.StatusOK, nil)
			})

			It("should propagate the error", func() {
				Expect(err.Error()).To(ContainSubstring(errorText))
			})
		})

		Context("when the config server's encrypt endpoint returns a response body which cannot be closed", func() {
			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(&badCloser{reader: bytes.NewBufferString(cipherText), closeErr: testError}, http.StatusOK, nil)
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

	Describe("EncryptFile", func() {
		var testFile string

		BeforeEach(func() {
			testDir := config.CreateTempDir()
			testFile = config.CreateFile(testDir, "file-to-encrypt.txt")
		})

		JustBeforeEach(func() {
			_, err = encrypter.EncryptFile(serviceRegistryInstance, testFile)
		})

		itAuthenticatesAndResolvesTheServiceInstanceUrl()

		It("calls the config server's /encrypt endpoint with content from a file", func() {
			Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(1))
			url, bodyType, body, token := fakeAuthClient.DoAuthenticatedPostArgsForCall(0)
			Expect(url).To(Equal(encryptURI))
			Expect(bodyType).To(Equal("text/plain"))
			Expect(body).To(Equal("Hello\nWorld\n"))
			Expect(token).To(Equal(accessToken))
		})

		Context("when the given file does not exist", func() {
			BeforeEach(func() {
				testFile = "bogus.txt"
			})

			It("fails", func() {
				Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(0))
				Expect(err.Error()).To(Equal("Error opening file at path bogus.txt : open bogus.txt: no such file or directory"))
			})
		})
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

func (bc *badCloser) Close() error {
	return bc.closeErr
}
