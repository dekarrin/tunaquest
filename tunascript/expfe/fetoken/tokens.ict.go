// Package fetoken contains the token classes used by the frontend
// of TQTextExpansion. It is in a separate package so that it can be imported and
// used by external packages while still allowing those external packages to be
// imported by the rest of the frontend.
package fetoken

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --slr -l TQTextExpansion -v 1.0 -d tte --sim-off --ir github.com/dekarrin/tunaq/tunascript/syntax.ExpansionAST --hooks ./tunascript/syntax --hooks-table ExpHooksTable --dest ./tunascript/expfe tunascript/expansion.md
*/

import (
	"github.com/dekarrin/ictiobus/lex"
)

var (
	// TCElse is the token class representing an else in TQTextExpansion.
	TCElse = lex.NewTokenClass("else", "else")

	// TCElseif is the token class representing an elseif in TQTextExpansion.
	TCElseif = lex.NewTokenClass("elseif", "elseif")

	// TCEndif is the token class representing an endif in TQTextExpansion.
	TCEndif = lex.NewTokenClass("endif", "endif")

	// TCFlag is the token class representing a flag in TQTextExpansion.
	TCFlag = lex.NewTokenClass("flag", "flag")

	// TCIf is the token class representing an if in TQTextExpansion.
	TCIf = lex.NewTokenClass("if", "if")

	// TCText is the token class representing a text in TQTextExpansion.
	TCText = lex.NewTokenClass("text", "text")
)

var all = map[string]lex.TokenClass{
	"else":   TCElse,
	"elseif": TCElseif,
	"endif":  TCEndif,
	"flag":   TCFlag,
	"if":     TCIf,
	"text":   TCText,
}

// ByID returns the TokenClass in TQTextExpansion that has the given ID. If no token
// class with that ID exists, nil is returned.
func ByID(id string) lex.TokenClass {
	return all[id]
}