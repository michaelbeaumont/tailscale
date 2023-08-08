// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package webui

import (
	"io"
	"net/http"
	"strings"

	"tailscale.com/client/tailscale"
	"tailscale.com/client/tailscale/apitype"
)

var localClient tailscale.LocalClient

type apiHandler struct {
	lc *tailscale.LocalClient
}

func newAPIHandler() *apiHandler {
	return &apiHandler{lc: &localClient}
}

// proxyRequestToLocalAPI proxies the web API request to the localapi.
//
// The web API request path is expected to exactly match a localapi path,
// with prefix /localapi/.
func (h *apiHandler) proxyRequestToLocalAPI(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/localapi/") {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if h.lc == nil {
		http.Error(w, "server has no local client", http.StatusInternalServerError)
		return
	}

	localAPIURL := "http://" + apitype.LocalAPIHost + r.URL.Path
	req, err := http.NewRequestWithContext(r.Context(), r.Method, localAPIURL, r.Body)
	if err != nil {
		http.Error(w, "failed to construct request", http.StatusInternalServerError)
		return
	}

	// Make request to tailscaled localapi.
	resp, err := h.lc.DoLocalRequest(req)
	if err != nil {
		http.Error(w, err.Error(), resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	// Send back to web frontend.
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
