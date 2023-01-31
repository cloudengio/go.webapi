// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICEN	 file.

package openapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Walker represents the interface implemented by all walkers.
type Walker interface {
	Walk(doc *openapi3.T) error
}

type walkerOptions struct {
	followRefs    bool
	visitPrefixes [][]string
	applyPrefix   bool
	trace         bool
}

// WalkerOption represents an option for use when creating a new walker.
type WalkerOption func(o *walkerOptions)

// WalkerFollowRefs controls wether the walker will follow $ref's
// and flatten them in place.
func WalkerFollowRefs(v bool) WalkerOption {
	return func(o *walkerOptions) {
		o.followRefs = v
	}
}

// WalkerVisitPrefix adds a prefix that the walk should call the Visitor
// function for. All other paths will be ignored.
func WalkerVisitPrefix(path ...string) WalkerOption {
	return func(o *walkerOptions) {
		o.visitPrefixes = append(o.visitPrefixes, path)
		o.applyPrefix = true
	}
}

func WalkerTracePaths(v bool) WalkerOption {
	return func(o *walkerOptions) {
		o.trace = v
	}
}

// Visitor is called for every node in the walk. It returns true for the
// walk to continue, false otherwise. The walk will stop when an error is
// returned.
type Visitor func(path []string, parent, node any) (ok bool, err error)

// NewWalker returns a Walker that will visit every node in an openapi3 document.
func NewWalker(v Visitor, opts ...WalkerOption) Walker {
	w := &nodeWalker{visitor: v}
	for _, opt := range opts {
		opt(&w.opts)
	}
	return w
}

type nodeWalker struct {
	opts    walkerOptions
	visitor Visitor
}

// returns true if b is a prefix of a
func prefixMatch(a, b []string) bool {
	if len(b) > len(a) {
		return false
	}
	for i := range b {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (wn nodeWalker) visit(path []string, parent, node any) (ok bool, err error) {
	if wn.opts.trace {
		fmt.Println(strings.Join(path, ":"))
	}
	if wn.opts.applyPrefix {
		match := false
		for _, p := range wn.opts.visitPrefixes {
			if prefixMatch(path, p) {
				match = true
				break
			}
		}
		if !match {
			return true, nil
		}
	}
	return wn.visitor(path, parent, node)
}

func (wn nodeWalker) Walk(doc *openapi3.T) error {
	if doc.Info != nil {
		ok, err := wn.visit([]string{"info"}, doc, doc.Info)
		if !ok || err != nil {
			return err
		}
	}
	ok, err := wn.components([]string{"components"}, doc, doc.Components)
	if !ok || err != nil {
		return err
	}
	ok, err = wn.paths([]string{"paths"}, doc, &doc.Paths)
	if err != nil {
		return err
	}
	if ok, err = wn.servers([]string{"servers"}, doc, &doc.Servers); !ok || err != nil {
		return err
	}
	if ok, err = wn.securityReqs([]string{"security"}, doc, &doc.Security); !ok || err != nil {
		return err
	}
	if ok, err = wn.visit([]string{"externalDocs"}, doc, doc.ExternalDocs); !ok || err != nil {
		return err
	}
	if ok, err = wn.tags([]string{"tags"}, doc, &doc.Tags); !ok || err != nil {
		return err
	}
	return nil
}

func (wn nodeWalker) tags(path []string, parent any, tags *openapi3.Tags) (ok bool, err error) {
	if tags == nil || len(*tags) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, tags); err != nil {
		return
	}
	for i, tag := range *tags {
		if ok, err = wn.tag(append(path, strconv.Itoa(i)), parent, tag); err != nil {
			return
		}
	}
	return false, nil
}

func (wn nodeWalker) tag(path []string, parent any, tag *openapi3.Tag) (ok bool, err error) {
	if tag == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, tag); err != nil {
		return
	}
	return wn.externalDocs(append(path, "externalDocs"), parent, tag.ExternalDocs)
}

func (wn nodeWalker) externalDocs(path []string, parent any, edocs *openapi3.ExternalDocs) (ok bool, err error) {
	if edocs == nil {
		return true, nil
	}
	return wn.visit(path, parent, edocs)
}

func (wn nodeWalker) securityReqs(path []string, parent any, reqs *openapi3.SecurityRequirements) (ok bool, err error) {
	if reqs == nil || len(*reqs) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, reqs); err != nil {
		return
	}
	for i, req := range *reqs {
		if ok, err = wn.visit(append(path, strconv.Itoa(i)), reqs, req); err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) paths(path []string, parent any, paths *openapi3.Paths) (ok bool, err error) {
	if paths == nil || len(*paths) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, paths); !ok || err != nil {
		return
	}
	for p, pi := range *paths {
		if ok, err = wn.pathItem(append(path, p), paths, pi); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) pathItem(path []string, parent any, pi *openapi3.PathItem) (ok bool, err error) {
	if pi == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, pi); !ok || err != nil {
		return
	}
	for _, ops := range []struct {
		name string
		op   *openapi3.Operation
	}{
		{"connect", pi.Connect},
		{"delete", pi.Delete},
		{"get", pi.Get},
		{"head", pi.Head},
		{"options", pi.Options},
		{"patch", pi.Patch},
		{"post", pi.Post},
		{"put", pi.Put},
		{"trace", pi.Trace},
	} {
		if ok, err = wn.operation(append(path, ops.name), pi, ops.op); !ok || err != nil {
			return
		}
	}
	if ok, err = wn.servers(append(path, "servers"), pi, &pi.Servers); !ok || err != nil {
		return
	}
	return wn.parameters(append(path, "parameters"), pi, &pi.Parameters)
}

func (wn nodeWalker) operation(path []string, parent any, op *openapi3.Operation) (ok bool, err error) {
	if op == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, op); !ok || err != nil {
		return
	}
	if ok, err = wn.parameters(append(path, "parameters"), op, &op.Parameters); !ok || err != nil {
		return
	}
	if ok, err = wn.servers(append(path, "servers"), op, op.Servers); !ok || err != nil {
		return
	}
	if ok, err = wn.responses(append(path, "responses"), op, &op.Responses); !ok || err != nil {
		return
	}
	if ok, err = wn.securityReqs(append(path, "security"), op, op.Security); !ok || err != nil {
		return
	}
	if ok, err = wn.requestBodyRef(append(path, "requestBody"), op, op.RequestBody); !ok || err != nil {
		return
	}
	if ok, err = wn.callbacks(append(path, "callbacks"), op, &op.Callbacks); !ok || err != nil {
		return
	}
	return wn.externalDocs(append(path, "externalDocs"), op, op.ExternalDocs)
}

func (wn nodeWalker) callbacks(path []string, parent any, cbs *openapi3.Callbacks) (ok bool, err error) {
	if cbs == nil || len(*cbs) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, cbs); !ok || err != nil {
		return
	}
	for n, cb := range *cbs {
		if ok, err = wn.callbackRef(append(path, n), parent, cb); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) callbackRef(path []string, parent any, req *openapi3.CallbackRef) (ok bool, err error) {
	if req == nil || len(*req.Value) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, req); !ok || err != nil {
		return
	}
	for n, pi := range *req.Value {
		if ok, err = wn.visit(append(path, n), parent, pi); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) responses(path []string, parent any, resps *openapi3.Responses) (ok bool, err error) {
	if resps == nil || len(*resps) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, resps); !ok || err != nil {
		return
	}
	for n, r := range *resps {
		if ok, err = wn.responseRef(append(path, n), parent, r); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) responseRef(path []string, parent any, resp *openapi3.ResponseRef) (ok bool, err error) {
	if resp == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, resp); !ok || err != nil {
		return
	}
	rv := resp.Value
	if ok, err = wn.headers(append(path, "headers"), parent, &rv.Headers); !ok || err != nil {
		return
	}
	if ok, err = wn.content(append(path, "content"), parent, &rv.Content); !ok || err != nil {
		return
	}
	return wn.links(append(path, "links"), parent, &resp.Value.Links)
}

func (wn nodeWalker) links(path []string, parent any, lnks *openapi3.Links) (ok bool, err error) {
	if lnks == nil || len(*lnks) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, lnks); !ok || err != nil {
		return
	}
	for n, lnk := range *lnks {
		if ok, err = wn.linkRef(append(path, n), parent, lnk); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) linkRef(path []string, parent any, lnk *openapi3.LinkRef) (ok bool, err error) {
	if lnk == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, lnk); !ok || err != nil {
		return
	}
	for n, par := range lnk.Value.Parameters {
		if ok, err = wn.visit(append(path, n), lnk, par); !ok || err != nil {
			return
		}
	}
	return wn.server(append(path, "servers"), lnk, lnk.Value.Server)
}

func (wn nodeWalker) servers(path []string, parent any, srvs *openapi3.Servers) (ok bool, err error) {
	if srvs == nil || len(*srvs) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, srvs); !ok || err != nil {
		return
	}
	for i, srv := range *srvs {
		if ok, err = wn.server(append(path, strconv.Itoa(i)), parent, srv); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) server(path []string, parent any, s *openapi3.Server) (ok bool, err error) {
	if s == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, s); !ok || err != nil {
		return
	}
	for v, sv := range s.Variables {
		if ok, err = wn.visit(append(path, v), parent, sv); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) parametersMap(path []string, parent any, pars *openapi3.ParametersMap) (ok bool, err error) {
	if pars == nil || len(*pars) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, pars); !ok || err != nil {
		return
	}
	for n, par := range *pars {
		if ok, err = wn.parameterRef(append(path, n), parent, par); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) parameters(path []string, parent any, pars *openapi3.Parameters) (ok bool, err error) {
	if pars == nil || len(*pars) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, pars); !ok || err != nil {
		return
	}
	for _, par := range *pars {
		if ok, err = wn.parameterRef(append(path, par.Value.Name), parent, par); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) parameter(path []string, parent any, par *openapi3.Parameter) (ok bool, err error) {
	if par == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, par); !ok || err != nil {
		return
	}
	if ok, err = wn.schemaRef(append(path, "schema"), parent, par.Schema); !ok || err != nil {
		return
	}
	if ok, err = wn.examples(append(path, "examples"), parent, &par.Examples); !ok || err != nil {
		return
	}
	return wn.content(append(path, "content"), parent, &par.Content)
}

func (wn nodeWalker) parameterRef(path []string, parent any, pr *openapi3.ParameterRef) (ok bool, err error) {
	if pr == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, pr); !ok || err != nil {
		return
	}
	prv := pr.Value
	if ok, err = wn.schemaRef(append(path, "schema"), parent, prv.Schema); !ok || err != nil {
		return
	}
	if ok, err = wn.examples(append(path, "examples"), parent, &prv.Examples); !ok || err != nil {
		return
	}
	if ok, err = wn.content(append(path, "content"), parent, &prv.Content); !ok || err != nil {
		return
	}
	return true, nil
}

func (wn nodeWalker) examples(path []string, parent any, egs *openapi3.Examples) (ok bool, err error) {
	if egs == nil || len(*egs) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, egs); !ok || err != nil {
		return
	}
	for n, eg := range *egs {
		if ok, err = wn.visit(append(path, n), parent, eg); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) content(path []string, parent any, c *openapi3.Content) (ok bool, err error) {
	if c == nil || len(*c) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, c); !ok || err != nil {
		return
	}
	for n, mt := range *c {
		if ok, err = wn.mediaType(append(path, n), parent, mt); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) mediaType(path []string, parent any, mt *openapi3.MediaType) (ok bool, err error) {
	if mt == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, mt); !ok || err != nil {
		return
	}
	if ok, err = wn.schemaRef(append(path, "schema"), parent, mt.Schema); !ok || err != nil {
		return
	}
	if ok, err = wn.examples(append(path, "examples"), parent, &mt.Examples); !ok || err != nil {
		return
	}
	for n, enc := range mt.Encoding {
		if ok, err = wn.encoding(append(path, n), parent, enc); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) encoding(path []string, parent any, mt *openapi3.Encoding) (ok bool, err error) {
	if mt == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, mt); !ok || err != nil {
		return
	}
	return wn.headers(path, parent, &mt.Headers)
}

func (wn nodeWalker) headers(path []string, parent any, hdrs *openapi3.Headers) (ok bool, err error) {
	if hdrs == nil || len(*hdrs) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, hdrs); !ok || err != nil {
		return
	}
	for n, hdr := range *hdrs {
		if ok, err = wn.headerRef(append(path, n), parent, hdr); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) headerRef(path []string, parent any, hdr *openapi3.HeaderRef) (ok bool, err error) {
	if hdr == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, hdr); !ok || err != nil {
		return
	}
	return wn.parameter(path, parent, &hdr.Value.Parameter)
}

func (wn nodeWalker) components(path []string, parent any, c *openapi3.Components) (ok bool, err error) {
	if c == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, c); !ok || err != nil {
		return
	}
	if ok, err = wn.schemas(append(path, "schemas"), parent, &c.Schemas); !ok || err != nil {
		return
	}
	if ok, err = wn.parametersMap(append(path, "parameters"), parent, &c.Parameters); !ok || err != nil {
		return
	}
	if ok, err = wn.headers(append(path, "headers"), parent, &c.Headers); !ok || err != nil {
		return
	}
	if ok, err = wn.responses(append(path, "responses"), parent, &c.Responses); !ok || err != nil {
		return
	}
	if ok, err = wn.links(append(path, "links"), parent, &c.Links); !ok || err != nil {
		return
	}
	if ok, err = wn.examples(append(path, "examples"), parent, &c.Examples); !ok || err != nil {
		return
	}
	if ok, err = wn.callbacks(append(path, "callbacks"), parent, &c.Callbacks); !ok || err != nil {
		return
	}
	ok, err = wn.securitySchemes(append(path, "securitySchemes"), parent, &c.SecuritySchemes)
	return true, nil
}

func (wn nodeWalker) securitySchemes(path []string, parent any, scs *openapi3.SecuritySchemes) (ok bool, err error) {
	if scs == nil || len(*scs) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, scs); !ok || err != nil {
		return
	}
	for n, sc := range *scs {
		if ok, err = wn.securitySchemeRef(append(path, n), parent, sc); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) securitySchemeRef(path []string, parent any, sr *openapi3.SecuritySchemeRef) (ok bool, err error) {
	if sr == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, sr); !ok || err != nil {
		return
	}
	ok, err = wn.securityScheme(path, parent, sr.Value)
	return true, nil
}

func (wn nodeWalker) securityScheme(path []string, parent any, sr *openapi3.SecurityScheme) (ok bool, err error) {
	if sr == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, sr); !ok || err != nil {
		return
	}
	ok, err = wn.oauthFlows(append(path, "flows"), parent, sr.Flows)
	return true, nil
}

func (wn nodeWalker) oauthFlows(path []string, parent any, flws *openapi3.OAuthFlows) (ok bool, err error) {
	if flws == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, flws); !ok || err != nil {
		return
	}
	if ok, err = wn.oauthFlow(append(path, "implicit"), parent, flws.Implicit); !ok || err != nil {
		return
	}
	if ok, err = wn.oauthFlow(append(path, "password"), parent, flws.Password); !ok || err != nil {
		return
	}
	if ok, err = wn.oauthFlow(append(path, "clientCredentials"), parent, flws.ClientCredentials); !ok || err != nil {
		return
	}
	ok, err = wn.oauthFlow(append(path, "authorizationCode"), parent, flws.AuthorizationCode)
	return true, nil
}

func (wn nodeWalker) oauthFlow(path []string, parent any, flw *openapi3.OAuthFlow) (ok bool, err error) {
	if flw == nil {
		return true, nil
	}
	return wn.visit(path, parent, flw)
}

func (wn nodeWalker) requestBodies(path []string, parent any, rbs *openapi3.RequestBodies) (ok bool, err error) {
	if rbs == nil || len(*rbs) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, &rbs); !ok || err != nil {
		return
	}
	for n, rb := range *rbs {
		if ok, err = wn.requestBodyRef(append(path, n), parent, rb); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) requestBodyRef(path []string, parent any, req *openapi3.RequestBodyRef) (ok bool, err error) {
	if req == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, req); !ok || err != nil {
		return
	}
	return wn.content(append(path, "content"), parent, &req.Value.Content)
}

func (wn nodeWalker) schemas(path []string, parent any, schemas *openapi3.Schemas) (ok bool, err error) {
	if schemas == nil || len(*schemas) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, schemas); !ok || err != nil {
		return
	}
	for name, schema := range *schemas {
		if ok, err = wn.schemaRef(append(path, name), parent, schema); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) schemaRefs(path []string, parent any, srefs *openapi3.SchemaRefs) (ok bool, err error) {
	if srefs == nil || len(*srefs) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, srefs); !ok || err != nil {
		return
	}
	for i, sref := range *srefs {
		if ok, err = wn.schemaRef(append(path, strconv.Itoa(i)), parent, sref); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) schemaRef(path []string, parent any, sref *openapi3.SchemaRef) (ok bool, err error) {
	if sref == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, sref); !ok || err != nil {
		return
	}
	if !wn.opts.followRefs && len(sref.Ref) > 0 {
		return
	}
	if ok, err = wn.schemaRefs(append(path, "oneOf"), sref, &sref.Value.OneOf); !ok || err != nil {
		return
	}
	if ok, err = wn.schemaRefs(append(path, "anyOf"), sref, &sref.Value.AnyOf); !ok || err != nil {
		return
	}
	if ok, err = wn.schemaRefs(append(path, "allOf"), sref, &sref.Value.AllOf); !ok || err != nil {
		return
	}
	if ok, err = wn.schemaRef(append(path, "not"), sref, sref.Value.Not); !ok || err != nil {
		return
	}
	if ok, err = wn.schemas(append(path, "properties"), sref, &sref.Value.Properties); !ok || err != nil {
		return
	}
	if ok, err = wn.schemaRef(append(path, "items"), sref, sref.Value.Items); !ok || err != nil {
		return
	}
	if ok, err = wn.additionalProperties(append(path, "additionalProperties"), sref, sref.Value.AdditionalProperties); !ok || err != nil {
		return
	}
	if ok, err = wn.extensions(append(path, "extensions"), sref, &sref.Value.Extensions); !ok || err != nil {
		return
	}
	return wn.discriminator(append(path, "discriminator"), sref, sref.Value.Discriminator)
}

func (wn nodeWalker) additionalProperties(path []string, parent any, props openapi3.AdditionalProperties) (ok bool, err error) {
	if props.Has == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, props); !ok || err != nil {
		return
	}
	return wn.schemaRef(path, props, props.Schema)
}

func (wn nodeWalker) extensions(path []string, parent any, exts *map[string]interface{}) (ok bool, err error) {
	if exts == nil || len(*exts) == 0 {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, exts); !ok || err != nil {
		return
	}
	for n, ext := range *exts {
		if ok, err = wn.visit(append(path, n), parent, ext); !ok || err != nil {
			return
		}
	}
	return true, nil
}

func (wn nodeWalker) discriminator(path []string, parent any, disc *openapi3.Discriminator) (ok bool, err error) {
	if disc == nil {
		return true, nil
	}
	if ok, err = wn.visit(path, parent, disc); !ok || err != nil {
		return
	}
	if ok, err = wn.visit(append(path, "mapping"), parent, disc.Mapping); !ok || err != nil {
		return
	}
	if ok, err = wn.extensions(path, parent, &disc.Extensions); !ok || err != nil {
		return
	}
	for name, mapping := range disc.Mapping {
		if ok, err = wn.visit(append(path, "mapping", name), parent, mapping); !ok || err != nil {
			return
		}
	}
	return true, nil
}
