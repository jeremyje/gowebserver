// Copyright 2022 Jeremy Edwards
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpprobe

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ProbeError struct {
	Code    int
	Message string
}

func (p ProbeError) Error() string {
	return fmt.Sprintf("HTTP Code: %d - %s", p.Code, p.Message)
}

type Args struct {
	CertificatePool *x509.CertPool
	Timeout         time.Duration
	URL             string
}

func Probe(args Args) error {
	u := normalizeURL(args.URL)

	t := &http.Transport{}
	if args.CertificatePool != nil {
		t = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: args.CertificatePool,
			},
		}
	}

	c := &http.Client{
		Timeout:   args.Timeout,
		Transport: t,
	}

	resp, err := c.Get(u)
	if err != nil {
		return ProbeError{
			Code:    http.StatusServiceUnavailable,
			Message: fmt.Sprintf("%s is not available, %v", u, err),
		}
	}

	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return nil
	}
	return ProbeError{
		Code:    resp.StatusCode,
		Message: resp.Status,
	}
}

func normalizeURL(urlString string) string {
	u, err := url.Parse(urlString)
	if err != nil {
		return urlString
	}
	return u.String()
}
