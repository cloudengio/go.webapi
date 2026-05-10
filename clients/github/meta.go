// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package github

import (
	"context"

	"cloudeng.io/webapi/operations"
)

// SSHKeyFingerprints holds the SHA256 fingerprints for each of GitHub's host keys.
type SSHKeyFingerprints struct {
	SHA256RSA     string `json:"SHA256_RSA"`
	SHA256DSA     string `json:"SHA256_DSA"`
	SHA256ECDSA   string `json:"SHA256_ECDSA"`
	SHA256ED25519 string `json:"SHA256_ED25519"`
}

// MetaDomains holds the domain names used by various GitHub services.
type MetaDomains struct {
	Website    []string `json:"website"`
	Codespaces []string `json:"codespaces"`
	Copilot    []string `json:"copilot"`
	Packages   []string `json:"packages"`
}

// Meta is the response from the GET /meta endpoint. IP ranges are in CIDR notation.
type Meta struct {
	VerifiablePasswordAuthentication bool               `json:"verifiable_password_authentication"`
	SSHKeyFingerprints               SSHKeyFingerprints `json:"ssh_key_fingerprints"`
	SSHKeys                          []string           `json:"ssh_keys"`
	Hooks                            []string           `json:"hooks"`
	Web                              []string           `json:"web"`
	API                              []string           `json:"api"`
	Git                              []string           `json:"git"`
	Packages                         []string           `json:"packages"`
	Pages                            []string           `json:"pages"`
	Importer                         []string           `json:"importer"`
	Actions                          []string           `json:"actions"`
	Dependabot                       []string           `json:"dependabot"`
	Domains                          MetaDomains        `json:"domains"`
}

// GetMeta returns GitHub's meta information including IP ranges used by GitHub
// services and SSH host key fingerprints.
func GetMeta(ctx context.Context, opts ...operations.Option) (Meta, error) {
	ep := operations.NewEndpoint[Meta](opts...)
	meta, _, _, err := ep.Get(ctx, APIHost+"/meta")
	return meta, err
}
