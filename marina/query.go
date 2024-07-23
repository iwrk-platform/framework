package marina

import (
	"bytes"
	"context"
	"errors"
	"github.com/ettle/strcase"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"strings"
)

type QueryInterface interface {
	Index(name string) *Query
	Model(value any) *Query
	Where(field string, condition string, value any) *Query
}

type Query struct {
	queryType QueryType
	conn      *sqlx.DB
	index     string
	fields    []string
	query     []string
	where     [][]byte
	err       error
}

// Exec method make request and return error
func (q *Query) Exec(ctx context.Context) error {
	if q.err != nil {
		return q.err
	}
	_, err := q.conn.QueryContext(ctx, q.buildQuery(q.queryType))
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1064 && strings.Contains(mysqlErr.Message, "duplicate") {
			return err
		}
		return err
	}
	if q.queryType == UpsertQuery {
		_, err = q.conn.QueryContext(ctx, q.buildQuery(OptimizeQuery))
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *Query) Query() string {
	return q.buildQuery(q.queryType)
}

// Index set index table name
func (q *Query) Index(name string) *Query {
	if len(name) == 0 {
		q.err = errors.New("invalid index name")
	}
	q.index = name
	return q
}

// Model set request model
func (q *Query) Model(value any) *Query {
	fields, err := getFieldNames(value)
	if err != nil {
		q.err = err
		return q
	}
	for _, field := range fields {
		val, fieldType, err := getFieldValueAndType(value, field)
		if err != nil {
			q.err = err
			return q
		}
		fieldValue, err := buildFieldValue(val, fieldType)
		if err != nil {
			q.err = err
			return q
		}
		q.addFieldValue(fieldValue)
		q.fields = append(q.fields, strcase.ToSnake(field))
	}
	return q
}

func (q *Query) Where(query string, args ...any) *Query {
	qr, err := replacePlaceholders(query, args)
	if err != nil {
		q.err = err
	}
	sb := bytes.NewBufferString(qr)
	q.where = append(q.where, sb.Bytes())
	return q
}

func (q *Query) addFieldValue(value string) {
	q.query = append(q.query, value)
}

func (q *Query) buildQuery(qType QueryType) string {
	sb := bytes.NewBufferString("")
	switch qType {
	case OptimizeQuery:
		sb := bytes.NewBufferString("OPTIMIZE INDEX ")
		sb.WriteString(q.index)
		sb.WriteString(";")
		return sb.String()
	case DeleteQuery:
		sb.WriteString("DELETE FROM ")
		sb.WriteString(q.index)
		if len(q.where) != 0 {
			sb.WriteString(" WHERE ")
		}
		if len(q.where) != 0 {
			sb.Write(bytes.Join(q.where, []byte(" AND ")))
		}
		sb.WriteString(";")
		return sb.String()
	case InsertQuery:
		sb.WriteString("INSERT INTO ")
		sb.WriteString(q.index)
		sb.WriteString("(")
		sb.WriteString(strings.Join(q.fields, ", "))
		sb.WriteString(") VALUES (")
		sb.WriteString(strings.Join(q.query, ", "))
		sb.WriteString(");")
	case UpsertQuery:
		sb.WriteString("REPLACE INTO ")
		sb.WriteString(q.index)
		sb.WriteString("(")
		sb.WriteString(strings.Join(q.fields, ", "))
		sb.WriteString(") VALUES (")
		sb.WriteString(strings.Join(q.query, ", "))
		sb.WriteString(")")
		if len(q.where) != 0 {
			sb.WriteString(" WHERE ")
		}
		if len(q.where) != 0 {
			sb.Write(bytes.Join(q.where, []byte(" AND ")))
		}
		sb.WriteString(";")
		return sb.String()
	default:
		q.err = errors.New("invalid query type")
		return ""
	}
	return sb.String()
}
