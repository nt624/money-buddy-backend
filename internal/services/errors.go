package services

// ValidationError はサービス層が返す入力バリデーションエラーを表します。
// 具体的な型にすることで、呼び出し側は errors.As などでエラーの種類を判別できます。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return "validation error"
	}
	return e.Message
}
