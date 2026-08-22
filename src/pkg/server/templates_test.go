package server

import (
	"io/fs"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// templateRefPattern matches every {{template "name"}} and {{block "name"}}
// reference in a template source file.
var templateRefPattern = regexp.MustCompile(`{{-?\s*(?:template|block)\s+"([^"]+)"`)

// sharedTemplateFiles are the layouts, partials and components parsed into
// every page's template set.
func sharedTemplateFiles(t *testing.T) []string {
	t.Helper()

	var files []string
	for _, dir := range []string{"layouts", "partials", "components"} {
		entries, err := templatesFS.ReadDir("templates/" + dir)
		if err != nil {
			t.Fatalf("read templates/%s: %v", dir, err)
		}
		for _, entry := range entries {
			files = append(files, dir+"/"+entry.Name())
		}
	}
	return files
}

// TestEveryTemplateReferenceResolves walks each page's template set and checks
// that every {{template "name"}} reference it can reach is actually defined.
// A missing definition is not a parse error in Go — it only surfaces at render
// time as a 500, so nothing but an explicit check catches it.
func TestEveryTemplateReferenceResolves(t *testing.T) {
	renderer, err := NewTemplateRenderer(false)
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	pages, err := pageTemplatePaths()
	if err != nil {
		t.Fatalf("pageTemplatePaths: %v", err)
	}
	if len(pages) == 0 {
		t.Fatal("no page templates were discovered")
	}

	shared := sharedTemplateFiles(t)

	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			set, ok := renderer.sets[page]
			if !ok {
				t.Fatalf("no template set was built for %s", page)
			}

			defined := make(map[string]bool)
			for _, tmpl := range set.Templates() {
				defined[tmpl.Name()] = true
			}

			for _, file := range append(append([]string{}, shared...), page) {
				content, err := templatesFS.ReadFile("templates/" + file)
				if err != nil {
					t.Fatalf("read %s: %v", file, err)
				}
				for _, match := range templateRefPattern.FindAllStringSubmatch(string(content), -1) {
					if !defined[match[1]] {
						t.Errorf("%s references template %q, which no template in the %s set defines", file, match[1], page)
					}
				}
			}
		})
	}
}

// TestEachPageOwnsItsContentBlock guards the per-page template sets. Every page
// defines a block named "content", so a renderer that parsed all pages into one
// namespace would silently serve the last-parsed page's body for every route.
func TestEachPageOwnsItsContentBlock(t *testing.T) {
	renderer, err := NewTemplateRenderer(false)
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	pages, err := pageTemplatePaths()
	if err != nil {
		t.Fatalf("pageTemplatePaths: %v", err)
	}

	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			set := renderer.sets[page]
			if set == nil {
				t.Fatalf("no template set was built for %s", page)
			}

			content := set.Lookup("content")
			if content == nil || content.Tree == nil {
				t.Fatalf("%s does not define a content block", page)
			}

			source, err := templatesFS.ReadFile("templates/" + page)
			if err != nil {
				t.Fatalf("read %s: %v", page, err)
			}

			// The content block registered in this set must have come from
			// this page's own file, not from another page parsed alongside it
			body := content.Tree.Root.String()
			if marker := firstMarkup(string(source)); marker != "" && !strings.Contains(body, marker) {
				t.Errorf("the content block in the %s set does not come from %s (missing %q)", page, page, marker)
			}
		})
	}
}

// firstMarkup returns the first non-blank, non-action line of a page template,
// used as a fingerprint identifying which file a content block came from.
func firstMarkup(source string) string {
	for _, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "{{") {
			continue
		}
		return line
	}
	return ""
}

// staticRefPattern matches every /static/... asset URL referenced by a
// template.
var staticRefPattern = regexp.MustCompile(`["'](/static/[^"'?#]+)`)

// TestStaticAssetsResolve checks that every asset a template links to exists in
// the embedded filesystem at the path the static file server will look it up
// under. A mismatch between the embed root and the stripped request prefix
// makes every stylesheet and script 404 without failing the build.
func TestStaticAssetsResolve(t *testing.T) {
	pages, err := pageTemplatePaths()
	if err != nil {
		t.Fatalf("pageTemplatePaths: %v", err)
	}

	seen := make(map[string]bool)
	for _, file := range append(sharedTemplateFiles(t), pages...) {
		content, err := templatesFS.ReadFile("templates/" + file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, match := range staticRefPattern.FindAllStringSubmatch(string(content), -1) {
			ref := match[1]
			if seen[ref] {
				continue
			}
			seen[ref] = true

			// The file server strips /static/ before looking the path up
			lookup := strings.TrimPrefix(ref, "/static/")
			if _, err := fs.Stat(staticRoot, lookup); err != nil {
				t.Errorf("%s references %s, which does not exist in the embedded static tree: %v", file, ref, err)
			}
		}
	}

	if len(seen) == 0 {
		t.Fatal("no static asset references were found in the templates")
	}
}

// TestRenderSettingsPage renders the admin settings page against the same
// typed view-model the handler builds. The page previously received an
// untyped map whose keys did not match the fields the template read, so it
// returned 500 on every request without failing the build.
func TestRenderSettingsPage(t *testing.T) {
	renderer, err := NewTemplateRenderer(false)
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	data := SettingsData{
		PageData: PageData{Title: "Settings", SiteName: "CASCI", AdminPath: "administration", APIVersion: "v1"},
		Config: SettingsView{
			Server: SettingsServerView{
				Title:      "CASCI",
				FQDN:       "ci.example.com",
				Port:       8080,
				Address:    "[::]",
				Mode:       "production",
				AdminPath:  "administration",
				APIVersion: "v1",
				SSL:        SettingsSSLView{Enabled: true, MinVersion: "TLS1.2"},
				Database:   SettingsDatabaseView{Driver: "sqlite", MaxConns: 25, MaxIdleConn: 5},
			},
			Web: SettingsWebView{CORS: "*", CookieConsentEnabled: true},
		},
	}

	rec := httptest.NewRecorder()
	if err := renderer.Render(rec, "layouts/admin.tmpl", "admin/settings.tmpl", data); err != nil {
		t.Fatalf("Render(admin/settings.tmpl): %v", err)
	}

	body := rec.Body.String()
	for _, want := range []string{"ci.example.com", "administration", "TLS1.2", "sqlite"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered settings page is missing %q", want)
		}
	}
}

// TestRenderUnknownPageFails ensures a mistyped page name is reported rather
// than silently rendering an empty body.
func TestRenderUnknownPageFails(t *testing.T) {
	renderer, err := NewTemplateRenderer(false)
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	rec := httptest.NewRecorder()
	if err := renderer.Render(rec, "layouts/base.tmpl", "pages/does-not-exist.tmpl", PageData{}); err == nil {
		t.Fatal("Render accepted an unknown page template")
	}
}

// TestNoInlineEventHandlers keeps the templates compatible with the nonce CSP,
// which allows neither inline event handlers nor eval.
func TestNoInlineEventHandlers(t *testing.T) {
	pages, err := pageTemplatePaths()
	if err != nil {
		t.Fatalf("pageTemplatePaths: %v", err)
	}

	files := append(sharedTemplateFiles(t), pages...)
	for _, file := range files {
		content, err := templatesFS.ReadFile("templates/" + file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, attr := range []string{"onclick=", "onchange=", "onsubmit=", "onkeypress=", "onkeydown=", "onload="} {
			if strings.Contains(string(content), attr) {
				t.Errorf("%s contains inline handler %q, which the CSP blocks", file, attr)
			}
		}
	}
}
