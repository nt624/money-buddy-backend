package models

type CreateExpenseInput struct {
	Amount     *int   `json:"amount" binding:"required"`
	CategoryID *int   `json:"category_id" binding:"required"`
	Memo       string `json:"memo"`
	SpentAt    string `json:"spent_at" binding:"required"`
	Status     string `json:"status"`
}

type Expense struct {
	ID       int      `json:"id"`
	Amount   int      `json:"amount"`
	Memo     string   `json:"memo"`
	SpentAt  string   `json:"spent_at"`
	Status   string   `json:"status"`
	Category Category `json:"category"`
}
