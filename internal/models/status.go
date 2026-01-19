package models

import "strings"

// Status は支出の状態を表す列挙型です。
type Status string

const (
	StatusPlanned   Status = "planned"
	StatusConfirmed Status = "confirmed"
)

// IsValidStatus は有効なステータスかを判定します。
func IsValidStatus(s string) bool {
	switch strings.ToLower(s) {
	case string(StatusPlanned), string(StatusConfirmed):
		return true
	default:
		return false
	}
}

// NormalizeStatus はステータス文字列を正規化（小文字化）し、妥当性も判定します。
// 有効な場合は正規化済み文字列と true を返し、無効な場合は空文字と false を返します。
func NormalizeStatus(s string) (string, bool) {
	if s == "" {
		return "", false
	}
	lower := strings.ToLower(s)
	if IsValidStatus(lower) {
		return lower, true
	}
	return "", false
}
