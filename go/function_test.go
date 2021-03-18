package p_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Webhook", func() {
	var (
		request string
	)

	BeforeEach(func() {
		request = "test string"
	})

	Describe("Category of webhook tests", func() {
		It("Should successfully parse out download url", func() {
			Expect(request).To(Equal("test string"))
		})
	})
})
