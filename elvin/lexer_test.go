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

package main

import (
	"testing"
)

func TestRightShiftZero(t *testing.T) {
	s := " 42 >>> 3 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalBIT_LSR {
		t.Error("Expected token TerminalBIT_LSR; lexer reported ", tokens[1].token)
	}
}

func TestRightShift(t *testing.T) {
	s := " 42 >> 3 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalBIT_SHR {
		t.Error("Expected token TerminalBIT_SHR; lexer reported ", tokens[1].token)
	}
}

func TestLeftShift(t *testing.T) {
	s := " 42 << 3 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalBIT_SHL {
		t.Error("Expected token TerminalBIT_SHL; lexer reported ", tokens[1].token)
	}
}

func TestLogicalAnd(t *testing.T) {
	s := " a && b "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalAND {
		t.Error("Expected token TerminalAND; lexer reported ", tokens[1].token)
	}
}

func TestLogicalOr(t *testing.T) {
	s := " a || b "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalOR {
		t.Error("Expected token TerminalOR; lexer reported ", tokens[1].token)
	}
}

func TestLogicalXor(t *testing.T) {
	s := " a ^^ b "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalXOR {
		t.Error("Expected token TerminalXOR; lexer reported ", tokens[1].token)
	}
}

func TestEquals(t *testing.T) {
	s := " a == b "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalEQ {
		t.Error("Expected token TerminalEQ; lexer reported ", tokens[1].token)
	}
}

func TestNotEquals(t *testing.T) {
	s := " a != b "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalNEQ {
		t.Error("Expected token TerminalNEQ; lexer reported ", tokens[1].token)
	}
}

func TestLessThanOrEquals(t *testing.T) {
	s := " a <= b "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalLE {
		t.Error("Expected token TerminalLE; lexer reported ", tokens[1].token)
	}
}

func TestGreaterThanOrEquals(t *testing.T) {
	s := " a >= b "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalGE {
		t.Error("Expected token TerminalGE; lexer reported ", tokens[1].token)
	}
}

func TestIdentifier(t *testing.T) {
	s := " a == 1 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[0].token != TerminalID {
		t.Error("Expected token TerminalID; lexer reported ", tokens[0].token)
	}

	if tokens[0].value != "a" {
		t.Error("Expected token value 'a'; lexer reported ", tokens[0].value)
	}
}

func TestIdentifierWithLeadingEscape(t *testing.T) {
	t.Skip("Until it actually works")
	s := " \require == 1 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[0].token != TerminalID {
		t.Error("Expected token TerminalID; lexer reported ", tokens[0].token)
	}

	if tokens[0].value != "require" {
		t.Error("Expected token value ''; lexer reported ", tokens[0].value)
	}
}

func TestLeftParenthesis(t *testing.T) {
	s := " (foo == 1 || bar == 1) && baz == 3 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 14 {
		t.Error("Expected 14 tokens; lexer reported ", len(tokens))
	}

	if tokens[0].token != TerminalLPAREN {
		t.Error("Expected LPAREN; lexer reported ", tokens[0].token)
	}
}

func TestRightParenthesis(t *testing.T) {
	s := " (foo == 1 || bar == 1) && baz == 3 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 14 {
		t.Error("Expected 14 tokens; lexer reported ", len(tokens))
	}

	if tokens[8].token != TerminalRPAREN {
		t.Error("Expected RPAREN; lexer reported ", tokens[8].token)
	}
}

func TestComma(t *testing.T) {
	s := "equals(Group, 'foo', 'bar', 'baz')"
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 11 {
		t.Error("Expecting 11 tokens; lexer reported ", len(tokens))
	}

	if tokens[3].token != TerminalCOMMA {
		t.Error("Expected COMMA; lexer reported ", tokens[3].token)
	}
}

func TestBitwiseAnd(t *testing.T) {
	s := " b & 15 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalBIT_AND {
		t.Error("Expected bitwise-AND; lexer reported ", tokens[1].token)
	}
}

func TestBitwiseOr(t *testing.T) {
	s := " b | 15 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalBIT_OR {
		t.Error("Expected bitwise-OR; lexer reported ", tokens[1].token)
	}
}

func TestBitwiseXor(t *testing.T) {
	s := " b ^ 15 "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 4 {
		t.Error("Expected 4 tokens; lexer reported ", len(tokens))
	}

	if tokens[1].token != TerminalBIT_XOR {
		t.Error("Expected bitwise-XOR; lexer reported ", tokens[1].token)
	}
}

func TestBitwiseComplement(t *testing.T) {
	s := " ~b "
	var tokens []tokenInfo
	tokens = lexer(s)
	if len(tokens) != 3 {
		t.Error("Expected 3 tokens; lexer reported ", len(tokens))
	}

	if tokens[0].token != TerminalNEG {
		t.Error("Expected bitwise-complement; lexer reported ", tokens[0].token)
	}
}
