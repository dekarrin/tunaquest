package tunascript

import (
	"fmt"
	"strings"
)

// tunascript execution engine

// FuncCall is the implementation of a function in tunaquest. It receives some
// number of values defined externally and returns a Value of the appropriate
// type.
type FuncCall func(val []Value) Value

// Function defines a function that can be executed in tunascript.
type Function struct {
	// Name is the name of the function. It would be called with $Name(). Name
	// is case-insensitive and must follow identifier naming rules ([A-Z0-9_]+)
	Name string

	// RequiredArgs is the number of required arguments. This many arguments is
	// gauratneed to be passed to Call.
	RequiredArgs int

	// OptionalArgs is the number of optional arguments. Up to the
	OptionalArgs int

	// Call is a function point to the golang implementation of the function. It
	// is guaranteed to receive RequiredArgs, and may receive up to OptionalArgs
	// additional args.
	Call FuncCall
}

// Flag is a variable in the engine.
type Flag struct {
	Name string
	Value
}

type Interpreter struct {
	fn    map[string]Function
	flags map[string]*Flag
	world WorldInterface
}

type WorldInterface interface {

	// InInventory returns whether the given label Item is in the player
	// inventory.
	InInventory(label string) bool

	// Move moves the label to the dest. The label can be an NPC or an Item. If
	// label is "@PLAYER", the player will be moved. Returns whether the thing
	// moved.
	Move(label string, dest string) bool

	// Output prints the given string. Returns whether it did successfully.
	Output(s string) bool
}

func NewInterpreter(w WorldInterface) Interpreter {
	inter := Interpreter{
		fn:    make(map[string]Function),
		world: w,
	}

	inter.fn["ADD"] = Function{Name: "ADD", RequiredArgs: 2, Call: builtIn_Add}
	inter.fn["SUB"] = Function{Name: "SUB", RequiredArgs: 2, Call: builtIn_Sub}
	inter.fn["MULT"] = Function{Name: "MULT", RequiredArgs: 2, Call: builtIn_Mult}
	inter.fn["DIV"] = Function{Name: "DIV", RequiredArgs: 2, Call: builtIn_Add}
	inter.fn["OR"] = Function{Name: "OR", RequiredArgs: 2, Call: builtIn_Or}
	inter.fn["AND"] = Function{Name: "AND", RequiredArgs: 2, Call: builtIn_And}
	inter.fn["NOT"] = Function{Name: "NOT", RequiredArgs: 1, Call: builtIn_Not}

	inter.fn["FLAG_ENABLED"] = Function{Name: "FLAG_ENABLED", RequiredArgs: 1, Call: inter.builtIn_FlagEnabled}
	inter.fn["FLAG_DISABLED"] = Function{Name: "FLAG_DISABLED", RequiredArgs: 1, Call: inter.builtIn_FlagDisabled}
	inter.fn["FLAG_IS"] = Function{Name: "FLAG_IS", RequiredArgs: 2, Call: inter.builtIn_FlagIs}
	inter.fn["FLAG_LESS_THAN"] = Function{Name: "FLAG_LESS_THAN", RequiredArgs: 2, Call: inter.builtIn_FlagLessThan}
	inter.fn["FLAG_GREATER_THAN"] = Function{Name: "FLAG_GREATER_THAN", RequiredArgs: 2, Call: inter.builtIn_FlagGreaterThan}
	inter.fn["IN_INVEN"] = Function{Name: "IN_INVEN", RequiredArgs: 1, Call: inter.builtIn_InInven}
	inter.fn["ENABLE"] = Function{Name: "ENABLE", RequiredArgs: 1, Call: inter.builtIn_Enable}
	inter.fn["DISABLE"] = Function{Name: "DISABLE", RequiredArgs: 1, Call: inter.builtIn_Disable}
	inter.fn["INC"] = Function{Name: "INC", RequiredArgs: 1, OptionalArgs: 1, Call: inter.builtIn_Inc}
	inter.fn["DEC"] = Function{Name: "DEC", RequiredArgs: 1, OptionalArgs: 1, Call: inter.builtIn_Dec}
	inter.fn["SET"] = Function{Name: "SET", RequiredArgs: 2, Call: inter.builtIn_Set}
	inter.fn["MOVE"] = Function{Name: "MOVE", RequiredArgs: 2, Call: inter.builtIn_Move}
	inter.fn["OUTPUT"] = Function{Name: "OUTPUT", RequiredArgs: 1, Call: inter.builtIn_Output}

	return inter
}

type symbolType int

const (
	symbolLeftParen symbolType = iota
	symbolRightParen
	symbolValue
	symbolDollar
	symbolIdentifier
	symbolComma
)

type symbol struct {
	sType  symbolType
	source string

	// only applies if sType is symbolValue
	forceType ValueType
}

type ASTNode struct {
	root     *ASTNode
	children []*ASTNode
	sym      symbol
	t        nodeType

	forceType ValueType
}

type lexerState int

const (
	lexDefault lexerState = iota
	lexIdent
	lexStr
)

type nodeType int

const (
	nodeItem nodeType = iota
)

// LexText lexes the text. Returns the AST, whether exiting on right paren, how
// many bytes were consumed, and whether an error was encountered.
func (inter Interpreter) LexText(s string, parent *ASTNode) (*ASTNode, bool, int, error) {
	node := &ASTNode{children: make([]*ASTNode, 0)}
	if parent == nil {
		node.root = node
	} else {
		node.root = parent.root
	}

	escaping := false
	mode := lexDefault
	node.t = nodeItem

	s = strings.TrimSpace(s)
	sRunes := []rune(s)
	sBytes := make([]int, len(sRunes))
	sBytesIdx := 0
	for b := range s {
		sBytes[sBytesIdx] = b
		sBytesIdx++
	}

	var buildingText string
	for i := 0; i < len(sRunes); i++ {
		ch := sRunes[i]

		switch mode {
		case lexIdent:
			if ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z') || ('0' <= ch && ch <= '9') || ch == '_' {
				buildingText += string(ch)
			} else {
				idNode := &ASTNode{
					root: node.root,
					sym:  symbol{sType: symbolIdentifier, source: buildingText},
				}
				node.children = append(node.children, idNode)
				buildingText = ""
				i--
				mode = lexDefault
			}
		case lexStr:
			if !escaping && ch == '\\' {
				// do not add a node for this
				escaping = true
			} else if !escaping && ch == '|' {
				symNode := &ASTNode{
					root: node.root,
					sym:  symbol{sType: symbolValue, source: "|" + buildingText + "|"},
				}
				node.children = append(node.children, symNode)
				buildingText = ""
				mode = lexDefault
			} else {
				buildingText += string(ch)
				escaping = false
			}
		case lexDefault:
			if !escaping && ch == '\\' {
				// do not add a node for this
				escaping = true
			} else if !escaping && ch == '$' {
				if buildingText != "" {
					textNode := &ASTNode{root: node.root, sym: symbol{sType: symbolValue, source: buildingText}}
					node.children = append(node.children, textNode)
					buildingText = ""
				}

				dNode := &ASTNode{
					root: node.root,
					sym:  symbol{sType: symbolDollar, source: "$"},
				}
				node.children = append(node.children, dNode)
				mode = lexIdent
			} else if !escaping && ch == '(' {
				if buildingText != "" {
					textNode := &ASTNode{root: node.root, sym: symbol{sType: symbolValue, source: buildingText}}
					node.children = append(node.children, textNode)
					buildingText = ""
				}

				pNode := &ASTNode{
					root: node.root,
					sym:  symbol{sType: symbolLeftParen, source: "("},
				}
				node.children = append(node.children, pNode)

				// need to find the next byte
				nextByteIdx := -1
				if i+1 < len(sRunes) {
					nextByteIdx = sBytes[i+1]
				}

				// parse the rest as tree, if there's more
				if nextByteIdx != -1 {
					subNode, addRightParen, consumed, err := inter.LexText(s[nextByteIdx:], node)
					if err != nil {
						return nil, false, 0, err
					}
					node.children = append(node.children, subNode)
					if addRightParen {
						node.children = append(node.children, &ASTNode{
							root: node.root,
							sym:  symbol{sType: symbolRightParen, source: ")"},
						})
					}
					i += consumed
				}
			} else if !escaping && ch == ',' {
				if buildingText != "" {
					textNode := &ASTNode{root: node.root, sym: symbol{sType: symbolValue, source: buildingText}}
					node.children = append(node.children, textNode)
					buildingText = ""
				}

				cNode := &ASTNode{
					root: node.root,
					sym:  symbol{sType: symbolComma, source: ","},
				}
				node.children = append(node.children, cNode)
			} else if !escaping && ch == '|' {
				if buildingText != "" {
					textNode := &ASTNode{root: node.root, sym: symbol{sType: symbolValue, source: buildingText}}
					node.children = append(node.children, textNode)
					buildingText = ""
				}

				// string start
				mode = lexStr
			} else if !escaping && ch == ')' {
				// we have reached the end of our parsing. if we are the PARENT,
				// this is an error
				if parent == nil {
					return nil, false, 0, fmt.Errorf("unexpected end of expression (unmatched right-parenthesis)")
				}

				if buildingText != "" {
					textNode := &ASTNode{root: node.root, sym: symbol{sType: symbolValue, source: buildingText}}
					node.children = append(node.children, textNode)
					buildingText = ""
				}

				// don't add it bc parent will
				return node, true, sBytes[i], nil
			} else {
				buildingText += string(ch)
				escaping = false
			}
		}

	}

	// if we get to the end but we are not the parent, we have a paren mismatch
	if parent != nil {
		return nil, false, 0, fmt.Errorf("unexpected end of expression (unmatched left-parenthesis)")
	}

	if buildingText != "" {
		textNode := &ASTNode{root: node.root, sym: symbol{sType: symbolValue, source: buildingText}}
		node.children = append(node.children, textNode)
		buildingText = ""
	}

	// okay now go through and update (for instance, make the values not have double quotes but force to str type)

	return node, false, len(s), nil
}