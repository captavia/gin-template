package utils

import (
	"errors"
	"net/http"
	"testing"
)

func TestJsonResponse(t *testing.T) {
	var errIfc error = &JsonResponse{}
	_ = errIfc

	var nilResp *JsonResponse
	if got := nilResp.Error(); got != EmptyMessage {
		t.Fatalf("nil JsonResponse.Error() = %q, want %q", got, EmptyMessage)
	}

	okResp := Ok(map[string]string{"k": "v"})
	if okResp.Code != CodeSuccess || okResp.Message != SuccessMessage {
		t.Fatalf("Ok() = %+v, want Code=%d Message=%q", okResp, CodeSuccess, SuccessMessage)
	}

	errResp := Err(CodeInvalidParameter, errors.New("bad param"))
	if errResp.Code != CodeInvalidParameter || errResp.Error() != "bad param" {
		t.Fatalf("Err() = %+v, want Code=%d Message=%q", errResp, CodeInvalidParameter, "bad param")
	}

	nilErrResp := Err(CodeError, nil)
	if nilErrResp.Code != CodeError || nilErrResp.Error() != EmptyMessage {
		t.Fatalf("Err(nil) = %+v, want Code=%d Message=%q", nilErrResp, CodeError, EmptyMessage)
	}
}

func TestHTMLResponse(t *testing.T) {
	var errIfc error = &HTMLResponse{}
	_ = errIfc

	var nilResp *HTMLResponse
	if got := nilResp.Error(); got != EmptyMessage {
		t.Fatalf("nil HTMLResponse.Error() = %q, want %q", got, EmptyMessage)
	}

	raw := &HTMLResponse{Code: http.StatusOK, Body: "<p>hello</p>"}
	if got := raw.Error(); got != "<p>hello</p>" {
		t.Fatalf("raw HTMLResponse.Error() = %q, want %q", got, "<p>hello</p>")
	}

	tpl := &HTMLResponse{Code: http.StatusOK, Template: "index.tmpl", IsTemplate: true}
	if got := tpl.Error(); got != "index.tmpl" {
		t.Fatalf("template HTMLResponse.Error() = %q, want %q", got, "index.tmpl")
	}

	res := HTML[string]("<h1>hi</h1>")
	if !res.IsError() {
		t.Fatal("HTML() expected error-wrapped mo.Result")
	}
	htmlResp, ok := errors.AsType[*HTMLResponse](res.Error())
	if !ok {
		t.Fatalf("HTML() error type = %T, want *HTMLResponse", res.Error())
	}
	if htmlResp.Code != http.StatusOK || htmlResp.Body != "<h1>hi</h1>" || htmlResp.IsTemplate {
		t.Fatalf("HTML() = %+v, unexpected fields", htmlResp)
	}

	tplRes := HTMLTemplate[int]("page.tmpl", map[string]string{"title": "test"})
	if !tplRes.IsError() {
		t.Fatal("HTMLTemplate() expected error-wrapped mo.Result")
	}
	tplResp, ok := errors.AsType[*HTMLResponse](tplRes.Error())
	if !ok {
		t.Fatalf("HTMLTemplate() error type = %T, want *HTMLResponse", tplRes.Error())
	}
	if tplResp.Code != http.StatusOK || tplResp.Template != "page.tmpl" || !tplResp.IsTemplate {
		t.Fatalf("HTMLTemplate() = %+v, unexpected fields", tplResp)
	}
}

func TestRedirectResponse(t *testing.T) {
	var errIfc error = &RedirectResponse{}
	_ = errIfc

	var nilResp *RedirectResponse
	if got := nilResp.Error(); got != EmptyMessage {
		t.Fatalf("nil RedirectResponse.Error() = %q, want %q", got, EmptyMessage)
	}

	r := &RedirectResponse{Code: http.StatusMovedPermanently, Location: "/new"}
	if got := r.Error(); got != "/new" {
		t.Fatalf("RedirectResponse.Error() = %q, want %q", got, "/new")
	}

	res := Redirect[struct{}]("https://example.com/login")
	if !res.IsError() {
		t.Fatal("Redirect() expected error-wrapped mo.Result")
	}
	redirectResp, ok := errors.AsType[*RedirectResponse](res.Error())
	if !ok {
		t.Fatalf("Redirect() error type = %T, want *RedirectResponse", res.Error())
	}
	if redirectResp.Code != http.StatusFound || redirectResp.Location != "https://example.com/login" {
		t.Fatalf("Redirect() = %+v, unexpected fields", redirectResp)
	}
}
