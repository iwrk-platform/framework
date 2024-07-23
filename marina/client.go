package marina

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type MarinaClient interface {
	NewSearch() *SearchQuery
	NewInsert() *Query
	NewUpsert() *Query
	NewDelete() *Query
	NewCreate() *IndexQuery
	NewDrop() *IndexQuery
	NewBulkInsert() *BulkQuery
	NewBulkUpsert() *BulkQuery
}

type Facet struct {
	FilterName string
	Data       []*FacetItem
}

type FacetItem struct {
	Variant string
	Value   int64
}

type SearchResult struct {
	Ids   []uint32
	Count int64
}

type CountResult struct {
	Count int64
}

type SearchWithFacetResult struct {
	Ids    []uint32
	Total  int64
	Facets []*Facet
}

type marinaClient struct {
	Conn   *sqlx.DB
	Logger *zap.Logger
}

func (m *marinaClient) Close() error {
	return m.Conn.Close()
}

// NewSearch method initialize new search request
func (m *marinaClient) NewSearch() *SearchQuery {
	return &SearchQuery{
		conn:        m.Conn,
		where:       make([][]byte, 0),
		whereOr:     make([][]byte, 0),
		order:       make([][]byte, 0),
		facet:       make([][]byte, 0),
		match:       make([][]byte, 0),
		facetFields: make([]string, 0),
	}
}

// NewInsert method initialize new insert request
func (m *marinaClient) NewInsert() *Query {
	return &Query{
		queryType: InsertQuery,
		conn:      m.Conn,
		fields:    make([]string, 0),
		where:     make([][]byte, 0),
		query:     make([]string, 0),
	}
}

// NewUpsert method initialize new upsert request
func (m *marinaClient) NewUpsert() *Query {
	return &Query{
		queryType: UpsertQuery,
		conn:      m.Conn,
		fields:    make([]string, 0),
		where:     make([][]byte, 0),
		query:     make([]string, 0),
	}
}

// NewDelete method initialize new delete request
func (m *marinaClient) NewDelete() *Query {
	return &Query{
		queryType: DeleteQuery,
		conn:      m.Conn,
		fields:    make([]string, 0),
		where:     make([][]byte, 0),
		query:     make([]string, 0),
	}
}

// NewCreate method initialize new create index request
func (m *marinaClient) NewCreate() *IndexQuery {
	return &IndexQuery{
		conn:      m.Conn,
		queryType: CreateQuery,
		fields:    make([][]byte, 0),
	}
}

// NewDrop method initialize new drop index request
func (m *marinaClient) NewDrop() *IndexQuery {
	return &IndexQuery{
		conn:      m.Conn,
		queryType: DropQuery,
		fields:    make([][]byte, 0),
	}
}

// NewBulkInsert method initialize new bulk insert request
func (m *marinaClient) NewBulkInsert() *BulkQuery {
	return &BulkQuery{
		queryType: BulkInsertQuery,
		conn:      m.Conn,
		fields:    make([]string, 0),
		query:     make([]string, 0),
	}
}

// NewBulkUpsert method initialize new bulk upsert request
func (m *marinaClient) NewBulkUpsert() *BulkQuery {
	return &BulkQuery{
		queryType: BulkUpsertQuery,
		conn:      m.Conn,
		fields:    make([]string, 0),
		query:     make([]string, 0),
	}
}
