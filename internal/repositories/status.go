package repositories

import "money-buddy-backend/internal/models"

// defaultStatus はリポジトリ層の防御的なフォールバックを担います。
// サービス層が `status` を明示的に設定しない（空文字のまま渡す）ケースに備え、
// モデルの正規化ロジックを用いて有効値へ正規化し、無効または空の場合は既定値
// （`confirmed`）を返します。ビジネスルールの検証はサービス層が担い、ここでは
// ストレージ整合性の確保のみを目的とします。
func defaultStatus(s string) string {
	if normalized, ok := models.NormalizeStatus(s); ok {
		return normalized
	}
	// 入力が空や無効な場合は既定で confirmed を返す
	return string(models.StatusConfirmed)
}
