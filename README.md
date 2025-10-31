# htmldriver

**純粋なHTMLに対して "人間の操作" を自動テストするための Go 向け補助ライブラリ**。

* 既存のHTML文字列（SSRの出力、テンプレート、ファイル、スナップショット）を読み込み
* フォームの入力/送信、リンク/ボタンのクリックをエミュレート
* 見出し/リスト/テーブル/任意テキストの存在確認
* **JavaScriptは実行しません**（純HTML + Web標準仕様に基づく動作に限定）

---

## 目次

* [Why htmldriver?](#why-htmldriver)
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

    form.Fill("username", "alice").Fill("password", "secret")

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

```go
form, err := dom.Form("@login")
if err != nil {
    return err
}

form.Fill("username", "tester").
    Fill("password", "mypassword").
    Check("remember_me")

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

form.Fill("data", "value")

res, err := form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

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
func (f *Form) Fill(name, value string) *Form           // テキスト入力
func (f *Form) Check(name string) *Form                 // チェックボックス/ラジオを選択
func (f *Form) Uncheck(name string) *Form               // チェックボックスの選択解除
func (f *Form) Select(name, value string) *Form         // セレクトボックス選択
func (f *Form) Choose(name, value string) *Form         // ラジオボタン選択
func (f *Form) Submit() (Response, error)               // 送信
func (f *Form) GetValue(name string) (string, error)    // 入力値取得
func (f *Form) HasField(name string) bool               // フィールド存在確認


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
    "github.com/ppdx999/htmldriver/integrations/chi_transport"
)

func Test_LoginFlow(t *testing.T) {
    r := chi.NewRouter()

    transport := chi_transport.NewChiTransport(r)

    dom := h.New(transport).Parse(renderLoginPage())

    form, err := dom.Form("#login-form")
    if err != nil {
        t.Fatal(err)
    }

    form.Fill("username", "alice").
        Fill("password", "secret")

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
* [ ] `Table`/`List` の差分レポート強化（どのセルが不一致かを詳細表示）
* [ ] リッチなセレクタ拡張（クラス、タグ、属性セレクタ、`:has()`, `:nth-of-type()` など）
* [ ] エラーメッセージの見やすい差分表示（カラー出力）
