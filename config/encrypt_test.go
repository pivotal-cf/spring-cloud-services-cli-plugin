package config_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"io"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/config"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("Encrypt", func() {

	const (
		serviceRegistryInstance = "some-service-registry"
		serviceRegistryKey      = "some-service-registry-key"
		plainText               = "plain-text"
		errorText               = "to err is human"
		statusMessage           = "server error"
		serviceKey              = "somekey"
		accessToken             = "access-token"
		serviceURI              = "service-uri"
		encryptURI              = "service-uri/encrypt"
		cipherText              = "cipher-text"
	)

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient
		testError         error
		postResponse      io.ReadCloser
		postStatusCode    int
		postErr           error
		output            string
		err               error
		serviceKeysOutput []string
		serviceKeysErr    error
		serviceKeyOutput  []string
		serviceKeyErr     error
		accessTokenURI    string
		clientID          string
		clientSecret      string
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		testError = errors.New(errorText)
		postResponse, postStatusCode, postErr = ioutil.NopCloser(bytes.NewBufferString(cipherText)), http.StatusOK, nil
		fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString("https://fake.com")), 200, nil)
		serviceKeysOutput, serviceKeysErr = []string{"", "", "", serviceKey}, nil
		accessTokenURI = "access-token-uri"
		clientID = "client-id"
		clientSecret = "client-secret"
		serviceKeyOutput, serviceKeyErr = strings.Split(fmt.Sprintf(`Getting key ...

{
 "access_token_uri": "%s",
 "client_id": "%s",
 "client_secret": "%s",
 "uri": "%s"
}`, accessTokenURI, clientID, clientSecret, serviceURI), "\n"), nil
		fakeAuthClient.GetClientCredentialsAccessTokenReturns(accessToken, nil)
	})

	JustBeforeEach(func() {
		fakeAuthClient.DoAuthenticatedPostReturns(postResponse, postStatusCode, postErr)
		fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
			switch args[0] {
			case "service-keys":
				return serviceKeysOutput, serviceKeysErr
			case "service-key":
				return serviceKeyOutput, serviceKeyErr
			default:
				Fail("stub detected unexpected cf operation")
				return []string{}, nil
			}
		}

		output, err = config.Encrypt(fakeCliConnection, serviceRegistryInstance, plainText, fakeAuthClient)
	})

	It("should obtain the service keys associated with the service instance", func() {
		Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(2))
		args := fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)
		Expect(args).To(Equal([]string{"service-keys", serviceRegistryInstance}))
	})

	Context("when obtaining the service keys fails", func() {
		BeforeEach(func() {
			serviceKeysOutput, serviceKeysErr = []string{}, testError
		})

		It("should propagate the error", func() {
			Expect(err.Error()).To(ContainSubstring(errorText))
		})
	})

	Context("when obtaining the service keys returns unexpected output", func() {
		BeforeEach(func() {
			serviceKeysOutput, serviceKeysErr = []string{}, nil
		})

		It("should use the default value of the service key", func() {
			Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(2))
			args := fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(1)
			Expect(args).To(Equal([]string{"service-key", serviceRegistryInstance, serviceRegistryKey}))
		})
	})

	It("should use the service key returned by cf service-keys", func() {
		Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(2))
		args := fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(1)
		Expect(args).To(Equal([]string{"service-key", serviceRegistryInstance, serviceKey}))
	})

	Context("when obtaining the service key information fails", func() {
		BeforeEach(func() {
			serviceKeyOutput, serviceKeyErr = []string{}, testError

		})

		It("should propagate the error", func() {
			Expect(err.Error()).To(ContainSubstring(errorText))
		})
	})

	Context("when the service key information has too few lines", func() {
		BeforeEach(func() {
			serviceKeyOutput, serviceKeyErr = []string{}, nil
		})

		It("should return a suitable error", func() {
			Expect(err).To(MatchError("Malformed service key info: []"))
		})
	})

	Context("when the service key information contains invalid JSON", func() {
		BeforeEach(func() {
			serviceKeyOutput, serviceKeyErr = []string{"", "", ""}, nil
		})

		It("should return a suitable error", func() {
			Expect(err.Error()).To(ContainSubstring("Failed to unmarshal service key info:"))
		})
	})

	It("should create an access token using the correct parameters", func() {
		Expect(fakeAuthClient.GetClientCredentialsAccessTokenCallCount()).To(Equal(1))
		uri, id, secret := fakeAuthClient.GetClientCredentialsAccessTokenArgsForCall(0)
		Expect(uri).To(Equal(accessTokenURI))
		Expect(id).To(Equal(clientID))
		Expect(secret).To(Equal(clientSecret))
	})

	Context("when creating an access token fails", func() {
		BeforeEach(func() {
			fakeAuthClient.GetClientCredentialsAccessTokenReturns("", testError)
		})

		It("should propagate the error", func() {
			Expect(err.Error()).To(ContainSubstring(errorText))
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
