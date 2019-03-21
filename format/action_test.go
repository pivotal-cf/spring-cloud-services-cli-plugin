/*
 * Copyright 2016-2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package format_test

import (
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"

	"bytes"
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunAction", func() {

	const (
		testMessage = "some message"
		failMessage = "FAILED"
		certHint    = "Hint: try --skip-ssl-validation at your own risk.\n"
	)

	var (
		ok                = format.Bold(format.Green("OK"))
		fakeCliConnection *pluginfakes.FakeCliConnection
		action            func() (string, error)
		onFailure         func()
		output            string
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}

		fakeCliConnection.GetCurrentOrgStub = func() (plugin_models.Organization, error) {
			return plugin_models.Organization{
				OrganizationFields: plugin_models.OrganizationFields{
					Name: "someOrg",
				},
			}, nil
		}

		fakeCliConnection.GetCurrentSpaceStub = func() (plugin_models.Space, error) {
			return plugin_models.Space{
				SpaceFields: plugin_models.SpaceFields{
					Name: "someSpace",
				},
			}, nil
		}

		fakeCliConnection.UsernameStub = func() (string, error) {
			return "someUser", nil
		}

		action = func() (string, error) {
			return "", nil
		}

		onFailure = func() {}
	})

	JustBeforeEach(func() {
		writer := &bytes.Buffer{}
		format.RunAction(fakeCliConnection, testMessage, action, writer, onFailure)
		output = writer.String()
	})

	It("should print a suitable progress message", func() {
		Expect(output).To(Equal(testMessage + fmt.Sprintf(" in org %s / space %s as %s...\n%s\n\n",
			format.Bold(format.Cyan("someOrg")), format.Bold(format.Cyan("someSpace")), format.Bold(format.Cyan("someUser")), ok)))
	})

	Context("when no org is targetted", func() {
		BeforeEach(func() {
			fakeCliConnection.GetCurrentOrgStub = func() (plugin_models.Organization, error) {
				return plugin_models.Organization{}, errors.New("Org not targetted")
			}
		})

		It("should print output with no progress message", func() {
			Expect(output).To(Equal(fmt.Sprintf("%s\n\n", ok)))
		})
	})

	Context("when no space is targetted", func() {
		BeforeEach(func() {
			fakeCliConnection.GetCurrentSpaceStub = func() (plugin_models.Space, error) {
				return plugin_models.Space{}, errors.New("Space not targetted")
			}
		})

		It("should print output with no progress message", func() {
			Expect(output).To(Equal(fmt.Sprintf("%s\n\n", ok)))
		})
	})

	Context("when no user is logged in", func() {
		Context("when Username returns an error", func() {
			BeforeEach(func() {
				fakeCliConnection.UsernameStub = func() (string, error) {
					return "", errors.New("user not logged in")
				}
			})

			It("should print output with no progress message", func() {
				Expect(output).To(Equal(fmt.Sprintf("%s\n\n", ok)))
			})
		})

		Context("when Username returns an empty string", func() {
			BeforeEach(func() {
				fakeCliConnection.UsernameStub = func() (string, error) {
					return "", nil
				}
			})

			It("should print output with no progress message", func() {
				Expect(output).To(Equal(fmt.Sprintf("%s\n\n", ok)))
			})
		})
	})

	Context("when the action fails", func() {
		BeforeEach(func() {
			action = func() (string, error) {
				return "", errors.New("Fake Error")
			}
		})

		It("should print a failure message", func() {
			Expect(output).To(ContainSubstring(failMessage))
		})
	})

	Context("when the action fails with a certificate error", func() {
		BeforeEach(func() {
			action = func() (string, error) {
				return "", errors.New("Error: unknown authority")
			}
		})

		It("should print a suitable hint", func() {
			Expect(output).To(ContainSubstring(certHint))
		})
	})
})
