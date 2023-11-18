package src

import (
	"container/list"
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/verniyyy/verniy-mq/src/testhelper"
)

func TestNewQueue(t *testing.T) {
	tests := []struct {
		name string
		want Queue[string]
	}{
		{
			name: "can be get new queue",
			want: Queue[string](&queue[string]{
				m: sync.Mutex{},
				l: list.New(),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueue[string](); !reflect.DeepEqual(got, tt.want) {
				t.Errorf(cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_queue_Enqueue(t *testing.T) {
	tests := []struct {
		name    string
		v       string
		wantErr error
	}{
		{
			name:    "can be get new queue",
			v:       "example value",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &queue[string]{
				l: list.New().Init(),
			}
			if err := q.Enqueue(tt.v); !testhelper.EqualError(err, tt.wantErr) {
				t.Errorf("error = %v, want error = %v", err, tt.wantErr)
			}
		})
	}
}

func Test_queue_Dequeue(t *testing.T) {
	type fields struct {
		l *list.List
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr error
	}{
		{
			name: "can be dequeued",
			fields: fields{
				l: func() *list.List {
					l := list.New()
					l.PushBack("foo")
					l.PushBack("bar")
					l.PushBack("baz")
					return l
				}(),
			},
			want:    "foo",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &queue[string]{
				l: tt.fields.l,
			}
			got, err := q.Dequeue()
			if got != tt.want {
				t.Errorf(cmp.Diff(got, tt.want))
			}
			if !testhelper.EqualError(err, tt.wantErr) {
				t.Errorf("error = %v, want error = %v", err, tt.wantErr)
			}
		})
	}
}
