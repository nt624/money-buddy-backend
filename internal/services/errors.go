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

// NotFoundError はリソースが見つからないことを表します。
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	if e == nil {
		return "not found"
	}
	return e.Message
}

// InternalError は内部エラーを表します（外部に詳細を漏らさないためのラップ）。
type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	if e == nil {
		return "internal error"
	}
	return e.Message
}
