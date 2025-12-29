CREATE TABLE expenses (
  id SERIAL PRIMARY KEY,
  amount INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  memo TEXT,
  spent_at DATE NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);
