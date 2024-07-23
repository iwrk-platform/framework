package marina

import (
	"bytes"
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"strconv"
	"strings"
	"unicode/utf8"
)

type SearchQueryInterface interface {
	Index(name string) *SearchQuery
	Where(field string, condition string, value any) *SearchQuery
	WhereOr(field string, condition string, value any) *SearchQuery
	Limit(limit int64) *SearchQuery
	Offset(offset int64) *SearchQuery
	Match(query string) *SearchQuery
	Order(query string) *SearchQuery
	Facet(fields ...string) *SearchQuery
	In(field string, value ...any) *SearchQuery
}

type SearchQuery struct {
	conn        *sqlx.DB
	count       bool
	meta        bool
	index       string
	limit       int64
	offset      int64
	groups      []*WhereGroup
	facetFields []string
	where       [][]byte
	whereOr     [][]byte
	order       [][]byte
	facet       [][]byte
	match       [][]byte
	err         error
}

// Scan method return founded entity ids and count
func (sq *SearchQuery) Scan(ctx context.Context) (*SearchResult, error) {
	if sq.err != nil {
		return nil, sq.err
	}
	searchResult := new(SearchResult)

	rows, err := sq.conn.QueryContext(ctx, sq.selectQuery())
	if err != nil {
		return nil, err
	}
	searchResult.Ids, err = getResult(rows)
	if err != nil {
		return nil, err
	}

	if !sq.meta {
		sq.count = true
		countResponse, errCount := sq.conn.QueryContext(ctx, sq.selectQuery())
		if errCount != nil {
			return nil, errCount
		}
		searchResult.Count, err = getCount(countResponse)
		if err != nil {
			return nil, err
		}
	} else {
		searchResult.Count, err = getCountFromMeta(rows)
		if err != nil {
			return nil, err
		}
	}

	return searchResult, nil
}

// Query method return raw string query to index
func (sq *SearchQuery) Query() (string, error) {
	if sq.err != nil {
		return "", sq.err
	}
	return sq.selectQuery(), nil
}

// Count method return founded entity count
func (sq *SearchQuery) Count(ctx context.Context) (*CountResult, error) {
	if sq.err != nil {
		return nil, sq.err
	}
	sq.count = true
	countResult := new(CountResult)
	countResponse, err := sq.conn.QueryContext(ctx, sq.selectQuery())
	if err != nil {
		return nil, err
	}
	countResult.Count, err = getCount(countResponse)
	if err != nil {
		return nil, err
	}
	return countResult, nil
}

// ScanWithFacet method return founded entity ids, count and facet
func (sq *SearchQuery) ScanWithFacet(ctx context.Context) (*SearchWithFacetResult, error) {
	if sq.err != nil {
		return nil, sq.err
	}
	var total int64
	var result []uint32
	var facets []*Facet

	rows, err := sq.conn.QueryContext(ctx, sq.selectQuery())
	if err != nil {
		return nil, err
	}
	result, err = getResult(rows)
	if err != nil {
		return nil, err
	}
	facets, err = getFacets(rows, sq.facetFields)
	if err != nil {
		return nil, err
	}

	if !sq.meta {
		sq.count = true
		count, errCount := sq.conn.QueryContext(ctx, sq.selectQuery())
		if errCount != nil {
			return nil, err
		}
		total, err = getCount(count)
		if err != nil {
			return nil, err
		}
	} else {
		total, err = getCountFromMeta(rows)
		if err != nil {
			return nil, err
		}
	}

	return &SearchWithFacetResult{
		Ids:    result,
		Total:  total,
		Facets: facets,
	}, nil
}

// Index set index table name
func (sq *SearchQuery) Index(name string) *SearchQuery {
	if len(name) == 0 {
		sq.err = errors.New("invalid index name")
	}
	sq.index = name
	return sq
}

func (sq *SearchQuery) WhereGroup(group *WhereGroup) {
	sq.groups = append(sq.groups, group)
}

func (sq *SearchQuery) Where(query string, args ...any) *SearchQuery {
	qr, err := replacePlaceholders(query, args)
	if err != nil {
		sq.err = err
	}
	sb := bytes.NewBufferString(qr)
	sq.where = append(sq.where, sb.Bytes())
	return sq
}

func (sq *SearchQuery) WhereOr(query string, args ...any) *SearchQuery {
	qr, err := replacePlaceholders(query, args)
	if err != nil {
		sq.err = err
	}
	sb := bytes.NewBufferString(qr)
	sq.whereOr = append(sq.whereOr, sb.Bytes())
	return sq
}

func (sq *SearchQuery) Limit(limit int64) *SearchQuery {
	sq.limit = limit
	return sq
}

func (sq *SearchQuery) Offset(offset int64) *SearchQuery {
	sq.offset = offset
	return sq
}

func (sq *SearchQuery) Match(query string) *SearchQuery {
	if query == "" {
		return sq
	}
	query = clearString(query)
	queryElements := strings.Split(query, " ")
	lengthQuery := len(queryElements)
	sb := bytes.NewBufferString("")
	switch lengthQuery {
	case 1:
		if utf8.RuneCountInString(query) < 3 {
			sb.WriteString("MATCH('^")
		} else {
			sb.WriteString("MATCH('")
		}
		sb.WriteString(query)
		sb.WriteString("*')")
	default:
		elemsAsterisk := make([][]byte, 0, lengthQuery)
		for _, element := range queryElements {
			qry := bytes.NewBufferString("")
			qry.WriteString("*")
			qry.WriteString(element)
			qry.WriteString("*")
			elemsAsterisk = append(elemsAsterisk, qry.Bytes())
		}
		cookedQuery := bytes.Join(elemsAsterisk, []byte(" "))
		sb.WriteString("MATCH('\"")
		sb.Write(cookedQuery)
		sb.WriteString("\"/1')")
	}
	sq.where = append(sq.where, sb.Bytes())
	sq.meta = true
	return sq
}

func (sq *SearchQuery) OrMatch(query string) *SearchQuery {
	if query == "" {
		return sq
	}
	query = clearString(query)
	queryElements := strings.Split(query, " ")
	lengthQuery := len(queryElements)
	sb := bytes.NewBufferString("")
	switch lengthQuery {
	case 1:
		if utf8.RuneCountInString(query) < 3 {
			sb.WriteString("MATCH('^")
		} else {
			sb.WriteString("MATCH('")
		}
		sb.WriteString(query)
		sb.WriteString("*')")
	default:
		elemsAsterisk := make([][]byte, 0, lengthQuery)
		for _, element := range queryElements {
			qry := bytes.NewBufferString("")
			qry.WriteString("*")
			qry.WriteString(element)
			qry.WriteString("*")
			elemsAsterisk = append(elemsAsterisk, qry.Bytes())
		}
		cookedQuery := bytes.Join(elemsAsterisk, []byte(" "))
		sb.WriteString("MATCH('\"")
		sb.Write(cookedQuery)
		sb.WriteString("\"/1')")
	}
	sq.whereOr = append(sq.whereOr, sb.Bytes())
	sq.meta = true
	return sq
}

func (sq *SearchQuery) Order(query string) *SearchQuery {
	if query != "" {
		sb := bytes.NewBufferString(query)
		sq.order = append(sq.order, sb.Bytes())
	}
	return sq
}

func (sq *SearchQuery) Facet(fields ...string) *SearchQuery {
	if len(fields) == 0 {
		return sq
	}
	for _, field := range fields {
		sf := bytes.NewBufferString(" FACET ")
		sf.WriteString(field)
		sq.facet = append(sq.facet, sf.Bytes())
		sq.facetFields = append(sq.facetFields, field)
	}
	return sq
}

func (sq *SearchQuery) In(field string, value any) *SearchQuery {
	sbp := bytes.NewBufferString(field)
	sbp.WriteString(" IN ")
	sbp.WriteString("(")
	vals := make([]string, 0)
	switch value.(type) {
	case []int:
		val := value.([]int)
		vals = append(vals, lo.Map(val, func(x int, index int) string {
			return strconv.FormatInt(int64(x), 10)
		})...)
	case []uint:
		val := value.([]uint)
		vals = append(vals, lo.Map(val, func(x uint, index int) string {
			return strconv.FormatUint(uint64(x), 10)
		})...)
	case []int32:
		val := value.([]int32)
		vals = append(vals, lo.Map(val, func(x int32, index int) string {
			return strconv.FormatInt(int64(x), 10)
		})...)
	case []uint32:
		val := value.([]uint32)
		vals = append(vals, lo.Map(val, func(x uint32, index int) string {
			return strconv.FormatUint(uint64(x), 10)
		})...)
	case []int64:
		val := value.([]int64)
		vals = append(vals, lo.Map(val, func(x int64, index int) string {
			return strconv.FormatInt(x, 10)
		})...)
	case []uint64:
		val := value.([]uint64)
		vals = append(vals, lo.Map(val, func(x uint64, index int) string {
			return strconv.FormatUint(x, 10)
		})...)
	case []string:
		val := value.([]string)
		vals = append(vals, lo.Map(val, func(x string, index int) string {
			return wrapString(x)
		})...)
	default:
		sq.err = errors.New("query have unsupported field: " + field)
	}
	sbp.WriteString(strings.Join(vals, ", "))
	sbp.WriteString(")")
	sq.where = append(sq.where, sbp.Bytes())
	return sq
}

func (sq *SearchQuery) OrIn(field string, value any) *SearchQuery {
	sbp := bytes.NewBufferString(field)
	sbp.WriteString(" IN ")
	sbp.WriteString("(")
	vals := make([]string, 0)
	switch value.(type) {
	case []int:
		val := value.([]int)
		vals = append(vals, lo.Map(val, func(x int, index int) string {
			return strconv.FormatInt(int64(x), 10)
		})...)
	case []uint:
		val := value.([]uint)
		vals = append(vals, lo.Map(val, func(x uint, index int) string {
			return strconv.FormatUint(uint64(x), 10)
		})...)
	case []int32:
		val := value.([]int32)
		vals = append(vals, lo.Map(val, func(x int32, index int) string {
			return strconv.FormatInt(int64(x), 10)
		})...)
	case []uint32:
		val := value.([]uint32)
		vals = append(vals, lo.Map(val, func(x uint32, index int) string {
			return strconv.FormatUint(uint64(x), 10)
		})...)
	case []int64:
		val := value.([]int64)
		vals = append(vals, lo.Map(val, func(x int64, index int) string {
			return strconv.FormatInt(x, 10)
		})...)
	case []uint64:
		val := value.([]uint64)
		vals = append(vals, lo.Map(val, func(x uint64, index int) string {
			return strconv.FormatUint(x, 10)
		})...)
	case []string:
		val := value.([]string)
		vals = append(vals, lo.Map(val, func(x string, index int) string {
			return wrapString(x)
		})...)
	default:
		sq.err = errors.New("query have unsupported field: " + field)
	}
	sbp.WriteString(strings.Join(vals, ", "))
	sbp.WriteString(")")
	sq.where = append(sq.whereOr, sbp.Bytes())
	return sq
}

func (sq *SearchQuery) selectQuery() string {
	if len(sq.index) == 0 {
		sq.err = errors.New("invalid index name")
	}
	sb := bytes.NewBufferString("SELECT ")
	if !sq.count {
		sb.WriteString("id")
	} else {
		sb.WriteString("COUNT(*)")
	}
	sb.WriteString(" FROM ")
	sb.WriteString(sq.index)
	if len(sq.where) != 0 || len(sq.whereOr) != 0 || len(sq.groups) != 0 {
		sb.WriteString(" WHERE ")
	}
	if len(sq.where) != 0 {
		sb.Write(bytes.Join(sq.where, []byte(" AND ")))
	}
	if len(sq.whereOr) != 0 {
		if len(sq.where) != 0 {
			sb.WriteString(" OR ")
		}
		sb.Write(bytes.Join(sq.whereOr, []byte(" OR ")))
	}
	for i, group := range sq.groups {
		if i != 0 || len(sq.where) != 0 || len(sq.whereOr) != 0 {
			sb.WriteString(group.sep)
		}
		sb.WriteString(" (")
		if len(group.where) != 0 {
			sb.Write(bytes.Join(sq.where, []byte(" AND ")))
		}
		if len(group.whereOr) != 0 {
			if len(group.where) != 0 {
				sb.WriteString(" OR ")
			}
			sb.Write(bytes.Join(sq.whereOr, []byte(" OR ")))
		}
		sb.WriteString(")")
	}
	if len(sq.order) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.Write(bytes.Join(sq.order, []byte(", ")))
		sb.WriteString(" ")
	}
	sb.Write(bytes.Join(sq.order, []byte{}))
	if sq.limit > 0 && !sq.count {
		sb.WriteString(" LIMIT ")
		sb.WriteString(strconv.FormatInt(sq.limit, 10))
	}
	if sq.offset > 0 && !sq.count {
		sb.WriteString(" OFFSET ")
		sb.WriteString(strconv.FormatInt(sq.offset, 10))
	}
	sb.Write(bytes.Join(sq.facet, []byte(" ")))
	sb.WriteString(";")
	if sq.meta {
		sb.WriteString("SHOW META;")
	}
	return sb.String()
}
