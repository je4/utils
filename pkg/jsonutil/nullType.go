package jsonutil

type JSONBytes []byte

func (n *JSONBytes) MarshalJSON() ([]byte, error) {
	return *n, nil
}

func (n *JSONBytes) UnmarshalJSON(data []byte) error {
	nt := JSONBytes(data)
	*n = nt
	return nil
}
