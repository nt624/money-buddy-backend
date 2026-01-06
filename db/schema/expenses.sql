CREATE TABLE expenses (
  id SERIAL PRIMARY KEY,
  amount INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  memo TEXT,
  spent_at DATE NOT NULL,
  status TEXT NOT NULL DEFAULT 'confirmed',
  created_at TIMESTAMP NOT NULL DEFAULT now()
);

ALTER TABLE expenses
ADD CONSTRAINT expenses_status_check
CHECK (status IN ('planned', 'confirmed'));
