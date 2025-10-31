# go-htmltest

**純粋なHTMLに対して “人間の操作” を自動テストするための Go 向け補助ライブラリ**。

* 既存のHTML文字列（SSRの出力、テンプレート、ファイル、スナップショット）を読み込み
* フォームの入力/送信、リンク/ボタンのクリックをエミュレート
* 見出し/リスト/テーブル/任意テキストの存在確認
* **JavaScriptは実行しません**（純HTML + Web標準仕様に基づく動作に限定）

---

## 目次

* [特徴](#特徴)
* [制限事項](#制限事項)
* [クイックスタート](#クイックスタート)
* [フォーム操作例](#フォーム操作例)
* [リンク例](#リンク例)
* [Cookie の扱い](#cookie-の扱い)
* [Selector について](#selector-について)
* [Transport について](#transport-について)
* [API 概要](#api-概要)
* [フレームワーク統合](#フレームワーク統合)
* [ロードマップ](#ロードマップ)

---

## 特徴

* **シンプルなAPI**: フォーム入力やリンククリックを直感的に操作可能
* **Transport抽象化**: 任意のHTTPクライアント/サーバーと連携可能
* **Cookie自動管理**: セッション維持や認証シナリオをサポート
* **テストフレームワーク統合**: `testing.T` とシームレスに連携
* **軽量**: 依存関係が少なく、セットアップが簡単

## 制限事項

* **JavaScript非対応**: 純粋なHTML + Web標準仕様に基づく動作に限定
* **CSS/レイアウト未評価**: 見た目の崩れ検出は不可（文字列/DOM構造に限定）
* **ブラウザ差異非再現**: ブラウザ実機に依存する挙動は対象外

もしJavaScript実行が必要な場合は、`playwright-go`や`chromedp`などのヘッドレスブラウザ操作ライブラリを検討してください。



## クイックスタート

```go
package login_test

import (
    "net/http"
    "testing"

    h "github.com/yourname/go-htmltest"
)

type MockTransport struct{}

func (m MockTransport) Do(req h.Request) (h.Response, error) {
    // 送信されたフォーム/URLに応じて任意のレスポンスを返す
    if req.Method == http.MethodPost && req.URL.Path == "/login" {
        user := req.Form.Get("username")
        return h.Response{Status: 200, Body: "<p>Welcome, " + user + "</p>"}, nil
    }
    return h.Response{Status: 404, Body: "not found"}, nil
}

func Test_LoginFlow(t *testing.T) {
    html := `
    <form id="login-form" action="/login" method="post">
      <label>User</label><input type="text" name="username">
      <label>Pass</label><input type="password" name="password">
      <button type="submit">Login</button>
    </form>`

    dom := h.New(MockTransport{}).ParseString(html)

    form := dom.Form("#login-form")
    form.Fill("username", "alice").Fill("password", "secret")

    res := form.Submit(t)
    if res.Status != 200 {
        t.Fatalf("expected status 200, got %d", res.Status)
    }
}
```

## フォーム操作例

```go
form := dom.Form("@login").
     Fill("username", "tester").
     Fill("password", "mypassword").
     Check("remember_me")
     
res := form.Submit(t)
if res.Status != 200 {
    t.Fatalf("expected status 200, got %d", res.Status)
}
```

## リンク例

```go
link := dom.Link("@profile")
if link.GetText() != "View Profile" {
    t.Fatalf("expected link text 'View Profile', got '%s'", link.GetText())
}
if link.GetURL().Path != "/users/123" {
    t.Fatalf("expected link URL '/users/123', got '%s'", link.GetURL().Path)
}

res := link.Click()

if res.Status != 200 {
    t.Fatalf("expected status 200, got %d", res.Status)
}

```


## Cookie の扱い


`go-htmltest`はResponseに含まれるCookieを自動的に保存し、以降のフォーム送信やリンククリック時に適切に送信します。これにより、セッション管理やユーザー認証のシナリオを簡単にテストできます。

事前にCookieを設定したい場合は以下のようにします。

```go
dom := h.New(transport).ParseString(html)
dom.SetCookie("session_id", "abc123")

// 以降の操作で "session_id=abc123" が送信される
form := dom.Form("#protected-form").Fill("data", "value")
res := form.Submit(t)
if res.Status != 200 {
    t.Fatalf("expected status 200, got %d", res.Status)
}
```

## Selector について

`selector` には CSS セレクタをベースにしたロケータ文字列を指定します。


| セレクタ | 説明 |
|----------|------|
| @xxxxx | `test-id` 属性が `xxxxx` の要素を特定します。 |
| #xxxxx | `id` 属性が `xxxxx` の要素を特定します。 |


## Transport について

`Transport` は I/O を抽象化するインタフェースです。

`Form.Submit()`や`Link.Click()`は内部で`Transport.Do()`を呼び出し、HTTPリクエストを送信します。
そして、その結果を`Form.Submit()`や`Link.Click()`の呼び出し元に返します。

この仕組みにより`go-htmltest`は特定のHTTPクライアントやサーバーフレームワークに依存せず、任意の実装と連携可能です。


```go
type Request struct {
    Method string
    URL    *url.URL
    Header http.Header
    Form   url.Values    // x-www-form-urlencoded 用
    Files  []FormFile    // multipart 用
}

type Response struct {
    Status int
    Body   string
    Header http.Header
    URL    *url.URL
}

type Transport interface {
    Do(req Request) (Response, error)
}
```

## API 概要

```go
// ルート
func New(transport Transport) *DOM
func (d *DOM) Parse(html string) *DOM
func (d *DOM) SetCookie(name, value string) *DOM

// 要素取得
func (d *DOM) Form(selector string) (*Form, error)
func (d *DOM) Link(selector string) (*Link, error)
func (d *DOM) Button(selector string) (*Button, error)
func (d *DOM) Table(selector string) (*Table, error)
func (d *DOM) List(selector string) (*List, error)
func (d *DOM) Text(selector string) (string, error)
func (d *DOM) Title() (string, error)
func (d *DOM) Meta(name string) (string, error)
func (d *DOM) Img(selector string) (*Img, error)

// フォーム操作
func (f *Form) Fill(name, value string) *Form // テキスト入力
func (f *Form) Check(name string) *Form // チェックボックス/ラジオを選択
func (f *Form) Uncheck(name string) *Form // チェックボックスの選択解除
func (f *Form) Select(name, value string) *Form // セレクトボックス選択
func (f *Form) Choose(name, value string) *Form // ラジオボタン選択
func (f *Form) Submit(t *testing.T) Response // 送信
func (f *Form) GetValue(name string) (string, error) // 入力値取得
func (f *Form) HasField(name string) bool // フィールド存在確認


// リンク
func (l *Link) Click(t *testing.T) Response
func (l *Link) GetURL(t *testing.T) *url.URL
func (l *Link) GetText(t *testing.T) string

// ボタン
func (b *Button) GetText(t *testing.T) string

// テーブル
func (tbl *Table) GetRows() [][]string
func (tbl *Table) GetRowCount() int
func (tbl *Table) GetColCount() int

// リスト
func (lst *List) GetItems() []string
func (lst *List) GetItemCount() int

// 画像
func (img *Img) GetSrc() string
func (img *Img) GetAlt() string

// レスポンス
type Response struct {
    Status int
    Body   string
    Header http.Header
}
```

## フレームワーク統合

メジャーなHTTPサーバーフレームワーク向けのTransport実装を提供しています。

### Chi 統合

Chi フレームワークを使用している場合`ChiTransport`を利用して、`httptest.Server`と連携できます。

```go
import (
    "github.com/go-chi/chi/v5"
    h "github.com/yourname/go-htmltest"
    "github.com/yourname/go-htmltest/integrations/chi_transport"

)

func Test_LoginFlow(t *testing.T) {
    r := chi.NewRouter()

    transport := chi_transport.NewChiTransport(r)

    dom := h.New(transport).ParseString(renderLoginPage())

    form := dom.Form("#login-form").
                Fill("username", "alice").
                Fill("password", "secret")

    res := form.Submit(t)
    if res.Status != 200 {
        t.Fatalf("expected status 200, got %d", res.Status)
    }
}
```

### Echo 統合

準備中...

### Gin 統合

準備中...


### Fiber 統合

準備中...

## ロードマップ

* [ ] `multipart/form-data` とファイル添付
* [ ] `<select multiple>` / `<input type=date|time|number>` の入力補助
* [ ] リダイレクト追従（`3xx`）とCookieサポート（`Set-Cookie`/`Cookie`）
* [ ] `Table`/`List` の差分レポート強化（どのセルが不一致かを詳細表示）
* [ ] セレクタ拡張（`:has()`, `:nth-of-type()` の一部）
* [ ] エラーメッセージの見やすい差分表示（カラー出力）
