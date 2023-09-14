package logrus

import (
	"bytes"
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

/*
	This unit test is adapted from go 1.21.1 log/slog/handler_test.go
	with slight modifications to test the logrus handler
*/

func TestDefaultHandle(t *testing.T) {
	ctx := context.Background()
	preAttrs := []slog.Attr{slog.Int("pre", 0)}
	attrs := []slog.Attr{slog.Int("a", 1), slog.String("b", "two")}
	for _, test := range []struct {
		name  string
		with  func(h slog.Handler) slog.Handler
		attrs []slog.Attr
		want  string
	}{
		{
			name: "no attrs",
			want: "level=info msg=message\n",
		},
		{
			name:  "attrs",
			attrs: attrs,
			want:  "level=info msg=message a=1 b=two\n",
		},
		{
			name:  "preformatted",
			with:  func(h slog.Handler) slog.Handler { return h.WithAttrs(preAttrs) },
			attrs: attrs,
			want:  "level=info msg=message a=1 b=two pre=0\n",
		},
		{
			name: "groups",
			attrs: []slog.Attr{
				slog.Int("a", 1),
				slog.Group("g",
					slog.Int("b", 2),
					slog.Group("h", slog.Int("c", 3)),
					slog.Int("d", 4)),
				slog.Int("e", 5),
			},
			want: "level=info msg=message a=1 e=5 g.b=2 g.d=4 g.h.c=3\n",
		},
		{
			name:  "group",
			with:  func(h slog.Handler) slog.Handler { return h.WithAttrs(preAttrs).WithGroup("s") },
			attrs: attrs,
			want:  "level=info msg=message pre=0 s.a=1 s.b=two\n",
		},
		{
			name: "preformatted groups",
			with: func(h slog.Handler) slog.Handler {
				return h.WithAttrs([]slog.Attr{slog.Int("p1", 1)}).
					WithGroup("s1").
					WithAttrs([]slog.Attr{slog.Int("p2", 2)}).
					WithGroup("s2")
			},
			attrs: attrs,
			want:  "level=info msg=message p1=1 s1.p2=2 s1.s2.a=1 s1.s2.b=two\n",
		},
		{
			name: "two with-groups",
			with: func(h slog.Handler) slog.Handler {
				return h.WithAttrs([]slog.Attr{slog.Int("p1", 1)}).
					WithGroup("s1").
					WithGroup("s2")
			},
			attrs: attrs,
			want:  "level=info msg=message p1=1 s1.s2.a=1 s1.s2.b=two\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var got string

			buf := &bytes.Buffer{}

			var h slog.Handler = NewHandler(buf, &HandlerOptions{})

			if test.with != nil {
				h = test.with(h)
			}

			r := slog.NewRecord(time.Time{}, slog.LevelInfo, "message", 0)
			r.AddAttrs(test.attrs...)
			if err := h.Handle(ctx, r); err != nil {
				t.Fatal(err)
			}

			got = buf.String()
			if got != test.want {
				t.Errorf("\ngot  %s\nwant %s", got, test.want)
			}
		})
	}
}

func TestConcurrentWrites(t *testing.T) {
	ctx := context.Background()
	count := 1000
	for _, handlerType := range []string{"text", "json"} {
		t.Run(handlerType, func(t *testing.T) {
			var buf bytes.Buffer
			var h slog.Handler
			switch handlerType {
			case "text":
				h = NewHandler(&buf, &HandlerOptions{})
			case "json":
				h = NewHandler(&buf, &HandlerOptions{JSONFormatter: true})
			default:
				t.Fatalf("unexpected handlerType %q", handlerType)
			}
			sub1 := h.WithAttrs([]slog.Attr{slog.Bool("sub1", true)})
			sub2 := h.WithAttrs([]slog.Attr{slog.Bool("sub2", true)})
			var wg sync.WaitGroup
			for i := 0; i < count; i++ {
				sub1Record := slog.NewRecord(time.Time{}, slog.LevelInfo, "hello from sub1", 0)
				sub1Record.AddAttrs(slog.Int("i", i))
				sub2Record := slog.NewRecord(time.Time{}, slog.LevelInfo, "hello from sub2", 0)
				sub2Record.AddAttrs(slog.Int("i", i))
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := sub1.Handle(ctx, sub1Record); err != nil {
						t.Error(err)
					}
					if err := sub2.Handle(ctx, sub2Record); err != nil {
						t.Error(err)
					}
				}()
			}
			wg.Wait()
			for i := 1; i <= 2; i++ {
				want := "hello from sub" + strconv.Itoa(i)
				n := strings.Count(buf.String(), want)
				if n != count {
					t.Fatalf("want %d occurrences of %q, got %d", count, want, n)
				}
			}
		})
	}
}
