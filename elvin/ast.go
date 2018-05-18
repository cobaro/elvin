// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package elvin

const (
	EmptyTypeCode = 0
	NameTypeCode = 1
	Int32TypeCode = 2
	Int64TypeCode = 3
	Real64TypeCode = 4
	StringTypeCode = 5
	EqualsTypeCode = 8
	NotEqualsTypeCode = 9
	LessThanTypeCode = 10
	LessThanOrEqualsTypeCode = 11
	GreaterThanTypeCode = 12
	GreaterThanOrEqualsTypeCode = 13
	LogicalOrTypeCode = 16
	LogicalExclusiveOrTypeCode = 17
	LogicalAndTypeCode = 18
	LogicalNotTypeCode = 19
	UnaryPlusTypeCode = 20
	UnaryMinusTypeCode = 21
	MultiplyTypeCode = 22
	DivideTypeCode = 23
	ModuloTypeCode = 24
	AddTypeCode = 25
	SubtractTypeCode = 26
	ShiftLeftTypeCode = 27
	ShiftRightTypeCode = 28
	LogicalShiftRightTypeCode = 29
	BinaryAndTypeCode = 30
	BinaryExclusiveOrTypeCode = 31
	BinaryOrTypeCode = 32
	BinaryNotTypeCode = 33
	FuncInt32TypeCode = 40
	FuncInt64TypeCode = 41
	FuncReal64TypeCode = 42
	FuncStringTypeCode = 43
	FuncOpaqueTypeCode = 44
	FuncNanTypeCode = 45
	FuncBeginsWithTypeCode = 48
	FuncContainsTypeCode = 49
	FuncEndsWithTypeCode = 50
	FuncWildcardTypeCode = 51
	FuncRegexTypeCode = 52
	FuncFoldCaseTypeCode = 56
	FuncDecomposeTypeCode = 57
	FuncDecomposeCompatTypeCode = 58
	FuncRequireTypeCode = 64
	FuncEqualsTypeCode = 65
	FuncSizeTypeCode = 66
)

const (
	LukTrue = 1
	LukFalse = 0
	LukBottom = -1
)

type AST struct {
	TypeCode int
	Value interface{}
	ID int
	BaseType int
}



func (node *AST) match(n map[string]interface{}) bool {
	return node.eval(n) == LukTrue
}


func (node *AST) eval(n map[string]interface{}) int {
	switch node.TypeCode {
	case FuncRequireTypeCode:
		if _, ok := n[node.Value]; !ok {
			return LukBottom
		}
		return LukTrue
	}
}

