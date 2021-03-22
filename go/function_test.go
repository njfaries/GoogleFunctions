package p_test

import (
	p "example.com/cloudfunction"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Webhook", func() {
	var (
		request  string
		testUrl  string
		testHook p.Hook
	)

	BeforeEach(func() {
		request = "test string"
		testUrl = "https://www.test.url"
		testHook = p.Hook{
			LinkList: p.Links{
				Url: testUrl,
			},
		}
	})

	Describe("Category of webhook tests", func() {
		It("Should successfully parse out download url", func() {
			Expect(request).To(Equal("test string"))
		})

		It("Should extract url from webhook json", func() {
			result := p.ExtractUrl(testHook)

			Expect(result).To(Equal(testUrl))
		})
	})
})
