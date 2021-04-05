package p_test

import (
	p "example.com/cloudfunction"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Webhook", func() {
	var (
		request    string
		testUrl    string
		testApiUrl string
		testHook   p.Hook
	)

	BeforeEach(func() {
		request = "test string"
		testUrl = "https://build-api.cloud.unity3d.com/api/url"
		testApiUrl = "/api/url"
		testHook = p.Hook{
			LinkList: p.Links{
				Url: p.Href{
					Url:    testApiUrl,
					Method: "get",
				},
			},
		}
	})

	Describe("Category of webhook tests", func() {
		It("Should successfully parse out download url", func() {
			Expect(request).To(Equal("test string"))
		})

		It("Should build curl URL from webhook json", func() {
			result := p.ConstructUrl(testHook)

			Expect(result).To(Equal(testUrl))
		})
	})
})
