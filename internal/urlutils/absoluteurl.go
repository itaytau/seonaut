package urlutils

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"golang.org/x/net/idna"
)

var percentHexRegex = regexp.MustCompile(`%[0-9a-fA-F]{2}`)

// normalizePercentEncoding converts percent-encoded hex characters to uppercase.
func normalizePercentEncoding(s string) string {
	return percentHexRegex.ReplaceAllStringFunc(s, func(m string) string {
		return strings.ToUpper(m)
	})
}

// NormalizeURLString parses, converts the host to Punycode, normalizes percent encoding,
// and returns the canonical URL string.
func NormalizeURLString(urlStr string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(urlStr))
	if err != nil {
		return "", err
	}

	if u.Host != "" {
		asciiHost, err := idna.ToASCII(u.Host)
		if err == nil {
			u.Host = asciiHost
		}
	}

	if u.RawPath != "" {
		u.RawPath = normalizePercentEncoding(u.RawPath)
	}
	u.Path = normalizePercentEncoding(u.Path)
	u.RawQuery = normalizePercentEncoding(u.RawQuery)
	u.Fragment = normalizePercentEncoding(u.Fragment)

	return u.String(), nil
}

// Return an absolute URL removing the URL fragment and taking into account
// the document's base tag if it exists. Absolute URLs are created taking into account
// the currentURL which should be absolute.
func AbsoluteURL(urlStr string, n *html.Node, currentURL *url.URL) (*url.URL, error) {
	if n == nil {
		return nil, errors.New("urlutils: empty node")
	}

	if currentURL == nil {
		return nil, errors.New("urlutils: empty url")
	}

	u, err := url.Parse(strings.TrimSpace(urlStr))
	if err != nil {
		return nil, err
	}

	if !u.IsAbs() {
		base, err := htmlBase(n, currentURL)
		if err != nil {
			u = currentURL.ResolveReference(u)
		} else {
			u = base.JoinPath(u.Path)
		}
	}

	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("protocol not supported")
	}

	normalizedStr, err := NormalizeURLString(u.String())
	if err != nil {
		return nil, err
	}

	normalizedURL, err := url.Parse(normalizedStr)
	if err != nil {
		return nil, err
	}

	return normalizedURL, nil
}

// htmlBase returns the url in the base tag if it exists. Otherwise it returns an error.
func htmlBase(n *html.Node, currentURL *url.URL) (*url.URL, error) {
	base, err := htmlquery.Query(n, "//head/base[@href]")
	if err != nil || base == nil {
		return nil, errors.New("base path is missing or empty")
	}

	href := strings.TrimSpace(htmlquery.SelectAttr(base, "href"))
	parsed, err := url.Parse(href)
	if err != nil {
		return nil, errors.New("error parsing base path")
	}

	if parsed.Host == "" {
		parsed.Host = currentURL.Host
	}

	if parsed.Scheme == "" {
		parsed.Scheme = currentURL.Scheme
	}

	return parsed, nil
}
