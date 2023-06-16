package output

func (d *dual) Bytes() ([]byte, error) {
	return d.buf.Bytes(), nil
}
