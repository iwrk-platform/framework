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

type IndexQueryInterface interface {
	Exec(ctx context.Context) error
	Index(name string) *IndexQuery
	Model(value any) *IndexQuery
}

type IndexQuery struct {
	queryType QueryType
	conn      *sqlx.DB
	index     string
	fields    [][]byte
	err       error
}

// Exec method make request and return error
func (tq *IndexQuery) Exec(ctx context.Context) error {
	if tq.err != nil {
		return tq.err
	}
	_, err := tq.conn.QueryContext(ctx, tq.buildTableQuery())
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1064 && strings.Contains(mysqlErr.Message, "duplicate") {
			return err
		}
		return err
	}
	return nil
}

// Query method return raw string query to index
func (tq *IndexQuery) Query() string {
	return tq.buildTableQuery()
}

// Index set request index name
func (tq *IndexQuery) Index(name string) *IndexQuery {
	if len(name) == 0 {
		tq.err = errors.New("invalid index name")
	}
	tq.index = name
	return tq
}

// Model set request model
func (tq *IndexQuery) Model(value any) *IndexQuery {
	fields, err := getFieldNames(value)
	if err != nil {
		tq.err = err
		return tq
	}
	for _, field := range fields {
		fieldType, err := getFieldType(value, field)
		if err != nil {
			tq.err = err
			break
		}
		tq.addField(field, fieldType)
		if tq.err != nil {
			break
		}
	}
	return tq
}

func (tq *IndexQuery) addField(field string, fieldType string) {
	sb := bytes.NewBufferString(strcase.ToSnake(field))
	switch fieldType {
	case "string", "[]string":
		sb.WriteString(" text indexed")
	case "bool":
		sb.WriteString(" bool")
	case "int", "uint", "uint32", "int32":
		sb.WriteString(" int")
	case "uint64", "int64":
		sb.WriteString(" bigint")
	case "float32", "float64":
		sb.WriteString(" float")
	case "[]int", "[]uint", "[]uint32", "[]int32", "[]float32":
		sb.WriteString(" multi")
	case "[]uint64", "[]int64", "[]float64":
		sb.WriteString(" multi64")
	default:
		tq.err = errors.New("index query have unsupported field type: " + fieldType)
		return
	}
	tq.fields = append(tq.fields, sb.Bytes())
}

func (tq *IndexQuery) buildTableQuery() string {
	sb := bytes.NewBufferString("")
	switch tq.queryType {
	case CreateQuery:
		sb.WriteString("CREATE TABLE ")
		sb.WriteString(tq.index)
		sb.WriteString(" (")
		sb.Write(bytes.Join(tq.fields, []byte(", ")))
		sb.WriteString(") dict='keywords' index_exact_words='1' min_infix_len='2'")
	case DropQuery:
		sb.WriteString("DROP TABLE IF EXISTS ")
		sb.WriteString(tq.index)
	default:
		tq.err = errors.New("invalid query type")
		return ""
	}
	return sb.String()
}
