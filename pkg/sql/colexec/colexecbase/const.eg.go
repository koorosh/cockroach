// Code generated by execgen; DO NOT EDIT.
// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexecbase

import (
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecutils"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/errors"
)

// Workaround for bazel auto-generated code. goimports does not automatically
// pick up the right packages when run within the bazel sandbox.
var (
	_ apd.Context
	_ duration.Duration
	_ json.JSON
)

// NewConstOp creates a new operator that produces a constant value constVal of
// type t at index outputIdx.
func NewConstOp(
	allocator *colmem.Allocator,
	input colexecop.Operator,
	t *types.T,
	constVal interface{},
	outputIdx int,
) (colexecop.Operator, error) {
	input = colexecutils.NewVectorTypeEnforcer(allocator, input, t, outputIdx)
	switch typeconv.TypeFamilyToCanonicalTypeFamily(t.Family()) {
	case types.BoolFamily:
		switch t.Width() {
		case -1:
		default:
			return &constBoolOp{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(bool),
			}, nil
		}
	case types.BytesFamily:
		switch t.Width() {
		case -1:
		default:
			return &constBytesOp{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.([]byte),
			}, nil
		}
	case types.DecimalFamily:
		switch t.Width() {
		case -1:
		default:
			return &constDecimalOp{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(apd.Decimal),
			}, nil
		}
	case types.IntFamily:
		switch t.Width() {
		case 16:
			return &constInt16Op{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(int16),
			}, nil
		case 32:
			return &constInt32Op{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(int32),
			}, nil
		case -1:
		default:
			return &constInt64Op{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(int64),
			}, nil
		}
	case types.FloatFamily:
		switch t.Width() {
		case -1:
		default:
			return &constFloat64Op{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(float64),
			}, nil
		}
	case types.TimestampTZFamily:
		switch t.Width() {
		case -1:
		default:
			return &constTimestampOp{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(time.Time),
			}, nil
		}
	case types.IntervalFamily:
		switch t.Width() {
		case -1:
		default:
			return &constIntervalOp{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(duration.Duration),
			}, nil
		}
	case types.JsonFamily:
		switch t.Width() {
		case -1:
		default:
			return &constJSONOp{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(json.JSON),
			}, nil
		}
	case typeconv.DatumVecCanonicalTypeFamily:
		switch t.Width() {
		case -1:
		default:
			return &constDatumOp{
				OneInputHelper: colexecop.MakeOneInputHelper(input),
				allocator:      allocator,
				outputIdx:      outputIdx,
				constVal:       constVal.(interface{}),
			}, nil
		}
	}
	return nil, errors.Errorf("unsupported const type %s", t.Name())
}

type constBoolOp struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  bool
}

func (c constBoolOp) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Bool()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					//gcassert:bce
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constBytesOp struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  []byte
}

func (c constBytesOp) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Bytes()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constDecimalOp struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  apd.Decimal
}

func (c constDecimalOp) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Decimal()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					//gcassert:bce
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constInt16Op struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  int16
}

func (c constInt16Op) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Int16()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					//gcassert:bce
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constInt32Op struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  int32
}

func (c constInt32Op) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Int32()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					//gcassert:bce
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constInt64Op struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  int64
}

func (c constInt64Op) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Int64()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					//gcassert:bce
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constFloat64Op struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  float64
}

func (c constFloat64Op) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Float64()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					//gcassert:bce
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constTimestampOp struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  time.Time
}

func (c constTimestampOp) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Timestamp()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					//gcassert:bce
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constIntervalOp struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  duration.Duration
}

func (c constIntervalOp) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Interval()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					//gcassert:bce
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constJSONOp struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  json.JSON
}

func (c constJSONOp) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.JSON()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

type constDatumOp struct {
	colexecop.OneInputHelper

	allocator *colmem.Allocator
	outputIdx int
	constVal  interface{}
}

func (c constDatumOp) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}
	vec := batch.ColVec(c.outputIdx)
	col := vec.Datum()
	c.allocator.PerformOperation(
		[]*coldata.Vec{vec},
		func() {
			// Shallow copy col to work around Go issue
			// https://github.com/golang/go/issues/39756 which prevents bound check
			// elimination from working in this case.
			col := col
			if sel := batch.Selection(); sel != nil {
				for _, i := range sel[:n] {
					col.Set(i, c.constVal)
				}
			} else {
				_ = col.Get(n - 1)
				for i := 0; i < n; i++ {
					col.Set(i, c.constVal)
				}
			}
		},
	)
	return batch
}

// NewConstNullOp creates a new operator that produces a constant (untyped) NULL
// value at index outputIdx.
func NewConstNullOp(
	allocator *colmem.Allocator, input colexecop.Operator, outputIdx int,
) colexecop.Operator {
	input = colexecutils.NewVectorTypeEnforcer(allocator, input, types.Unknown, outputIdx)
	return &constNullOp{
		OneInputHelper: colexecop.MakeOneInputHelper(input),
		outputIdx:      outputIdx,
	}
}

type constNullOp struct {
	colexecop.OneInputHelper
	outputIdx int
}

var _ colexecop.Operator = &constNullOp{}

func (c constNullOp) Next() coldata.Batch {
	batch := c.Input.Next()
	n := batch.Length()
	if n == 0 {
		return coldata.ZeroBatch
	}

	batch.ColVec(c.outputIdx).Nulls().SetNulls()
	return batch
}
