package dat

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func BenchmarkUpdateValuesSql(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Update("alpha").Set("something_id", 1).Where("id", 1).ToSQL()
	}
}

func BenchmarkUpdateValueMapSql(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Update("alpha").Set("something_id", 1).SetMap(map[string]interface{}{"b": 1, "c": 2}).Where("id", 1).ToSQL()
	}
}

func TestUpdateAllToSql(t *testing.T) {
	sql, args, err := Update("a").Set("b", 1).Set("c", 2).From("d").ToSQL()

	assert.NoError(t, err)
	assert.Equal(t, sql, quoteSQL(`UPDATE a SET %s = $1, %s = $2 FROM d`, "b", "c"))
	assert.Equal(t, args, []interface{}{1, 2})
}

func TestUpdateSingleToSql(t *testing.T) {
	sql, args, err := Update("a").Set("b", 1).Set("c", 2).From("d").Where("id = $1", 1).ToSQL()
	assert.NoError(t, err)
	assert.Equal(t, sql, quoteSQL(`UPDATE a SET %s = $1, %s = $2 FROM d WHERE (id = $3)`, "b", "c"))
	assert.Equal(t, args, []interface{}{1, 2, 1})
}

func TestUpdateSetMapToSql(t *testing.T) {
	sql, args, err := Update("a").SetMap(map[string]interface{}{"b": 1, "c": 2}).From("d").Where("id = $1", 1).ToSQL()
	assert.NoError(t, err)

	if sql == quoteSQL(`UPDATE a SET %s = $1, %s = $2 FROM d WHERE (id = $3)`, "b", "c") {
		assert.Equal(t, args, []interface{}{1, 2, 1})
	} else {
		assert.Equal(t, sql, quoteSQL(`UPDATE a SET %s = $1, %s = $2 FROM d WHERE (id = $3)`, "c", "b"))
		assert.Equal(t, args, []interface{}{2, 1, 1})
	}
}

func TestUpdateSetExprToSql(t *testing.T) {
	sql, args, err := Update("a").Set("foo", 1).Set("bar", Expr("COALESCE(bar, 0) + 1")).Where("id = $1", 9).ToSQL()

	assert.NoError(t, err)
	assert.Equal(t, sql, quoteSQL(`UPDATE a SET %s = $1, %s = COALESCE(bar, 0) + 1 WHERE (id = $2)`, "foo", "bar"))
	assert.Equal(t, args, []interface{}{1, 9})

	sql, args, err = Update("a").Set("foo", 1).Set("bar", Expr("COALESCE(bar, 0) + $1", 2)).Where("id = $1", 9).ToSQL()

	assert.NoError(t, err)
	assert.Equal(t, sql, quoteSQL(`UPDATE a SET %s = $1, %s = COALESCE(bar, 0) + $2 WHERE (id = $3)`, "foo", "bar"))
	assert.Equal(t, args, []interface{}{1, 2, 9})
}

func TestUpdateTenStaringFromTwentyToSql(t *testing.T) {
	sql, args, err := Update("a").Set("b", 1).Limit(10).Offset(20).ToSQL()

	assert.NoError(t, err)
	assert.Equal(t, sql, quoteSQL(`UPDATE a SET %s = $1 LIMIT 10 OFFSET 20`, "b"))
	assert.Equal(t, args, []interface{}{1})
}

func TestUpdateWhitelist(t *testing.T) {
	// type someRecord struct {
	// 	SomethingID int   `db:"something_id"`
	// 	UserID      int64 `db:"user_id"`
	// 	Other       bool  `db:"other"`
	// }
	sr := &someRecord{1, 2, false}
	sql, args, err := Update("a").
		SetWhitelist(sr, "user_id", "other").
		ToSQL()

	assert.NoError(t, err)
	assert.Equal(t, sql, quoteSQL(`UPDATE a SET %s = $1, %s = $2`, "user_id", "other"))
	checkSliceEqual(t, args, []interface{}{2, false})
}

func TestUpdateBlacklist(t *testing.T) {
	sr := &someRecord{1, 2, false}
	sql, args, err := Update("a").
		SetBlacklist(sr, "something_id").
		ToSQL()
	assert.NoError(t, err)

	assert.Equal(t, sql, quoteSQL(`UPDATE a SET %s = $1, %s = $2`, "user_id", "other"))
	checkSliceEqual(t, args, []interface{}{2, false})
}

func TestUpdateWhereExprSql(t *testing.T) {
	expr := Expr("id=$1", 100)
	sql, args, err := Update("a").Set("b", 10).Where(expr).ToSQL()
	assert.NoError(t, err)
	assert.Equal(t, sql, `UPDATE a SET b = $1 WHERE (id=$2)`)
	assert.Exactly(t, args, []interface{}{10, 100})
}

func TestUpdateBeyondMaxLookup(t *testing.T) {
	sqlBuilder := Update("a")
	setClauses := []string{}
	expectedArgs := []interface{}{}
	for i := 1; i < maxLookup+1; i++ {
		sqlBuilder = sqlBuilder.Set("b", i)
		setClauses = append(setClauses, fmt.Sprintf(" %s = $%d", quoteSQL("%s", "b"), i))
		expectedArgs = append(expectedArgs, i)
	}
	sql, args, err := sqlBuilder.Where("id = $1", maxLookup+1).ToSQL()
	assert.NoError(t, err)
	expectedSQL := fmt.Sprintf(`UPDATE a SET%s WHERE (id = $%d)`, strings.Join(setClauses, ","), maxLookup+1)
	expectedArgs = append(expectedArgs, maxLookup+1)

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)

}
