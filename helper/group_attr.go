package helper

import (
	"log/slog"
)

type AttrGroup struct {
	name  string
	attrs []slog.Attr
	top   *AttrGroup
}

func (g *AttrGroup) clone() *AttrGroup {
	return &AttrGroup{
		name:  g.name,
		attrs: g.attrs,
		top:   g.top,
	}
}

func (g *AttrGroup) WithGroup(name string) *AttrGroup {
	if name == "" {
		return g.clone()
	}

	return &AttrGroup{
		name:  name,
		attrs: []slog.Attr{},
		top:   g,
	}
}

func (g *AttrGroup) WithAttrs(attrs []slog.Attr) *AttrGroup {
	if len(attrs) == 0 {
		return g.clone()
	}

	return &AttrGroup{
		name:  g.name,
		attrs: append(g.attrs, attrs...),
		top:   g.top,
	}
}

func (g *AttrGroup) Attrs() []slog.Attr {
	rt := []slog.Attr{}
	for head := g; head != nil; head = head.top {
		if head.name != "" {
			attrs := make([]any, 0, len(head.attrs)+len(rt))

			for i, attr := range head.attrs {
				if attr.Key == "" {
					continue
				}
				attrs = append(attrs, head.attrs[i])
			}

			for i, attr := range rt {
				if attr.Key == "" {
					continue
				}
				attrs = append(attrs, rt[i])
			}

			group := slog.Group(head.name, attrs...)
			rt = []slog.Attr{group}
		} else {
			rt = append(rt, head.attrs...)
		}
	}
	return rt
}
