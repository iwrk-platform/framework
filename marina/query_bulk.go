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

type BulkQueryInterface interface {
	Exec(ctx context.Context) error
	Query() (string, error)
	Index(name string) *BulkQuery
	Model(models ...any) *BulkQuery
}

type BulkQuery struct {
	queryType QueryType
	conn      *sqlx.DB
	index     string
	fields    []string
	query     []string
	err       error
}

// Exec method make bulk request and return error
func (bq *BulkQuery) Exec(ctx context.Context) error {
	if bq.err != nil {
		return bq.err
	}
	_, err := bq.conn.QueryContext(ctx, bq.buildBulkQuery())
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1064 && strings.Contains(mysqlErr.Message, "duplicate") {
			return err
		}
		return err
	}
	if bq.queryType == BulkUpsertQuery {
		_, err = bq.conn.QueryContext(ctx, bq.optimizeIndexQuery())
		if err != nil {
			return err
		}
	}
	return nil
}

// Query method return raw string query to index
func (bq *BulkQuery) Query() (string, error) {
	if bq.err != nil {
		return "", bq.err
	}
	return bq.buildBulkQuery(), nil
}

// Index set index table name
func (bq *BulkQuery) Index(name string) *BulkQuery {
	if len(name) == 0 {
		bq.err = errors.New("invalid index name")
	}
	bq.index = name
	return bq
}

// Model set request models
func (bq *BulkQuery) Model(models ...any) *BulkQuery {
	if len(models) == 0 {
		bq.err = errors.New("have not models to bulk insert")
	}
	fields, err := getFieldNames(models[0])
	if err != nil {
		bq.err = err
		return bq
	}
	for _, model := range models {
		bulkElement := make([]string, 0)
		for _, field := range fields {
			val, fieldType, errType := getFieldValueAndType(model, field)
			if err != nil {
				bq.err = errType
				return bq
			}
			fieldValue, errVal := buildFieldValue(val, fieldType)
			if err != nil {
				bq.err = errVal
				return bq
			}
			bulkElement = append(bulkElement, fieldValue)
		}
		bq.addBulkElement(bulkElement...)
	}
	for _, field := range fields {
		bq.fields = append(bq.fields, strcase.ToSnake(field))
	}

	return bq
}

func (bq *BulkQuery) addBulkElement(element ...string) {
	buffer := bytes.NewBufferString("(")
	buffer.WriteString(strings.Join(element, ", "))
	buffer.WriteString(")")
	bq.query = append(bq.query, buffer.String())
}

func (bq *BulkQuery) buildBulkQuery() string {
	sb := bytes.NewBufferString("")
	switch bq.queryType {
	case BulkInsertQuery:
		sb.WriteString("INSERT INTO ")
		sb.WriteString(bq.index)
		sb.WriteString("(")
		sb.WriteString(strings.Join(bq.fields, ", "))
		sb.WriteString(") VALUES")
	case BulkUpsertQuery:
		sb.WriteString("REPLACE INTO ")
		sb.WriteString(bq.index)
		sb.WriteString("(")
		sb.WriteString(strings.Join(bq.fields, ", "))
		sb.WriteString(") VALUES")
	default:
		return ""
	}
	sb.WriteString(strings.Join(bq.query, ", "))
	sb.WriteString(";")
	return sb.String()
}

func (bq *BulkQuery) optimizeIndexQuery() string {
	sb := bytes.NewBufferString("OPTIMIZE INDEX ")
	sb.WriteString(bq.index)
	sb.WriteString(";")
	return sb.String()
}
