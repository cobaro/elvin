// Generated by tpc version 0.6.3

package main

const (
	TerminalEOF = iota
	TerminalLPAREN
	TerminalRPAREN
	TerminalID
	TerminalCOMMA
	TerminalOR
	TerminalXOR
	TerminalAND
	TerminalEQ
	TerminalNEQ
	TerminalLT
	TerminalLE
	TerminalGT
	TerminalGE
	TerminalBANG
	TerminalSTRING
	TerminalBIT_OR
	TerminalBIT_XOR
	TerminalBIT_AND
	TerminalBIT_SHL
	TerminalBIT_SHR
	TerminalBIT_LSR
	TerminalPLUS
	TerminalMINUS
	TerminalTIMES
	TerminalDIV
	TerminalMOD
	TerminalINT32
	TerminalINT64
	TerminalREAL64
	TerminalNEG
)

type production struct {
	reduction       string
	nonTerminalType int
	count           int
}

var Productions = []production{

	// 0: <sub-exp> ::= <disjunction>
	{
		reduction:       "accept_sub",
		nonTerminalType: 0,
		count:           1,
	},

	// 1: <function-exp> ::= LPAREN <function-exp> RPAREN
	{
		reduction:       "identity2",
		nonTerminalType: 2,
		count:           3,
	},

	// 2: <function-exp> ::= <function>
	{
		reduction:       "identity",
		nonTerminalType: 2,
		count:           1,
	},

	// 3: <function> ::= ID LPAREN <args> RPAREN
	{
		reduction:       "create_function_n",
		nonTerminalType: 3,
		count:           4,
	},

	// 4: <function> ::= ID LPAREN RPAREN
	{
		reduction:       "create_function_0",
		nonTerminalType: 3,
		count:           3,
	},

	// 5: <args> ::= <args> COMMA <value>
	{
		reduction:       "extend_args",
		nonTerminalType: 4,
		count:           3,
	},

	// 6: <args> ::= <value>
	{
		reduction:       "create_args",
		nonTerminalType: 4,
		count:           1,
	},

	// 7: <disjunction> ::= <disjunction> OR <xor-exp>
	{
		reduction:       "extend_disjunction",
		nonTerminalType: 1,
		count:           3,
	},

	// 8: <disjunction> ::= <xor-exp>
	{
		reduction:       "create_disjunction",
		nonTerminalType: 1,
		count:           1,
	},

	// 9: <xor-exp> ::= <xor-exp> XOR <conjunction>
	{
		reduction:       "extend_xor_exp",
		nonTerminalType: 6,
		count:           3,
	},

	// 10: <xor-exp> ::= <conjunction>
	{
		reduction:       "create_xor_exp",
		nonTerminalType: 6,
		count:           1,
	},

	// 11: <conjunction> ::= <conjunction> AND <bool-exp>
	{
		reduction:       "extend_conjunction",
		nonTerminalType: 7,
		count:           3,
	},

	// 12: <conjunction> ::= <bool-exp>
	{
		reduction:       "create_conjunction",
		nonTerminalType: 7,
		count:           1,
	},

	// 13: <bool-exp> ::= <value> EQ <value>
	{
		reduction:       "create_eq_comparison",
		nonTerminalType: 8,
		count:           3,
	},

	// 14: <bool-exp> ::= <value> NEQ <value>
	{
		reduction:       "create_neq_comparison",
		nonTerminalType: 8,
		count:           3,
	},

	// 15: <bool-exp> ::= <bit-disjunction> LT <bit-disjunction>
	{
		reduction:       "create_lt_comparison",
		nonTerminalType: 8,
		count:           3,
	},

	// 16: <bool-exp> ::= <bit-disjunction> LE <bit-disjunction>
	{
		reduction:       "create_le_comparison",
		nonTerminalType: 8,
		count:           3,
	},

	// 17: <bool-exp> ::= <bit-disjunction> GT <bit-disjunction>
	{
		reduction:       "create_gt_comparison",
		nonTerminalType: 8,
		count:           3,
	},

	// 18: <bool-exp> ::= <bit-disjunction> GE <bit-disjunction>
	{
		reduction:       "create_ge_comparison",
		nonTerminalType: 8,
		count:           3,
	},

	// 19: <bool-exp> ::= <function-exp>
	{
		reduction:       "identity",
		nonTerminalType: 8,
		count:           1,
	},

	// 20: <bool-exp> ::= BANG <bool-exp>
	{
		reduction:       "create_not_op",
		nonTerminalType: 8,
		count:           2,
	},

	// 21: <bool-exp> ::= LPAREN <disjunction> RPAREN
	{
		reduction:       "identity2",
		nonTerminalType: 8,
		count:           3,
	},

	// 22: <value> ::= STRING
	{
		reduction:       "identity",
		nonTerminalType: 5,
		count:           1,
	},

	// 23: <value> ::= <bit-disjunction>
	{
		reduction:       "identity",
		nonTerminalType: 5,
		count:           1,
	},

	// 24: <bit-disjunction> ::= <bit-disjunction> BIT_OR <bit-xor-exp>
	{
		reduction:       "create_or_op",
		nonTerminalType: 9,
		count:           3,
	},

	// 25: <bit-disjunction> ::= <bit-xor-exp>
	{
		reduction:       "identity",
		nonTerminalType: 9,
		count:           1,
	},

	// 26: <bit-xor-exp> ::= <bit-xor-exp> BIT_XOR <bit-conjunction>
	{
		reduction:       "create_xor_op",
		nonTerminalType: 10,
		count:           3,
	},

	// 27: <bit-xor-exp> ::= <bit-conjunction>
	{
		reduction:       "identity",
		nonTerminalType: 10,
		count:           1,
	},

	// 28: <bit-conjunction> ::= <bit-conjunction> BIT_AND <bit-shift-exp>
	{
		reduction:       "create_and_op",
		nonTerminalType: 11,
		count:           3,
	},

	// 29: <bit-conjunction> ::= <bit-shift-exp>
	{
		reduction:       "identity",
		nonTerminalType: 11,
		count:           1,
	},

	// 30: <bit-shift-exp> ::= <bit-shift-exp> BIT_SHL <sum>
	{
		reduction:       "create_shl_op",
		nonTerminalType: 12,
		count:           3,
	},

	// 31: <bit-shift-exp> ::= <bit-shift-exp> BIT_SHR <sum>
	{
		reduction:       "create_shr_op",
		nonTerminalType: 12,
		count:           3,
	},

	// 32: <bit-shift-exp> ::= <bit-shift-exp> BIT_LSR <sum>
	{
		reduction:       "create_lsr_op",
		nonTerminalType: 12,
		count:           3,
	},

	// 33: <bit-shift-exp> ::= <sum>
	{
		reduction:       "identity",
		nonTerminalType: 12,
		count:           1,
	},

	// 34: <sum> ::= <sum> PLUS <product>
	{
		reduction:       "create_plus_op",
		nonTerminalType: 13,
		count:           3,
	},

	// 35: <sum> ::= <sum> MINUS <product>
	{
		reduction:       "create_minus_op",
		nonTerminalType: 13,
		count:           3,
	},

	// 36: <sum> ::= <product>
	{
		reduction:       "identity",
		nonTerminalType: 13,
		count:           1,
	},

	// 37: <product> ::= <product> TIMES <num-value>
	{
		reduction:       "create_times_op",
		nonTerminalType: 14,
		count:           3,
	},

	// 38: <product> ::= <product> DIV <num-value>
	{
		reduction:       "create_div_op",
		nonTerminalType: 14,
		count:           3,
	},

	// 39: <product> ::= <product> MOD <num-value>
	{
		reduction:       "create_mod_op",
		nonTerminalType: 14,
		count:           3,
	},

	// 40: <product> ::= <num-value>
	{
		reduction:       "identity",
		nonTerminalType: 14,
		count:           1,
	},

	// 41: <num-value> ::= INT32
	{
		reduction:       "identity",
		nonTerminalType: 15,
		count:           1,
	},

	// 42: <num-value> ::= INT64
	{
		reduction:       "identity",
		nonTerminalType: 15,
		count:           1,
	},

	// 43: <num-value> ::= REAL64
	{
		reduction:       "identity",
		nonTerminalType: 15,
		count:           1,
	},

	// 44: <num-value> ::= <name>
	{
		reduction:       "identity",
		nonTerminalType: 15,
		count:           1,
	},

	// 45: <num-value> ::= <function-exp>
	{
		reduction:       "identity",
		nonTerminalType: 15,
		count:           1,
	},

	// 46: <num-value> ::= PLUS <num-value>
	{
		reduction:       "create_uplus_op",
		nonTerminalType: 15,
		count:           2,
	},

	// 47: <num-value> ::= MINUS <num-value>
	{
		reduction:       "create_uminus_op",
		nonTerminalType: 15,
		count:           2,
	},

	// 48: <num-value> ::= NEG <num-value>
	{
		reduction:       "create_neg_op",
		nonTerminalType: 15,
		count:           2,
	},

	// 49: <num-value> ::= LPAREN <value> RPAREN
	{
		reduction:       "identity2",
		nonTerminalType: 15,
		count:           3,
	},

	// 50: <name> ::= ID
	{
		reduction:       "name_from_id",
		nonTerminalType: 16,
		count:           1,
	},
}

const ERR = -1
const ACC = 0

func R(x int) int {
	return x
}
func S(x int) int {
	return x + 51
}

var strTable = [][]int{
	{ERR, S(16), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(18), S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ACC, ERR, ERR, ERR, ERR, S(26), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(19), ERR, R(19), ERR, ERR, R(19), R(19), R(19), R(45), R(45), R(45), R(45), R(45), R(45), ERR, ERR, R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), ERR, ERR, ERR, ERR},
	{R(2), ERR, R(2), ERR, R(2), R(2), R(2), R(2), R(2), R(2), R(2), R(2), R(2), R(2), ERR, ERR, R(2), R(2), R(2), R(2), R(2), R(2), R(2), R(2), R(2), R(2), R(2), ERR, ERR, ERR, ERR},
	{ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(27), S(28), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(8), ERR, R(8), ERR, ERR, R(8), S(29), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(10), ERR, R(10), ERR, ERR, R(10), R(10), S(30), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(12), ERR, R(12), ERR, ERR, R(12), R(12), R(12), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{ERR, ERR, R(23), ERR, ERR, ERR, ERR, ERR, R(23), R(23), S(31), S(32), S(33), S(34), ERR, ERR, S(35), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(25), ERR, R(25), ERR, R(25), R(25), R(25), R(25), R(25), R(25), R(25), R(25), R(25), R(25), ERR, ERR, R(25), S(36), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(27), ERR, R(27), ERR, R(27), R(27), R(27), R(27), R(27), R(27), R(27), R(27), R(27), R(27), ERR, ERR, R(27), R(27), S(37), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(29), ERR, R(29), ERR, R(29), R(29), R(29), R(29), R(29), R(29), R(29), R(29), R(29), R(29), ERR, ERR, R(29), R(29), R(29), S(38), S(39), S(40), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(33), ERR, R(33), ERR, R(33), R(33), R(33), R(33), R(33), R(33), R(33), R(33), R(33), R(33), ERR, ERR, R(33), R(33), R(33), R(33), R(33), R(33), S(41), S(42), ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(36), ERR, R(36), ERR, R(36), R(36), R(36), R(36), R(36), R(36), R(36), R(36), R(36), R(36), ERR, ERR, R(36), R(36), R(36), R(36), R(36), R(36), R(36), R(36), S(43), S(44), S(45), ERR, ERR, ERR, ERR},
	{R(40), ERR, R(40), ERR, R(40), R(40), R(40), R(40), R(40), R(40), R(40), R(40), R(40), R(40), ERR, ERR, R(40), R(40), R(40), R(40), R(40), R(40), R(40), R(40), R(40), R(40), R(40), ERR, ERR, ERR, ERR},
	{R(44), ERR, R(44), ERR, R(44), R(44), R(44), R(44), R(44), R(44), R(44), R(44), R(44), R(44), ERR, ERR, R(44), R(44), R(44), R(44), R(44), R(44), R(44), R(44), R(44), R(44), R(44), ERR, ERR, ERR, ERR},
	{ERR, S(16), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(18), S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{R(50), S(49), R(50), ERR, R(50), R(50), R(50), R(50), R(50), R(50), R(50), R(50), R(50), R(50), ERR, ERR, R(50), R(50), R(50), R(50), R(50), R(50), R(50), R(50), R(50), R(50), R(50), ERR, ERR, ERR, ERR},
	{ERR, S(16), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(18), S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{R(22), ERR, R(22), ERR, R(22), R(22), R(22), R(22), R(22), R(22), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{R(41), ERR, R(41), ERR, R(41), R(41), R(41), R(41), R(41), R(41), R(41), R(41), R(41), R(41), ERR, ERR, R(41), R(41), R(41), R(41), R(41), R(41), R(41), R(41), R(41), R(41), R(41), ERR, ERR, ERR, ERR},
	{R(42), ERR, R(42), ERR, R(42), R(42), R(42), R(42), R(42), R(42), R(42), R(42), R(42), R(42), ERR, ERR, R(42), R(42), R(42), R(42), R(42), R(42), R(42), R(42), R(42), R(42), R(42), ERR, ERR, ERR, ERR},
	{R(43), ERR, R(43), ERR, R(43), R(43), R(43), R(43), R(43), R(43), R(43), R(43), R(43), R(43), ERR, ERR, R(43), R(43), R(43), R(43), R(43), R(43), R(43), R(43), R(43), R(43), R(43), ERR, ERR, ERR, ERR},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(16), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(18), S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(16), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(18), S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(16), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(18), S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, ERR, S(77), ERR, ERR, S(26), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{ERR, ERR, S(78), ERR, ERR, R(19), R(19), R(19), R(45), R(45), R(45), R(45), R(45), R(45), ERR, ERR, R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), ERR, ERR, ERR, ERR},
	{ERR, ERR, S(79), ERR, ERR, ERR, ERR, ERR, S(27), S(28), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{ERR, S(53), S(82), S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{R(20), ERR, R(20), ERR, ERR, R(20), R(20), R(20), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(45), ERR, R(45), ERR, R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), ERR, ERR, R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), ERR, ERR, ERR, ERR},
	{R(46), ERR, R(46), ERR, R(46), R(46), R(46), R(46), R(46), R(46), R(46), R(46), R(46), R(46), ERR, ERR, R(46), R(46), R(46), R(46), R(46), R(46), R(46), R(46), R(46), R(46), R(46), ERR, ERR, ERR, ERR},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{R(47), ERR, R(47), ERR, R(47), R(47), R(47), R(47), R(47), R(47), R(47), R(47), R(47), R(47), ERR, ERR, R(47), R(47), R(47), R(47), R(47), R(47), R(47), R(47), R(47), R(47), R(47), ERR, ERR, ERR, ERR},
	{R(48), ERR, R(48), ERR, R(48), R(48), R(48), R(48), R(48), R(48), R(48), R(48), R(48), R(48), ERR, ERR, R(48), R(48), R(48), R(48), R(48), R(48), R(48), R(48), R(48), R(48), R(48), ERR, ERR, ERR, ERR},
	{R(7), ERR, R(7), ERR, ERR, R(7), S(29), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(13), ERR, R(13), ERR, ERR, R(13), R(13), R(13), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(23), ERR, R(23), ERR, R(23), R(23), R(23), R(23), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(35), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(14), ERR, R(14), ERR, ERR, R(14), R(14), R(14), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(9), ERR, R(9), ERR, ERR, R(9), R(9), S(30), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(11), ERR, R(11), ERR, ERR, R(11), R(11), R(11), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(15), ERR, R(15), ERR, ERR, R(15), R(15), R(15), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(35), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(16), ERR, R(16), ERR, ERR, R(16), R(16), R(16), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(35), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(17), ERR, R(17), ERR, ERR, R(17), R(17), R(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(35), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(18), ERR, R(18), ERR, ERR, R(18), R(18), R(18), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(35), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(24), ERR, R(24), ERR, R(24), R(24), R(24), R(24), R(24), R(24), R(24), R(24), R(24), R(24), ERR, ERR, R(24), S(36), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(26), ERR, R(26), ERR, R(26), R(26), R(26), R(26), R(26), R(26), R(26), R(26), R(26), R(26), ERR, ERR, R(26), R(26), S(37), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(28), ERR, R(28), ERR, R(28), R(28), R(28), R(28), R(28), R(28), R(28), R(28), R(28), R(28), ERR, ERR, R(28), R(28), R(28), S(38), S(39), S(40), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(30), ERR, R(30), ERR, R(30), R(30), R(30), R(30), R(30), R(30), R(30), R(30), R(30), R(30), ERR, ERR, R(30), R(30), R(30), R(30), R(30), R(30), S(41), S(42), ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(31), ERR, R(31), ERR, R(31), R(31), R(31), R(31), R(31), R(31), R(31), R(31), R(31), R(31), ERR, ERR, R(31), R(31), R(31), R(31), R(31), R(31), S(41), S(42), ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(32), ERR, R(32), ERR, R(32), R(32), R(32), R(32), R(32), R(32), R(32), R(32), R(32), R(32), ERR, ERR, R(32), R(32), R(32), R(32), R(32), R(32), S(41), S(42), ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(34), ERR, R(34), ERR, R(34), R(34), R(34), R(34), R(34), R(34), R(34), R(34), R(34), R(34), ERR, ERR, R(34), R(34), R(34), R(34), R(34), R(34), R(34), R(34), S(43), S(44), S(45), ERR, ERR, ERR, ERR},
	{R(35), ERR, R(35), ERR, R(35), R(35), R(35), R(35), R(35), R(35), R(35), R(35), R(35), R(35), ERR, ERR, R(35), R(35), R(35), R(35), R(35), R(35), R(35), R(35), S(43), S(44), S(45), ERR, ERR, ERR, ERR},
	{R(37), ERR, R(37), ERR, R(37), R(37), R(37), R(37), R(37), R(37), R(37), R(37), R(37), R(37), ERR, ERR, R(37), R(37), R(37), R(37), R(37), R(37), R(37), R(37), R(37), R(37), R(37), ERR, ERR, ERR, ERR},
	{R(38), ERR, R(38), ERR, R(38), R(38), R(38), R(38), R(38), R(38), R(38), R(38), R(38), R(38), ERR, ERR, R(38), R(38), R(38), R(38), R(38), R(38), R(38), R(38), R(38), R(38), R(38), ERR, ERR, ERR, ERR},
	{R(39), ERR, R(39), ERR, R(39), R(39), R(39), R(39), R(39), R(39), R(39), R(39), R(39), R(39), ERR, ERR, R(39), R(39), R(39), R(39), R(39), R(39), R(39), R(39), R(39), R(39), R(39), ERR, ERR, ERR, ERR},
	{R(21), ERR, R(21), ERR, ERR, R(21), R(21), R(21), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(1), ERR, R(1), ERR, R(1), R(1), R(1), R(1), R(1), R(1), R(1), R(1), R(1), R(1), ERR, ERR, R(1), R(1), R(1), R(1), R(1), R(1), R(1), R(1), R(1), R(1), R(1), ERR, ERR, ERR, ERR},
	{R(49), ERR, R(49), ERR, R(49), R(49), R(49), R(49), R(49), R(49), R(49), R(49), R(49), R(49), ERR, ERR, R(49), R(49), R(49), R(49), R(49), R(49), R(49), R(49), R(49), R(49), R(49), ERR, ERR, ERR, ERR},
	{ERR, ERR, S(85), ERR, S(86), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{ERR, ERR, R(6), ERR, R(6), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(4), ERR, R(4), ERR, R(4), R(4), R(4), R(4), R(4), R(4), R(4), R(4), R(4), R(4), ERR, ERR, R(4), R(4), R(4), R(4), R(4), R(4), R(4), R(4), R(4), R(4), R(4), ERR, ERR, ERR, ERR},
	{ERR, ERR, S(78), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), R(45), ERR, ERR, ERR, ERR},
	{ERR, ERR, S(79), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR},
	{R(3), ERR, R(3), ERR, R(3), R(3), R(3), R(3), R(3), R(3), R(3), R(3), R(3), R(3), ERR, ERR, R(3), R(3), R(3), R(3), R(3), R(3), R(3), R(3), R(3), R(3), R(3), ERR, ERR, ERR, ERR},
	{ERR, S(53), ERR, S(17), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, S(19), ERR, ERR, ERR, ERR, ERR, ERR, S(20), S(21), ERR, ERR, ERR, S(22), S(23), S(24), S(25)},
	{ERR, ERR, R(5), ERR, R(5), ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR}}

var GotoTable = [][]int{
	{0, 1, 2, 3, 0, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 46, 47, 3, 0, 48, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 2, 3, 0, 4, 0, 0, 50, 8, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 52, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 54, 15},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 55, 15},
	{0, 0, 2, 3, 0, 4, 56, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 57, 0, 0, 0, 58, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 59, 0, 0, 0, 58, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 2, 3, 0, 4, 0, 60, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 2, 3, 0, 4, 0, 0, 61, 8, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 62, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 63, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 64, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 65, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 66, 10, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 67, 11, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 68, 12, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 70, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 71, 13, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 72, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 73, 14, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 74, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 75, 15},
	{0, 0, 51, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 76, 15},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 51, 3, 80, 81, 0, 0, 0, 58, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 83, 3, 0, 84, 0, 0, 0, 58, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 51, 3, 0, 87, 0, 0, 0, 58, 9, 10, 11, 12, 13, 14, 15},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}
