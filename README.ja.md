# htmldriver

[![Test](https://github.com/ppdx999/htmldriver/actions/workflows/test.yml/badge.svg)](https://github.com/ppdx999/htmldriver/actions/workflows/test.yml)
[![Lint](https://github.com/ppdx999/htmldriver/actions/workflows/lint.yml/badge.svg)](https://github.com/ppdx999/htmldriver/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ppdx999/htmldriver)](https://goreportcard.com/report/github.com/ppdx999/htmldriver)
[![codecov](https://codecov.io/gh/ppdx999/htmldriver/branch/main/graph/badge.svg)](https://codecov.io/gh/ppdx999/htmldriver)
[![GoDoc](https://godoc.org/github.com/ppdx999/htmldriver?status.svg)](https://godoc.org/github.com/ppdx999/htmldriver)

**純粋なHTMLに対して "人間の操作" を自動テストするための Go 向け補助ライブラリ**。

* 既存のHTML文字列（SSRの出力、テンプレート、ファイル、スナップショット）を読み込み
* フォームの入力/送信、リンク/ボタンのクリックをエミュレート
* 見出し/リスト/テーブル/任意テキストの存在確認
* **JavaScriptは実行しません**（純HTML + Web標準仕様に基づく動作に限定）

[English Documentation](./README.md)

---

## 目次

* [Why htmldriver?](#why-htmldriver)
* [特徴](#特徴)
* [制限事項](#制限事項)
* [クイックスタート](#クイックスタート)
* [フォーム操作例](#フォーム操作例)
* [リンク例](#リンク例)
* [Table / List の確認例](#table--list-の確認例)
* [連続した操作](#連続した操作)
* [Cookie の扱い](#cookie-の扱い)
* [Selector について](#selector-について)
* [Transport について](#transport-について)
* [API 概要](#api-概要)
* [フレームワーク統合](#フレームワーク統合)
* [ロードマップ](#ロードマップ)

---

## Why htmldriver?

GoでSSRアプリケーションやHTMLテンプレートのテストを書く際、既存のツールには以下の課題があります：

* **goquery**: DOMパースに特化しており、フォーム送信やHTTPリクエストの機能がない
* **net/http/httptest**: モックサーバーの構築は可能だが、フォーム操作のAPIが直感的でない
* **playwright-go / chromedp**: ヘッドレスブラウザは重く、起動が遅く、JavaScript実行が不要なケースでもオーバースペック

`htmldriver` は以下の特徴により、SSR/テンプレートテストに最適化されています：

* **軽量**: HTMLパースとHTTPリクエストのみ（ブラウザ不要）
* **直感的API**: `Fill()`, `Submit()`, `Click()` など、人間の操作に近い表現
* **Transport抽象化**: 任意のHTTPクライアント/フレームワークと統合可能
* **テストフレームワーク非依存**: `testing` パッケージへの依存なし

## 特徴

* **シンプルなAPI**: フォーム入力やリンククリックを直感的に操作可能
* **Transport抽象化**: 任意のHTTPクライアント/サーバーと連携可能
* **Cookie自動管理**: セッション維持や認証シナリオをサポート
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

    h "github.com/ppdx999/htmldriver"
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

    dom := h.New(MockTransport{}).Parse(html)

    form, err := dom.Form("#login-form")
    if err != nil {
        t.Fatal(err)
    }

    form.MustFill("username", "alice").MustFill("password", "secret")

    res, err := form.Submit()
    if err != nil {
        t.Fatal(err)
    }

    if res.Status != 200 {
        t.Fatalf("expected status 200, got %d", res.Status)
    }
}
```

## フォーム操作例

### エラーハンドリング版

```go
form, err := dom.Form("@login")
if err != nil {
    return err
}

form, err = form.Fill("username", "tester")
if err != nil {
    return err
}

form, err = form.Fill("password", "mypassword")
if err != nil {
    return err
}

form, err = form.CheckCheckbox("remember_me")
if err != nil {
    return err
}

res, err := form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

### メソッドチェーン版（Must系）

テスト時など、エラーが発生したら即座に失敗させたい場合は `Must` プレフィックス付きメソッドを使用できます。エラー時はパニックします。

```go
form, err := dom.Form("@login")
if err != nil {
    return err
}

form.MustFill("username", "tester").
    MustFill("password", "mypassword").
    MustCheckCheckbox("remember_me")

res, err := form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

## リンク例

```go
link, err := dom.Link("@profile")
if err != nil {
    return err
}

text, err := link.GetText()
if err != nil {
    return err
}
if text != "View Profile" {
    return fmt.Errorf("expected link text 'View Profile', got '%s'", text)
}

url, err := link.GetURL()
if err != nil {
    return err
}
if url.Path != "/users/123" {
    return fmt.Errorf("expected link URL '/users/123', got '%s'", url.Path)
}

res, err := link.Click()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```


## Table / List の確認例

### テーブル

```go
table, err := dom.Table("@user-list")
if err != nil {
    return err
}

rows, err := table.GetRows()
if err != nil {
    return err
}

// 行数の確認
if table.GetRowCount() != 3 {
    return fmt.Errorf("expected 3 rows, got %d", table.GetRowCount())
}

// セルの内容確認
if rows[0][0] != "Alice" {
    return fmt.Errorf("expected 'Alice', got '%s'", rows[0][0])
}

// 空テーブルの確認
if table.GetRowCount() == 0 {
    return fmt.Errorf("expected data, but table is empty")
}
```

### リスト

```go
list, err := dom.List("@todo-items")
if err != nil {
    return err
}

items, err := list.GetItems()
if err != nil {
    return err
}

// アイテム数の確認
if list.GetItemCount() != 5 {
    return fmt.Errorf("expected 5 items, got %d", list.GetItemCount())
}

// 内容の確認
if items[0] != "Buy groceries" {
    return fmt.Errorf("expected 'Buy groceries', got '%s'", items[0])
}
```

**注意**: 複雑な期待値比較や差分表示については、現在はロードマップで計画中です。

## 連続した操作

`Submit()` や `Click()` が返す `Response` には HTML 本文が含まれているため、再度 `Parse()` することで次の操作へと繋げられます。

### 方法1: CookieJar を共有する（推奨）

Cookie を複数のページ操作で共有したい場合は、`CookieJar` を明示的に作成して共有します。

```go
// Transport を作成
transport := NewMockTransport()

// Cookie を共有するための CookieJar を作成
jar := h.NewCookieJar()

// ログインページ
dom := h.NewWithCookieJar(transport, jar).Parse(loginPageHTML)

form, err := dom.Form("#login-form")
if err != nil {
    return err
}

form.MustFill("username", "alice").MustFill("password", "secret")

res, err := form.Submit()
if err != nil {
    return err
}

// ログイン後のページをパース（同じ jar を使用）
dashboardDOM := h.NewWithCookieJar(transport, jar).Parse(res.Body)

// プロフィール編集リンクをクリック
link, err := dashboardDOM.Link("@edit-profile")
if err != nil {
    return err
}

res, err = link.Click()
if err != nil {
    return err
}

// プロフィール編集ページをパース（同じ jar を使用）
editDOM := h.NewWithCookieJar(transport, jar).Parse(res.Body)

form, err = editDOM.Form("#profile-form")
if err != nil {
    return err
}

form.MustFill("bio", "Updated bio text")

res, err = form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

### 方法2: 同じ DOM インスタンスを使い続ける

Cookie を自動的に引き継ぎたい場合は、同じ DOM インスタンスで `Parse()` を繰り返すこともできます。

```go
transport := NewMockTransport()

// 最初のページ
dom := h.New(transport).Parse(loginPageHTML)

// ログイン
form, _ := dom.Form("#login-form")
form.MustFill("username", "alice").MustFill("password", "secret")
res, _ := form.Submit()

// 同じ DOM で次のページをパース（Cookie は自動的に引き継がれる）
dom = dom.Parse(res.Body)

// 以降も同様に dom を使い続ける
```

**重要**: `New()` で新しい DOM を作成すると Cookie は引き継がれません。`NewWithCookieJar()` で `CookieJar` を共有するか、同じ DOM インスタンスを使い続けてください。

## Cookie の扱い

`htmldriver`はResponseに含まれるCookieを自動的に保存し、以降のフォーム送信やリンククリック時に適切に送信します。これにより、セッション管理やユーザー認証のシナリオを簡単にテストできます。

事前にCookieを設定したい場合は以下のようにします。

```go
dom := h.New(transport).Parse(html)
dom.SetCookie("session_id", "abc123")

// 以降の操作で "session_id=abc123" が送信される
form, err := dom.Form("#protected-form")
if err != nil {
    return err
}

form.MustFill("data", "value")

res, err := form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

**注意**: Cookie の有効期限、パス、ドメインなどの詳細な属性については、現在はロードマップで計画中です。

## Selector について

`selector` には以下の2種類のロケータ文字列を指定できます。

| セレクタ | 説明 |
|----------|------|
| `@xxxxx` | `test-id` 属性が `xxxxx` の要素を特定します。 |
| `#xxxxx` | `id` 属性が `xxxxx` の要素を特定します。 |

より高度なCSSセレクタのサポートは将来的にロードマップで検討しています。


## Transport について

`Transport` は I/O を抽象化するインタフェースです。

`Form.Submit()`や`Link.Click()`は内部で`Transport.Do()`を呼び出し、HTTPリクエストを送信します。
そして、その結果を`Form.Submit()`や`Link.Click()`の呼び出し元に返します。

この仕組みにより`htmldriver`は特定のHTTPクライアントやサーバーフレームワークに依存せず、任意の実装と連携可能です。


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

### 実装例：標準 http.Client を使う場合

```go
type HTTPClientTransport struct {
    client  *http.Client
    baseURL string
}

func NewHTTPClientTransport(baseURL string) *HTTPClientTransport {
    return &HTTPClientTransport{
        client:  &http.Client{},
        baseURL: baseURL,
    }
}

func (t *HTTPClientTransport) Do(req h.Request) (h.Response, error) {
    // 相対URLを絶対URLに変換
    fullURL := t.baseURL + req.URL.String()

    var httpReq *http.Request
    var err error

    if req.Method == http.MethodPost {
        httpReq, err = http.NewRequest(req.Method, fullURL, strings.NewReader(req.Form.Encode()))
        if err != nil {
            return h.Response{}, err
        }
        httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    } else {
        httpReq, err = http.NewRequest(req.Method, fullURL, nil)
        if err != nil {
            return h.Response{}, err
        }
    }

    // ヘッダーをコピー
    for k, v := range req.Header {
        httpReq.Header[k] = v
    }

    resp, err := t.client.Do(httpReq)
    if err != nil {
        return h.Response{}, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return h.Response{}, err
    }

    return h.Response{
        Status: resp.StatusCode,
        Body:   string(body),
        Header: resp.Header,
        URL:    resp.Request.URL,
    }, nil
}
```

### Base URL の扱い

フォームの `action` や `<a>` の `href` が相対パスの場合、Transport 側で Base URL を管理します。

- **デフォルト**: `http://localhost` を使用
- **カスタマイズ**: Transport 実装時に `baseURL` を指定可能

## API 概要

```go
// ルート
func New(transport Transport) *DOM
func NewWithCookieJar(transport Transport, jar *CookieJar) *DOM
func NewCookieJar() *CookieJar
func (d *DOM) Parse(html string) *DOM
func (d *DOM) SetCookie(name, value string) *DOM
func (d *DOM) GetCookieJar() *CookieJar

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

// フォーム操作（エラーを返す版）
func (f *Form) Fill(name, value string) (*Form, error)           // テキスト入力
func (f *Form) CheckCheckbox(name string) (*Form, error)         // チェックボックスを選択
func (f *Form) UncheckCheckbox(name string) (*Form, error)       // チェックボックスの選択解除
func (f *Form) Select(name, value string) (*Form, error)         // セレクトボックス選択
func (f *Form) CheckRadio(name, value string) (*Form, error)     // ラジオボタン選択
func (f *Form) Submit() (Response, error)                        // 送信

// フォーム操作（メソッドチェーン版 - エラー時はパニック）
func (f *Form) MustFill(name, value string) *Form                // テキスト入力
func (f *Form) MustCheckCheckbox(name string) *Form              // チェックボックスを選択
func (f *Form) MustUncheckCheckbox(name string) *Form            // チェックボックスの選択解除
func (f *Form) MustSelect(name, value string) *Form              // セレクトボックス選択
func (f *Form) MustCheckRadio(name, value string) *Form          // ラジオボタン選択

// フォーム情報取得
func (f *Form) GetValue(name string) (string, error)             // 入力値取得
func (f *Form) HasField(name string) bool                        // フィールド存在確認


// リンク
func (l *Link) Click() (Response, error)
func (l *Link) GetURL() (*url.URL, error)
func (l *Link) GetText() (string, error)

// ボタン
func (b *Button) GetText() (string, error)

// テーブル
func (tbl *Table) GetRows() ([][]string, error)
func (tbl *Table) GetRowCount() int
func (tbl *Table) GetColCount() int

// リスト
func (lst *List) GetItems() ([]string, error)
func (lst *List) GetItemCount() int

// 画像
func (img *Img) GetSrc() (string, error)
func (img *Img) GetAlt() (string, error)
```

## フレームワーク統合

メジャーなHTTPサーバーフレームワーク向けのTransport実装を提供しています。

### Chi 統合

Chi フレームワークを使用している場合`ChiTransport`を利用して、`httptest.Server`と連携できます。

```go
import (
    "github.com/go-chi/chi/v5"
    h "github.com/ppdx999/htmldriver"
    "github.com/ppdx999/htmldriver/integrations/chitransport"
)

func Test_LoginFlow(t *testing.T) {
    r := chi.NewRouter()

    transport := chitransport.NewChiTransport(r)

    dom := h.New(transport).Parse(renderLoginPage())

    form, err := dom.Form("#login-form")
    if err != nil {
        t.Fatal(err)
    }

    form.MustFill("username", "alice").
        MustFill("password", "secret")

    res, err := form.Submit()
    if err != nil {
        t.Fatal(err)
    }

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
* [ ] リダイレクト追従（`3xx`）のサポート
* [ ] Cookie の詳細属性サポート（有効期限、パス、ドメイン、Secure、HttpOnly等）
* [ ] `Table`/`List` の差分レポート強化（どのセルが不一致かを詳細表示）
* [ ] リッチなセレクタ拡張（クラス、タグ、属性セレクタ、`:has()`, `:nth-of-type()` など）
* [ ] エラーメッセージの見やすい差分表示（カラー出力）
