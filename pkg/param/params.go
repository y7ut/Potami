package param

type Value any

type Text struct {
	Text  string
	Valid bool
}

func (t *Text) Scan(value any) error {
	if value == nil {
		t.Text, t.Valid = "", false
		return nil
	}
	t.Valid = true
	return assign(t.Text, value)
}

func (t *Text) Value() (Value, error) {
	if !t.Valid {
		return nil, nil
	}
	return t.Text, nil
}

type Number struct {
	Number float64
	Valid  bool
}

func (n *Number) Scan(value any) error {
	if value == nil {
		n.Number, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	return assign(n.Number, value)
}

func (n *Number) Value() (Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Number, nil
}

type Documents struct {
	Documents []Documents
	Content   string
	Valid     bool
}


