CREATE TABLE expenses (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    amount_cents  BIGINT NOT NULL CHECK (amount_cents > 0),
    category      TEXT   NOT NULL CHECK (btrim(category) <> ''),
    description   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
