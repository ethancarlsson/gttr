package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStructAndPackageName(t *testing.T) {
	st, err := getStructAndPackageName(t.Context(), Arguments{
		Type: "Arguments",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, st.packageName)
	assert.NotNil(t, st.s)
}

func TestGenerateGetters(t *testing.T) {
	contents, err := generateGetters(t.Context(), Arguments{Type: "T"})

	assert.NoError(t, err)
	assert.NotNil(t, contents)
	assert.Equal(t, `package main

import (
	"go/ast"
	"time"
)

func (t *T) ArrType() ast.ArrayType {
	return t.arrType
}
func (t *T) Id() string {
	return t.id
}
func (t *T) Num() int {
	return t.num
}
func (t *T) CreatedAt() time.Time {
	return t.createdAt
}
`, contents.GoString())
}
