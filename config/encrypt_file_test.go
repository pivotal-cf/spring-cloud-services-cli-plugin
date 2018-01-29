package config_test

import (
	"bytes"
	"errors"
	"fmt"
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
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/test_support"
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
		testDir := test_support.CreateTempDir()
		testFile := test_support.CreateFile(testDir, "file-to-encrypt.txt")
		output, err = config.Encrypt(fakeCliConnection, serviceRegistryInstance, "", testFile, fakeAuthClient)

		Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(1))
		url, bodyType, body, token := fakeAuthClient.DoAuthenticatedPostArgsForCall(0)
		Expect(url).To(Equal(encryptURI))
		Expect(bodyType).To(Equal("text/plain"))
		Expect(body).To(Equal("Hello\nWorld\n"))
		Expect(token).To(Equal(accessToken))
	})

	It("should call the config server's /encrypt endpoint with content from a relative path", func() {
		testDir := test_support.CreateTempDir()
		testFile := test_support.CreateFile(testDir, "file-to-encrypt.txt")
		relPath := test_support.GetRelativePath(testFile)
		output, err = config.Encrypt(fakeCliConnection, serviceRegistryInstance, "", relPath, fakeAuthClient)

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

	It("should fail when given a directory", func() {
		testDir := test_support.CreateTempDir()
		output, err = config.Encrypt(fakeCliConnection, serviceRegistryInstance, "", testDir, fakeAuthClient)

		Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(0))
		expectedErr := fmt.Sprintf("Error opening file at path %s : read %s: is a directory", testDir, testDir)
		Expect(err.Error()).To(Equal(expectedErr))
	})

})

// func check(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }
//
// func GetRelativePath(path string) string {
// 	cwd, err := os.Getwd()
// 	relPath, err := filepath.Rel(cwd, path)
// 	check(err)
// 	return relPath
// }
//
// func CreateFile(path string, fileName string) string {
// 	return CreateFileWithMode(path, fileName, os.FileMode(0666))
// }
//
// func CreateFileWithMode(path string, fileName string, mode os.FileMode) string {
// 	fp := filepath.Join(path, fileName)
// 	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
// 	check(err)
// 	defer f.Close()
// 	_, err = f.WriteString("Hello\nWorld\n")
// 	check(err)
// 	return fp
// }
//
// func CreateTempDir() string {
// 	tempDir, err := ioutil.TempDir("/tmp", "scs-cli-")
// 	check(err)
// 	return tempDir
// }
