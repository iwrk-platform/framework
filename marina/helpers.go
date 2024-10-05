package marina

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type QueryType string

const (
	BulkInsertQuery QueryType = "bulk_insert"
	BulkUpsertQuery QueryType = "bulk_upsert"
	InsertQuery     QueryType = "insert"
	UpsertQuery     QueryType = "upsert"
	CreateQuery     QueryType = "create"
	DropQuery       QueryType = "drop"
	DeleteQuery     QueryType = "delete"
	OptimizeQuery   QueryType = "optimize"
)

// parsers

func getCount(rows *sql.Rows) (int64, error) {
	var total int64
	var err error
	for rows.Next() {
		total, err = getIntValueFromColumn(rows, "count(*)")
		if err != nil {
			return 0, err
		}
	}
	return total, nil
}

func getResult(rows *sql.Rows) ([]uint32, error) {
	var result []uint32
	for rows.Next() {
		val, err := getIntValueFromColumn(rows, "id")
		if err != nil {
			return nil, err
		}
		result = append(result, uint32(val))
		continue
	}
	return result, nil
}

func getCountFromMeta(rows *sql.Rows) (int64, error) {
	for rows.NextResultSet() {
		var total int64
		for rows.Next() {
			value, err := getStringValueFromColumn(rows, "Variable_name")
			if err != nil {
				return 0, err
			}
			if value == "total_found" {
				total, err = getIntValueFromColumn(rows, "Value")
				if err != nil {
					return 0, err
				}
				return total, nil
			}
		}
	}
	return 0, nil
}

func getFacets(rows *sql.Rows, filterFields []string) ([]*Facet, error) {
	var facets []*Facet
	for rows.NextResultSet() {
		for rows.Next() {
			for _, fieldName := range filterFields {
				var (
					facet   Facet
					err     error
					variant string
					value   int64
				)
				facet.FilterName = fieldName
				variant, err = getStringValueFromColumn(rows, fieldName)
				if err != nil {
					return nil, err
				}
				value, err = getIntValueFromColumn(rows, "count(*)")
				if err != nil {
					return nil, err
				}
				facet.Data = append(facet.Data, &FacetItem{
					Variant: variant,
					Value:   value,
				})
				facets = append(facets, &facet)
			}
		}
	}
	return facets, nil
}

func getIntValueFromColumn(rows *sql.Rows, name string) (int64, error) {
	columns := make(map[string]interface{})
	err := sqlx.MapScan(rows, columns)
	if err != nil {
		return 0, err
	}
	value, ok := columns[name].([]byte)
	if !ok {
		return 0, errors.New(fmt.Sprintf("error get int value from %s", name))
	}
	result, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("error convert int value from %s", name))
	}
	return result, nil
}

func getStringValueFromColumn(rows *sql.Rows, name string) (string, error) {
	columns := make(map[string]interface{})
	err := sqlx.MapScan(rows, columns)
	if err != nil {
		return "", err
	}
	value, ok := columns[name].([]byte)
	if !ok {
		return "", errors.New(fmt.Sprintf("error get string value from %s", name))
	}
	return string(value), nil
}

// reflections

func getFieldNames(object any) ([]string, error) {
	if t := reflect.Indirect(reflect.ValueOf(object)); t.IsValid() {
		if t.Kind() == reflect.Struct {
			fields := make([]string, 0, t.NumField())
			for i := 0; i < t.NumField(); i++ {
				f := t.Type().Field(i)
				if f.IsExported() {
					fields = append(fields, f.Name)
				}
			}
			return fields, nil
		}
		return []string{}, errors.New("object is not a struct")
	}
	return []string{}, errors.New("object not valid")
}

func getFieldValueAndType(object any, name string) (any, string, error) {
	if t := reflect.Indirect(reflect.ValueOf(object)); t.IsValid() {
		if field := t.FieldByName(name); field.IsValid() {
			return field.Interface(), field.Type().String(), nil
		}
	}
	return nil, "", fmt.Errorf("field %s not valid", name)
}

func getFieldType(object any, name string) (string, error) {
	if t := reflect.Indirect(reflect.ValueOf(object)); t.IsValid() {
		if field := t.FieldByName(name); field.IsValid() {
			return field.Type().String(), nil
		}
	}
	return "", fmt.Errorf("field %s not valid", name)
}

// builders

func wrapString(value string) string {
	val := clearString(value)
	sb := bytes.NewBufferString("'")
	if len(val) == 0 {
		sb.WriteString(" '")
		return sb.String()
	}
	sb.WriteString(val)
	sb.WriteString("'")
	return sb.String()
}

func clearString(item string) string {
	var (
		reBreak    string
		reSymbol   []byte
		reSpace    []byte
		reBreakers = regexp.MustCompile(`<.*>`)
		reSymbols  = regexp.MustCompile("['\"<>/()\\[\\];:*{}!?=\\-+_~`$^&#№%\\\\]")
		reSpaces   = regexp.MustCompile(`\s+`)
	)
	reBreak = reBreakers.ReplaceAllString(item, "")
	reSymbol = reSymbols.ReplaceAll([]byte(reBreak), []byte(""))
	reSpace = reSpaces.ReplaceAll(reSymbol, []byte(" "))

	return strings.TrimSpace(string(reSpace))
}

func buildFieldValue(value any, fieldType string) (string, error) {
	switch fieldType {
	case "string":
		return wrapString(value.(string)), nil
	case "[]string":
		val := value.([]string)
		return wrapString(strings.Join(val, " ")), nil
	case "bool":
		if value.(bool) == true {
			return "true", nil
		}
		return "false", nil
	case "int":
		val := value.(int)
		return strconv.FormatInt(int64(val), 10), nil
	case "uint":
		val := value.(uint)
		return strconv.FormatUint(uint64(val), 10), nil
	case "int32":
		val := value.(int32)
		return strconv.FormatInt(int64(val), 10), nil
	case "uint32":
		val := value.(uint32)
		return strconv.FormatUint(uint64(val), 10), nil
	case "int64":
		return strconv.FormatInt(value.(int64), 10), nil
	case "uint64":
		return strconv.FormatUint(value.(uint64), 10), nil
	case "float32":
		val := value.(float32)
		return strconv.FormatFloat(float64(val), 'f', -1, 64), nil
	case "float64":
		return strconv.FormatFloat(value.(float64), 'f', -1, 64), nil
	case "[]int":
		val := value.([]int)
		return buildArrayElement(lo.Map(val, func(x int, index int) string {
			return strconv.FormatInt(int64(x), 10)
		})...), nil
	case "[]uint":
		val := value.([]uint)
		return buildArrayElement(lo.Map(val, func(x uint, index int) string {
			return strconv.FormatUint(uint64(x), 10)
		})...), nil
	case "[]int32":
		val := value.([]int32)
		return buildArrayElement(lo.Map(val, func(x int32, index int) string {
			return strconv.FormatInt(int64(x), 10)
		})...), nil
	case "[]uint32":
		val := value.([]uint32)
		return buildArrayElement(lo.Map(val, func(x uint32, index int) string {
			return strconv.FormatUint(uint64(x), 10)
		})...), nil
	case "[]int64":
		val := value.([]int64)
		return buildArrayElement(lo.Map(val, func(x int64, index int) string {
			return strconv.FormatInt(x, 10)
		})...), nil
	case "[]uint64":
		val := value.([]uint64)
		return buildArrayElement(lo.Map(val, func(x uint64, index int) string {
			return strconv.FormatUint(x, 10)
		})...), nil
	case "[]float32":
		val := value.([]float32)
		return buildArrayElement(lo.Map(val, func(x float32, index int) string {
			return strconv.FormatFloat(float64(x), 'f', -1, 64)
		})...), nil
	case "[]float64":
		val := value.([]float64)
		return buildArrayElement(lo.Map(val, func(x float64, index int) string {
			return strconv.FormatFloat(x, 'f', -1, 64)
		})...), nil
	default:
		return "", errors.New("query have unsupported field type: " + fieldType)
	}
}

func buildArrayElement(elements ...string) string {
	sb := bytes.NewBufferString("(")
	if len(elements) == 0 {
		sb.WriteString("0)")
		return sb.String()
	}
	sb.WriteString(strings.Join(elements, ","))
	sb.WriteString(")")
	return sb.String()
}

func replacePlaceholders(s string, arg []any) (string, error) {
	n := strings.Count(s, "?")
	if n != len(arg) {
		fmt.Println("wrong number of arguments")
		return "", errors.New("wrong number of arguments")
	}
	for i := 0; i < n; i++ {
		m := strings.Index(s, "?")
		if m < 0 {
			break
		}
		argument, err := bufArg(arg[i])
		if err != nil {
			return "", err
		}
		v := s[:m-1] + " " + argument + s[m+1:]
		s = v
	}
	return s, nil
}

func bufArg(arg any) (string, error) {
	switch arg.(type) {
	case int:
		val := arg.(int)
		return strconv.FormatInt(int64(val), 10), nil
	case uint:
		val := arg.(uint)
		return strconv.FormatUint(uint64(val), 10), nil
	case int32:
		val := arg.(int32)
		return strconv.FormatInt(int64(val), 10), nil
	case uint32:
		val := arg.(uint32)
		return strconv.FormatUint(uint64(val), 10), nil
	case string:
		return wrapString(arg.(string)), nil
	case int64:
		return strconv.FormatInt(arg.(int64), 10), nil
	case uint64:
		return strconv.FormatUint(arg.(uint64), 10), nil
	case bool:
		if arg.(bool) {
			return "true", nil
		} else {
			return "false", nil
		}
	default:
		fmt.Println("query have unsupported field")
		return "", errors.New("query have unsupported field")
	}
}
