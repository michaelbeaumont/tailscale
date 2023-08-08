// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

// Package webui provides the Tailscale client for web.
package webui

import (
	"fmt"
	"net/http"
	"strings"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world")

	apiHandler := newAPIHandler()
	switch {
	case strings.HasPrefix(r.URL.Path, "/localapi/"):
		apiHandler.proxyRequestToLocalAPI(w, r)

		// TODO: Maybe another case for non-proxied requests,
		// if we ever have any.
		// case strings.HasPrefix(r.URL.Path, "/api/"):
		// 	apiHandler.serveAPI(w, r)
	}
}
