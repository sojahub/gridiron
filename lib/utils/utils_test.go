// SPDX-License-Identifier: Apache-2.0
//

package utils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lib/utils")
}
