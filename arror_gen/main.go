package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync/atomic"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/apache/arrow/go/v18/arrow/memory"
)

var pool = memory.NewGoAllocator()

const (
	columnCount = 300
	rowsCount   = 1000
	precision   = 7
	scale       = 6
)

func main() {

	server := newServer()

	/* file, err := os.OpenFile("../my-grid/public/my_arrow.arrow", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("fail on open file 'my_arrow.arrow': %v", err)
	}
	defer file.Close()

	err = server.GenerateTable(file)

	if err != nil {
		log.Fatalf("failed generate data: %v", err)
	} */

	addFileRoute("table", server.GenerateTable)
	addFileRoute("update", server.GenerateUpdate)

	log.Println("start on 8081")

	http.ListenAndServe("localhost:8081", nil)
}

func addFileRoute(name string, handler func(io.Writer) error) {
	fileName := fmt.Sprintf("filename=%s.arrow", name)
	http.HandleFunc(fmt.Sprintf("/%s", name), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/octet-stream")
		w.Header().Add("Content-Disposition", fileName)
		w.Header().Add("Access-Control-Allow-Origin", "*")
		err := handler(w)
		if err != nil {
			slog.Error(fmt.Sprintf("fail on /%s", name), slog.Any("error", err))
			w.Write([]byte(err.Error()))
		}

		log.Println("call /", name)
	})
}

type server struct {
	schema   *arrow.Schema
	totalRow *atomic.Int32
}

func newServer() server {
	columns := make([]arrow.Field, columnCount+1)

	columns[0] = arrow.Field{Name: "ID", Type: arrow.PrimitiveTypes.Int32}

	for i := 1; i < len(columns); i++ {
		columns[i] = arrow.Field{Name: fmt.Sprintf("dwe %%%d", i), Type: &arrow.Float64Type{}}
	}

	schema := arrow.NewSchema(
		columns,
		nil,
	)

	return server{schema: schema, totalRow: &atomic.Int32{}}
}

func (s *server) GenerateUpdate(w io.Writer) error {
	added, err := s.generateData(w, 0, 10, ipc.WithLZ4())
	if err != nil {
		return fmt.Errorf("fail to update exists rows: %w", err)
	}

	s.totalRow.Add(int32(added))

	return nil
}

func (s *server) GenerateTable(w io.Writer) error {
	added, err := s.generateData(w, 0, rowsCount, ipc.WithSchema(s.schema), ipc.WithLZ4())
	if err == nil {
		s.totalRow.Swap(int32(added))
	}

	return err
}

func (s *server) generateData(w io.Writer, start int, count int, opts ...ipc.Option) (int, error) {

	b := array.NewRecordBuilder(pool, s.schema)
	defer b.Release()

	for i := 0; i < count; i++ {
		b.Field(0).(*array.Int32Builder).Append(int32(start + i + 1))
	}

	for j := 1; j < len(s.schema.Fields()); j++ {

		f := b.Field(j)
		row, ok := f.(*array.Float64Builder)

		if !ok {
			return 0, fmt.Errorf("can't cast field %v to Float64Builder", f)
		}

		for i := 0; i < count; i++ {
			val := rand.Float64()

			row.Append(val)
		}
	}

	rec1 := b.NewRecord()
	defer rec1.Release()

	opts = append(opts, ipc.WithAllocator(pool))

	ww := ipc.NewWriter(w, opts...)
	defer ww.Close()

	err := ww.Write(rec1)
	if err != nil {
		return 0, fmt.Errorf("fail to write arrow data: %w", err)
	}

	added := (start + count) - int(s.totalRow.Load())

	if added > 0 {
		return added, nil
	}

	return 0, nil
}
