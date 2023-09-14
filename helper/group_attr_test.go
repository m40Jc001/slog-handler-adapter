package helper

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

var emptyAttrs = []slog.Attr{}
var (
	groupName000 = "group000"
)
var (
	int001 = slog.Int("int001", 1)
	int002 = slog.Int("int002", 2)
	int003 = slog.Int("int003", 3)

	attrSetInt001002003 = []slog.Attr{int001, int002, int003}
)

func TestWithGroup(t *testing.T) {
	t.Run("with group with empty name", func(t *testing.T) {
		origin := &AttrGroup{}
		withEmptyNameGroup := origin.WithGroup("")
		assert.NotSame(t, origin, withEmptyNameGroup)
		assert.Equal(t, origin, withEmptyNameGroup)
	})
	t.Run("with group with name", func(t *testing.T) {
		origin := &AttrGroup{}
		group000 := origin.WithGroup(groupName000)
		assert.NotSame(t, origin, group000)
		assert.Same(t, origin, group000.top)
		assert.EqualValues(t, emptyAttrs, group000.attrs)
		assert.EqualValues(t, groupName000, group000.name)
	})
}

func TestWithAttrs(t *testing.T) {
	t.Run("with attrs with empty attrs", func(t *testing.T) {
		origin := &AttrGroup{}
		withEmptyAttrs := origin.WithAttrs(emptyAttrs)
		assert.NotSame(t, origin, withEmptyAttrs)
		assert.Equal(t, origin, withEmptyAttrs)
	})

	t.Run("with attrs with empty attrSetInt001002003", func(t *testing.T) {
		origin := &AttrGroup{attrs: emptyAttrs}
		attrs001 := origin.WithAttrs(attrSetInt001002003)
		assert.NotSame(t, origin, attrs001)
		assert.Same(t, origin.top, attrs001.top)
		assert.EqualValues(t, attrs001.attrs, attrSetInt001002003)
		assert.EqualValues(t, origin.attrs, emptyAttrs)
	})
}

func TestAttrs(t *testing.T) {
	t.Run("withattrs and attrs", func(t *testing.T) {
		origin := (&AttrGroup{}).WithAttrs(attrSetInt001002003)
		assert.EqualValues(t, origin.Attrs(), attrSetInt001002003)
	})

	t.Run("withattrs group and attrs", func(t *testing.T) {
		origin := (&AttrGroup{}).WithGroup(groupName000).WithAttrs(attrSetInt001002003)
		assert.EqualValues(t, origin.Attrs(), []slog.Attr{slog.Group(groupName000, int001, int002, int003)})
	})
}
