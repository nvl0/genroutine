package test

import (
	"context"
	"errors"
	"genroutine"
	"genroutine/transaction"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoopReturnDataList(t *testing.T) {
	r := require.New(t)

	type fields struct {
		sm *transaction.MockSessionManager
		ts *transaction.MockSession
	}
	type args[P, R any] struct {
		f         genroutine.LoadDataList[P, R]
		paramList []P
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args[int, int]
		err     error
		data    []int
	}{
		{
			name: "успешный результат",
			prepare: func(f *fields) {
				f.sm.EXPECT().CreateSession().Return(f.ts).Times(2)
				f.ts.EXPECT().Start().Return(nil).Times(2)
				f.ts.EXPECT().Rollback().Return(nil).Times(4)
			},
			args: args[int, int]{
				f: func(ts transaction.Session, param int) ([]int, error) {
					return []int{3, 4}, nil
				},
				paramList: []int{1, 2},
			},
			err:  nil,
			data: []int{3, 4, 3, 4},
		},
		{
			name: "неуспешный результат",
			prepare: func(f *fields) {
				f.sm.EXPECT().CreateSession().Return(f.ts).AnyTimes()
				f.ts.EXPECT().Start().Return(nil).AnyTimes()
				f.ts.EXPECT().Rollback().Return(nil).AnyTimes()
			},
			args: args[int, int]{
				f: func(ts transaction.Session, param int) ([]int, error) {
					return nil, errors.New("some err")
				},
				paramList: []int{1, 2},
			},
			err:  errors.New("some err"),
			data: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				sm: transaction.NewMockSessionManager(ctrl),
				ts: transaction.NewMockSession(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			data, err := genroutine.LoopReturnDataList(ctx, f.sm, tt.args.f, tt.args.paramList)
			r.Equal(tt.err, err)
			r.Equal(tt.data, data)
		})
	}
}

func TestOffsetLoopReturnErr(t *testing.T) {
	r := require.New(t)

	type fields struct {
		sm *transaction.MockSessionManager
		ts *transaction.MockSession
	}
	type args[P any] struct {
		f         genroutine.ExecList[P]
		paramList []P
		offset    int
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args[int]
		err     error
	}{
		{
			name: "успешный результат",
			prepare: func(f *fields) {
				f.sm.EXPECT().CreateSession().Return(f.ts).Times(2)
				f.ts.EXPECT().Start().Return(nil).Times(2)
				f.ts.EXPECT().Commit().Return(nil).Times(2)
				f.ts.EXPECT().Rollback().Return(nil).Times(2)
			},
			args: args[int]{
				f: func(ts transaction.Session, paramList []int) error {
					return nil
				},
				paramList: []int{1, 2, 3},
				offset:    2,
			},
			err: nil,
		},
		{
			name: "неуспешный результат",
			prepare: func(f *fields) {
				f.sm.EXPECT().CreateSession().Return(f.ts).AnyTimes()
				f.ts.EXPECT().Start().Return(nil).AnyTimes()
				f.ts.EXPECT().Commit().Return(nil).AnyTimes()
				f.ts.EXPECT().Rollback().Return(nil).AnyTimes()
			},
			args: args[int]{
				f: func(ts transaction.Session, paramList []int) error {
					return errors.New("some err")
				},
				paramList: []int{1, 2, 3},
				offset:    2,
			},
			err: errors.New("some err"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				sm: transaction.NewMockSessionManager(ctrl),
				ts: transaction.NewMockSession(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			r.Equal(tt.err, genroutine.OffsetLoopReturnErr(ctx, f.sm, tt.args.f,
				tt.args.paramList, tt.args.offset))
		})
	}
}
