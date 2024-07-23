package marina

import "bytes"

type WhereGroup struct {
	sep     string
	where   [][]byte
	whereOr [][]byte
	err     error
}

func NewWhereGroup(sep string) *WhereGroup {
	return &WhereGroup{
		sep:     sep,
		where:   make([][]byte, 0),
		whereOr: make([][]byte, 0),
	}
}

func (wg *WhereGroup) Where(query string, args ...any) *WhereGroup {
	qr, err := replacePlaceholders(query, args)
	if err != nil {
		wg.err = err
	}
	sb := bytes.NewBufferString(qr)
	wg.where = append(wg.where, sb.Bytes())
	return wg
}

func (wg *WhereGroup) WhereOr(query string, args ...any) *WhereGroup {
	qr, err := replacePlaceholders(query, args)
	if err != nil {
		wg.err = err
	}
	sb := bytes.NewBufferString(qr)
	wg.whereOr = append(wg.whereOr, sb.Bytes())
	return wg
}
