package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
	"fmt"
)

const (
	inLimbo = iota
	inIdentifier
	inSingleQuotedString
	inDoubleQuotedString
	inNumber
)

const invalidRune = '\u0000'
const terminalError = -1


type tokenInfo struct {
	token int
	value string
}

func isInitialNumberChar(r rune) bool {
	return  strings.ContainsRune("0123456789-", r)
}

func isNumberChar(r rune) bool {
	return strings.ContainsRune("0123456789-.eEIL", r)
}

func isInitialNameChar(r rune) bool {
	return strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_", r)
}

func isNameChar(r rune) bool {
	return !(r < 32 || r > 126 || strings.ContainsRune("\\()\"', ", r))
}

func lexer(buf string) []tokenInfo {
	var tokens []tokenInfo
	var i = 0
	var mode = 0
	var tokenValue strings.Builder

	for {
		expr := buf[i:]
		rune1, len1 := utf8.DecodeRuneInString(expr)
		if len1 == 0 {
			break
		}
		rune2, len2 := utf8.DecodeRuneInString(expr[len1:])
		rune3, _ := utf8.DecodeRuneInString(expr[len1 + len2:])

		s2 := string([]rune{rune1, rune2})
		s3 := string([]rune{rune1, rune2, rune3})

		if mode == inLimbo {
			if rune1 == '\'' {
				mode = inSingleQuotedString
				tokenValue.Reset()
			} else if rune1 == '"' {
				mode = inDoubleQuotedString
				tokenValue.Reset()
			} else if s3 == ">>>" {  // Must check before '>>' (below)
				tokens = append(tokens, tokenInfo{TerminalBIT_LSR, ""})
				i += 2
			} else if s2 == ">>" {
				tokens = append(tokens, tokenInfo{TerminalBIT_SHR, ""})
				i += 1
			} else if s2 == "<<" {
				tokens = append(tokens, tokenInfo{TerminalBIT_SHL, ""})
				i += 1
			} else if s2 == "&&" {
				tokens = append(tokens, tokenInfo{TerminalAND, ""})
				i += 1
			} else if s2 == "||" {
				tokens = append(tokens, tokenInfo{TerminalOR, ""})
				i += 1
			} else if s2 == "^^" {
				tokens = append(tokens, tokenInfo{TerminalXOR, ""})
				i += 1
			} else if s2 == "==" {
				tokens = append(tokens, tokenInfo{TerminalEQ, ""})
				i += 1
			} else if s2 == "!=" {
				tokens = append(tokens, tokenInfo{TerminalNEQ, ""})
				i += 1
			} else if s2 == "<=" {
				tokens = append(tokens, tokenInfo{TerminalLE, ""})
				i += 1
			} else if s2 == ">=" {
				tokens = append(tokens, tokenInfo{TerminalGE, ""})
				i += 1
			} else if rune1 == '\\' {
				mode = inIdentifier
				tokenValue.WriteRune(rune2)
				i += len2  // +=1 at end of loop covers backslash; this is for the real rune
			} else if rune1 == '(' {
				tokens = append(tokens, tokenInfo{TerminalLPAREN, ""})
			} else if rune1 == ')' {
				tokens = append(tokens, tokenInfo{TerminalRPAREN, ""})
			} else if rune1 == ',' {
				tokens = append(tokens, tokenInfo{TerminalCOMMA, ""})
			} else if rune1 == '&' {
				tokens = append(tokens, tokenInfo{TerminalBIT_AND, ""})
			} else if rune1 == '~' {
				tokens = append(tokens, tokenInfo{TerminalNEG, ""})
			} else if rune1 == '|' {
				tokens = append(tokens, tokenInfo{TerminalBIT_OR, ""})
			} else if rune1 == '^' {
				tokens = append(tokens, tokenInfo{TerminalBIT_XOR, ""})
			} else if rune1 == '<' {
				tokens = append(tokens, tokenInfo{TerminalLT, ""})
			} else if rune1 == '>' {
				tokens = append(tokens, tokenInfo{TerminalGT, ""})
			} else if rune1 == '+' {
				tokens = append(tokens, tokenInfo{TerminalPLUS, ""})
			} else if rune1 == '-' {
				tokens = append(tokens, tokenInfo{TerminalMINUS, ""})
			} else if rune1 == '*' {
				tokens = append(tokens, tokenInfo{TerminalTIMES, ""})
			} else if rune1 == '/' {
				tokens = append(tokens, tokenInfo{TerminalDIV, ""})
			} else if rune1 == '%' {
				tokens = append(tokens, tokenInfo{TerminalMOD, ""})
			} else if rune1 == '!' {
				tokens = append(tokens, tokenInfo{TerminalNEG, ""})
			} else if unicode.IsSpace(rune1) {
				// Whitespace is ignored in limbo mode
			} else if isInitialNameChar(rune1) {
				mode = inIdentifier
				tokenValue.WriteRune(rune1)
			} else if isInitialNumberChar(rune1) {
				mode = inNumber
				tokenValue.WriteRune(rune1)
			} else if rune1 == invalidRune {
				tokens = append(tokens, tokenInfo{TerminalEOF, ""})
				break
			} else {
				err := fmt.Sprintf("Error at index: %d", i)
				tokens = append(tokens, tokenInfo{terminalError, err})
			}

		} else if mode == inIdentifier {
			if rune1 == '\\' {
				tokenValue.WriteRune(rune2)
				i += len2
			} else if rune1 == '(' {
				tokens = append(tokens, tokenInfo{TerminalID, tokenValue.String()})
				tokens = append(tokens, tokenInfo{TerminalLPAREN, ""})
				tokenValue.Reset()
				mode = inLimbo
			} else if rune1 == ',' {
				tokens = append(tokens, tokenInfo{TerminalID, tokenValue.String()})
				tokens = append(tokens, tokenInfo{TerminalCOMMA, ""})
				tokenValue.Reset()
				mode = inLimbo
			} else if isNameChar(rune1) {
				tokenValue.WriteRune(rune1)
			} else {
				tokens = append(tokens, tokenInfo{TerminalID, tokenValue.String()})
				tokenValue.Reset()
				mode = inLimbo
				// Recheck this rune in limbo mode.
				continue
			}

		} else if mode == inSingleQuotedString {
			if rune1 == '\\' {
				tokenValue.WriteRune(rune2)
				i += len2
			} else if rune1 == '\'' {
				tokens = append(tokens, tokenInfo{TerminalSTRING, tokenValue.String()})
				tokenValue.Reset()
				mode = inLimbo
			} else if rune1 == invalidRune {
				err := fmt.Sprintf("String missing closing single quote at index %d", i)
				tokens = append(tokens, tokenInfo{terminalError, err})
			} else {
				tokenValue.WriteRune(rune1)
			}

		} else if mode == inDoubleQuotedString {
			if rune1 == '\\' {
				tokenValue.WriteRune(rune2)
				i += len2
			} else if rune1 == '"' {
				tokens = append(tokens, tokenInfo{TerminalSTRING, tokenValue.String()})
				tokenValue.Reset()
				mode = inLimbo
			} else if rune1 == invalidRune {
				err := fmt.Sprintf("String missing closing double quote at index %d", i)
				tokens = append(tokens, tokenInfo{terminalError, err})
			} else {
				tokenValue.WriteRune(rune1)
			}

		} else if mode == inNumber {
			if isNumberChar(rune1) {
				tokenValue.WriteRune(rune1)
			} else {
				tokens = append(tokens, tokenInfo{TerminalINT32, tokenValue.String()})
				tokenValue.Reset()
				mode = inLimbo
				// Recheck this rune in limbo mode.
				continue
			}

		} else {
			panic("panic")
		}

		fmt.Printf("i=%d, r1=%v, s2=%s, s3=%s, mode=%d, tokenValue=%s, tokens=%#v\n",
			i, rune1, s2, s3, mode, tokenValue.String(), tokens)
		i += len1
	}

	return tokens
}