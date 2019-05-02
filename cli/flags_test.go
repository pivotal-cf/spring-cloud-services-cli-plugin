/*
 * Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.
 *
 * This program and the accompanying materials are made available under
 * the terms of the under the Apache License, Version 2.0 (the "License”);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cli_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cli"
)

var _ = Describe("Flags", func() {

	var (
		args        = []string{"cf", "srd", "provision-service-registry", "provision-sr-1", "-i", "1", "-i", "2", "-i", "3"}
		instanceIdx *int
		err         error
	)

	JustBeforeEach(func() {
		instanceIdx, _, err = cli.ParseFlags(args)
	})

	Context("when duplicate flags are passed on command line", func() {

		It("the final of the duplicated flags will be chosen", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(*instanceIdx).To(Equal(3))
		})
	})

	Context("when an unexpected flag is received", func() {

		BeforeEach(func() {
			args = []string{"cf", "srd", "provision-service-registry", "provision-sr-1", "-i", "1", "-z"}
		})

		It("should raise a suitable error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Error parsing arguments: Invalid flag: -z"))
		})
	})

	Context("when no flag is passed for instance index", func() {

		BeforeEach(func() {
			args = []string{"cf", "srd", "provision-service-registry", "provision-sr-1"}
		})

		It("should be parsed as nil", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(instanceIdx).To(BeNil())
		})
	})

	Context("when a string value is passed for instance index", func() {

		BeforeEach(func() {
			args = []string{"cf", "srd", "provision-service-registry", "provision-sr-1", "-i", "one"}
		})

		It("should raise a suitable error", func() {
			Expect(err).To(MatchError("Error parsing arguments: Value for flag 'cf-instance-index' must be an integer"))
		})
	})

	Describe("ParseNoFlags", func() {
		var noFlagsPositionalArgs []string

		BeforeEach(func() {
			args = []string{"cf", "csev", "-x", "y", "-z"}
		})

		JustBeforeEach(func() {
			noFlagsPositionalArgs, err = cli.ParseNoFlags(args)
		})

		It("should not treat an argument with a leading hyphen as a flag", func() {
			Expect(noFlagsPositionalArgs).To(ConsistOf("cf", "csev", "-x", "y", "-z"))
		})
	})
})
