package KuaiYanSDK

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// HTTP header keys.
const (
	HTTPHeaderAccept      = "Accept"
	HTTPHeaderContentMD5  = "Content-MD5"
	HTTPHeaderContentType = "Content-Type"
	HTTPHeaderDate        = "Date"
	HTTPHeaderUserAgent   = "User-Agent"
)

// HTTP header keys used for Aliyun API gateway signature.
const (
	HTTPHeaderCAPrefix           = "X-Ca-"
	HTTPHeaderCAKey              = "X-Ca-Key"
	HTTPHeaderCANonce            = "X-Ca-Nonce"
	HTTPHeaderCASignature        = "X-Ca-Signature"
	HTTPHeaderCASignatureHeaders = "X-Ca-Signature-Headers"
	HTTPHeaderCASignatureMethod  = "X-Ca-Signature-Method"
	HTTPHeaderCATimestamp        = "X-Ca-Timestamp"
)

// HTTP header content-type values.
const (
	HTTPContentTypeForm                      = "application/x-www-form-urlencoded"
	HTTPContentTypeMultipartForm             = "multipart/form-data"
	HTTPContentTypeMultipartFormWithBoundary = "multipart/form-data; boundary="
	HTTPContentTypeStream                    = "application/octet-stream"
	HTTPContentTypeJson                      = "application/json"
	HTTPContentTypeXml                       = "application/xml"
	HTTPContentTypeText                      = "text/plain"
)

// HTTP method values.
const (
	HTTPMethodGet     = "GET"
	HTTPMethodPost    = "POST"
	HTTPMethodPut     = "PUT"
	HTTPMethodDelete  = "DELETE"
	HTTPMethodPatch   = "PATCH"
	HTTPMethodHead    = "HEAD"
	HTTPMethodOptions = "OPTIONS"
)

// default values.
const (
	defaultUserAgent  = "Go-Aliyun-Sign-Client"
	defaultAccept     = "*/*"
	defaultSignMethod = "HmacSHA256"
	defaultLF         = "\n"
	defaultSep        = ","
)

var signHeaders = map[string]struct{}{
	http.CanonicalHeaderKey(HTTPHeaderCAKey):             {},
	http.CanonicalHeaderKey(HTTPHeaderCANonce):           {},
	http.CanonicalHeaderKey(HTTPHeaderCASignatureMethod): {},
	http.CanonicalHeaderKey(HTTPHeaderCATimestamp):       {},
}

func getSortKeys(m map[string][]string, needFormat ...bool) []string {
	nf := false
	if len(needFormat) > 0 {
		nf = needFormat[0]
	}

	keys := make([]string, 0, len(m))
	for key := range m {
		if nf {
			key = http.CanonicalHeaderKey(key)
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// copyRequest returns the copied request.
func copyRequest(r *http.Request) (*http.Request, error) {
	clone := r.Clone(context.Background())
	if r.Body == nil {
		clone.Body = nil
	} else if r.Body == http.NoBody {
		clone.Body = http.NoBody
	} else if r.GetBody != nil {
		body, err := r.GetBody()
		if err != nil {
			return nil, fmt.Errorf("request get body err: %w", err)
		}

		clone.Body = body
	} else {
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(r.Body); err != nil {
			return nil, fmt.Errorf("read from request body err: %w", err)
		}
		if err := r.Body.Close(); err != nil {
			return nil, fmt.Errorf("request body close err: %w", err)
		}

		r.Body = ioutil.NopCloser(&buf)
		clone.Body = ioutil.NopCloser(bytes.NewBuffer(buf.Bytes()))
	}

	return clone, nil
}

// CurrentTimeMillis returns the millisecond representation of the current time.
func CurrentTimeMillis() string {
	t := time.Now().UnixNano() / 1000000

	return strconv.FormatInt(t, 10)
}

// CurrentGMTDate returns the GMT date representation of the current time.
func CurrentGMTDate() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

// UUID4 returns random generated UUID string.
func UUID4() string {
	u, err := uuid.NewRandom()
	if err != nil {
		return ""
	}

	return u.String()
}

// HmacSHA256 returns the string encrypted with HmacSHA256 method.
func HmacSHA256(b, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(b)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// MD5 returns the string hashed with MD5 method.
func MD5(b []byte) string {
	m := md5.New()
	m.Write(b)

	return base64.StdEncoding.EncodeToString(m.Sum(nil))
}

// Sign will sign the request with appKey and appKeySecret.
func aliyunSign(req *req.Request, appKeySecret, appKey string) error {
	req.SetHeader(HTTPHeaderCAKey, appKey)
	req.SetHeader(HTTPHeaderCANonce, UUID4())
	req.SetHeader(HTTPHeaderCASignatureMethod, defaultSignMethod)
	req.SetHeader(HTTPHeaderCATimestamp, CurrentTimeMillis())

	if req.Headers.Get(HTTPHeaderAccept) == "" {
		req.SetHeader(HTTPHeaderAccept, defaultAccept)
	}

	if req.Headers.Get(HTTPHeaderDate) == "" {
		req.SetHeader(HTTPHeaderDate, CurrentGMTDate())
	}

	if req.Headers.Get(HTTPHeaderUserAgent) == "" {
		req.SetHeader(HTTPHeaderUserAgent, defaultUserAgent)
	}

	ct := req.Headers.Get(HTTPHeaderContentType)
	if req.Body != nil && ct != HTTPContentTypeForm &&
		!strings.HasPrefix(ct, HTTPContentTypeMultipartFormWithBoundary) {
		req.SetHeader(HTTPHeaderContentMD5, MD5(req.Body))
	}

	stringToSign, err := buildStringToSign(req)
	if err != nil {
		return fmt.Errorf("build string to sign err: %w", err)
	}

	req.SetHeader(HTTPHeaderCASignature, HmacSHA256([]byte(stringToSign), []byte(appKeySecret)))

	return nil
}

func buildStringToSign(req *req.Request) (string, error) {
	var s strings.Builder
	s.WriteString(strings.ToUpper(req.Method) + defaultLF)

	s.WriteString(req.Headers.Get(HTTPHeaderAccept) + defaultLF)
	s.WriteString(req.Headers.Get(HTTPHeaderContentMD5) + defaultLF)
	s.WriteString(req.Headers.Get(HTTPHeaderContentType) + defaultLF)
	s.WriteString(req.Headers.Get(HTTPHeaderDate) + defaultLF)
	s.WriteString(buildHeaderStringToSign(req))
	s.WriteString(req.URL.Path + "?" + req.URL.RawQuery)

	return s.String(), nil
}

func buildHeaderStringToSign(req *req.Request) string {
	var builder strings.Builder
	signHeaderKeys := make([]string, 0)
	headerKeys := getSortKeys(req.Headers, true)

	for _, key := range headerKeys {
		if _, ok := signHeaders[key]; ok {
			signHeaderKeys = append(signHeaderKeys, key)
			builder.WriteString(key + ":" + req.Headers.Get(key) + defaultLF)
		}
	}

	req.SetHeader(HTTPHeaderCASignatureHeaders, strings.Join(signHeaderKeys, defaultSep))

	return builder.String()
}
