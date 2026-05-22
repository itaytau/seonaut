package urlutils_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stjudewashere/seonaut/internal/urlutils"
)

// Test AbsoluteURL with an html document that does not have a base tag.
func TestAbsoluteURLWithoutBase(t *testing.T) {
	urlStr := "https://example.com/"
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		t.Errorf("error parsing url: %v", err)
	}

	html := strings.NewReader(`
		<html>
			<head></head>
			<body></body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("error parsing html")
	}

	table := []struct {
		linkURL     string
		expectedURL string
	}{
		{"/test.html", "https://example.com/test.html"},
		{"test.html", "https://example.com/test.html"},
		{"/category/test.html", "https://example.com/category/test.html"},
		{"/category/../test.html", "https://example.com/test.html"},
		{"../test.html", "https://example.com/test.html"},
		{"https://example.com/test.html", "https://example.com/test.html"},
		{"https://external.com/test.html", "https://external.com/test.html"},
	}

	for _, u := range table {
		absolute, err := urlutils.AbsoluteURL(u.linkURL, doc, parsedURL)
		if err != nil {
			t.Errorf("absolute url error url %v", err)
		}

		if absolute.String() != u.expectedURL {
			t.Errorf("absolute url does not match expected value. Want %s Got %s", u.expectedURL, absolute)
		}
	}

}

// Test AbsoluteURL with a document that has a base tag.
func TestAbsoluteURLWithBase(t *testing.T) {
	urlStr := "https://example.com/"
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		t.Errorf("error parsing url: %v", err)
	}

	html := strings.NewReader(`
		<html>
			<head>
			<base href="https://example.com/base">
			</head>
			<body></body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Errorf("error parsing html")
	}

	table := []struct {
		linkURL     string
		expectedURL string
	}{
		{"/test.html", "https://example.com/base/test.html"},
		{"test.html", "https://example.com/base/test.html"},
		{"/category/test.html", "https://example.com/base/category/test.html"},
		{"/category/../test.html", "https://example.com/base/test.html"},
		{"../test.html", "https://example.com/test.html"},
		{"https://example.com/test.html", "https://example.com/test.html"},
		{"https://external.com/test.html", "https://external.com/test.html"},
	}

	for _, u := range table {
		absolute, err := urlutils.AbsoluteURL(u.linkURL, doc, parsedURL)
		if err != nil {
			t.Errorf("absolute url error url %v", err)
		}

		if absolute.String() != u.expectedURL {
			t.Errorf("absolute url does not match expected value. Want %s Got %s", u.expectedURL, absolute)
		}
	}
}

// Test AbsoluteURL with Punycode conversion and percent-encoding case normalization.
func TestAbsoluteURLPunycodeAndNormalization(t *testing.T) {
	urlStr := "https://xn--6dbbec0c.xn--4dbrk0ce/" // https://דוגמה.ישראל/
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("error parsing url: %v", err)
	}

	html := strings.NewReader(`
		<html>
			<head></head>
			<body></body>
		</html>`)

	doc, err := htmlquery.Parse(html)
	if err != nil {
		t.Fatalf("error parsing html")
	}

	table := []struct {
		linkURL     string
		expectedURL string
	}{
		// Lowercase percent encoding should be normalized to uppercase percent encoding
		{"/%d7%91%d7%a0%d7%99%d7%99%d7%aa-%d7%aa%d7%a7%d7%a6%d7%99%d7%91-%d7%9c%d7%a2%d7%a1%d7%a7/", "https://xn--6dbbec0c.xn--4dbrk0ce/%D7%91%D7%A0%D7%99%D7%99%D7%AA-%D7%AA%D7%A7%D7%A6%D7%99%D7%91-%D7%9C%D7%A2%D7%A1%D7%A7/"},
		// Uppercase percent encoding remains uppercase
		{"/%D7%91%D7%A0%D7%99%D7%99%D7%AA-%D7%AA%D7%A7%D7%A6%D7%99%D7%91-%D7%9C%D7%A2%D7%A1%D7%A7/", "https://xn--6dbbec0c.xn--4dbrk0ce/%D7%91%D7%A0%D7%99%D7%99%D7%AA-%D7%AA%D7%A7%D7%A6%D7%99%D7%91-%D7%9C%D7%A2%D7%A1%D7%A7/"},
		// Mixed case percent encoding gets normalized to uppercase
		{"/%d7%91%D7%A0%d7%99%d7%99%d7%aa/", "https://xn--6dbbec0c.xn--4dbrk0ce/%D7%91%D7%A0%D7%99%D7%99%D7%AA/"},
		// Hebrew hostname gets converted to Punycode
		{"https://דוגמה.ישראל/שלום", "https://xn--6dbbec0c.xn--4dbrk0ce/%D7%A9%D7%9C%D7%95%D7%9D"},
	}

	for _, u := range table {
		absolute, err := urlutils.AbsoluteURL(u.linkURL, doc, parsedURL)
		if err != nil {
			t.Errorf("absolute url error url %v: %v", u.linkURL, err)
			continue
		}

		if absolute.String() != u.expectedURL {
			t.Errorf("absolute url does not match expected value. Want %s Got %s", u.expectedURL, absolute)
		}
	}
}

