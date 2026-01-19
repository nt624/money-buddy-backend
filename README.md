# money-buddy-backend

このリポジトリは Money Buddy のバックエンドです。Go と `sqlc` を使って Postgres とやり取りする小さな API を提供します。

**重要:** このプロジェクトは sqlc によって生成されたコード（`/db/generated`）をリポジトリに含めています。理由と運用は下記を参照してください。

---

## 方針: sqlc 生成物をリポジトリに含める理由
- CI やローカルで `sqlc` をインストールしていない環境でもビルド可能にするため。
- 生成コードをコミットしておくとレビューやデバッグが容易になるため。

運用上の注意:
- `sqlc` のバージョンを固定してください（`sqlc.yaml` にバージョン情報を残すか、README に記載）。
- 生成 SQL や DB スキーマを変更したときは必ず `sqlc generate` を実行して `db/generated` を更新し、差分をコミットしてください。
- CI では生成物が最新かをチェックするステップを追加することを推奨します（例: `sqlc generate` の後 `git diff --exit-code`）。

---

## 開発環境セットアップ（ローカル）

前提:
- Go (推奨: 1.20+)
- Postgres（ローカルまたはリモート）

1. このリポジトリをクローン

```bash
git clone <repo-url>
cd money-buddy-backend
```

2. データベース接続設定
- デフォルトの DSN は `internal/db/db.go` に書かれています。必要に応じて編集してください。

3. 既に生成済みの sqlc コードはリポジトリに含めてあるため、通常は追加の生成は不要です。
	 ただし、SQL を変更した場合は以下を実行して生成物を更新してください。

```bash
# 必要なら sqlc をインストール
go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.30.0

cd db
sqlc generate
```

4. ビルド・起動

```bash
go build ./...
go run cmd/server/main.go
```

5. API の例（curl）

リクエスト例（`spent_at` は `YYYY-MM-DD` または RFC3339 を受け付けます）:

```bash
curl -X POST http://localhost:8080/expenses \
	-H "Content-Type: application/json" \
	-d '{
		"amount": 1500,
		"category_id": 2,
		"memo": "食費",
		"spent_at": "2025-01-03"
	}'
```

成功時のレスポンスは `{"expense": {...}}` です。

---

## 更新 API 例（PUT /expenses/:id）

- 経路: `PUT /expenses/:id`
- 仕様:
	- `status` は `planned` または `confirmed` のみ有効
	- 遷移ルール: `confirmed` → `planned` は禁止、`planned` → `confirmed` は許可
	- `spent_at` は `YYYY-MM-DD` または RFC3339 を受け付けます

リクエスト例:

```bash
curl -X PUT http://localhost:8080/expenses/42 \
	-H "Content-Type: application/json" \
	-d '{
		"amount": 700,
		"category_id": 5,
		"memo": "updated",
		"spent_at": "2025-07-01",
		"status": "confirmed"
	}'
```

成功レスポンス（200）例:

```json
{
	"expense": {
		"id": 42,
		"amount": 700,
		"memo": "updated",
		"spent_at": "2025-07-01",
		"status": "confirmed",
		"category": { "id": 5 }
	}
}
```

エラーレスポンス例:
- バリデーションエラー（400）: `{ "error": "amount must be greater than 0" }`
- ステータス遷移エラー（409）: `{ "error": "invalid status transition" }`
- 内部エラー（500）: `{ "error": "internal server error" }`

---

## 一覧 API 例（GET /expenses）

- 経路: `GET /expenses`
- レスポンスは `expenses` 配列を含むオブジェクト

リクエスト例:

```bash
curl -X GET http://localhost:8080/expenses
```

成功レスポンス（200）例:

```json
{
	"expenses": [
		{
			"id": 1,
			"amount": 1500,
			"memo": "食費",
			"spent_at": "2025-01-03T00:00:00Z",
			"status": "confirmed",
			"category": { "id": 2, "name": "food" }
		}
	]
}
```

---

## 作成 API 例（POST /expenses）

- 経路: `POST /expenses`
- `status` は省略可能（省略時は `confirmed` が適用）。有効値は `planned`/`confirmed`

リクエスト例:

```bash
curl -X POST http://localhost:8080/expenses \
	-H "Content-Type: application/json" \
	-d '{
		"amount": 1500,
		"category_id": 2,
		"memo": "食費",
		"spent_at": "2025-01-03",
		"status": "planned"
	}'
```

成功レスポンス（201）例:

```json
{
	"expense": {
		"id": 1,
		"amount": 1500,
		"memo": "食費",
		"spent_at": "2025-01-03",
		"status": "planned",
		"category": { "id": 2 }
	}
}
```

エラーレスポンス例:
- バリデーションエラー（400）: `{ "error": "amount must be greater than 0" }`
- 内部エラー（500）: `{ "error": "internal server error" }`

---

## 削除 API 例（DELETE /expenses/:id）

- 経路: `DELETE /expenses/:id`
- 成功時はボディなしで `204 No Content`

リクエスト例:

```bash
curl -X DELETE http://localhost:8080/expenses/123
```

レスポンス例:
- 成功（204）: ボディなし
- ID不正（400）: `{ "error": "invalid expense ID" }`
- バリデーションエラー（400）: `{ "error": "cannot delete planned expense" }`
- 内部エラー（500）: `{ "error": "internal server error" }`

---

## CI の推奨ステップ（例: GitHub Actions）

ワークフロー内に必ず `sqlc generate`（または生成済みの検証）を含めてください。例:

```yaml
- name: Install sqlc
	run: go install github.com/kyleconroy/sqlc/cmd/sqlc@v1.30.0

- name: Generate sqlc code
	working-directory: ./db
	run: sqlc generate

- name: Check generated code is up to date
	run: git diff --exit-code db/generated || (echo "Generated files out of date" && exit 1)

- name: Build
	run: go build ./...
```

---

## コード構成（重要なファイル）
- `cmd/server/main.go` - サーバー起点。sqlc の `Queries` を生成してリポジトリに渡します。
- `internal/db/db.go` - DB 接続（DSN 設定）
- `internal/handlers` - Gin ハンドラ
- `internal/services` - ビジネスロジック
- `internal/repositories` - リポジトリ層。インターフェースと実装（メモリ、sqlc）に分割しています。
- `db/` - SQL & `sqlc` 設定
	- `db/generated/` - sqlc による生成物（このリポジトリに含める）

---

## 運用ガイドライン
- SQL やスキーマ変更時: `db/expenses.sql` や `schema/` を更新し、`sqlc generate` を実行、`db/generated` をコミット。
- 生成物の差分は PR で確認すること。大きな差分が出た場合は sqlc のバージョン差の可能性を疑う。
- 生成物を更新する際は、他の開発者に通知するか PR に生成手順を含めてください。

---

必要ならこの README に CI 用のフルワークフローや Docker 起動手順も追記します。どの程度まで書きたいか教えてください。
