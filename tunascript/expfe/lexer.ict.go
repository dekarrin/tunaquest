package fe

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --slr -l TQTextExpansion -v 1.0 -d tte --sim-off --ir github.com/dekarrin/tunaq/tunascript/syntax.ExpansionAST --hooks ./tunascript/syntax --hooks-table ExpHooksTable --dest ./tunascript/expfe tunascript/expansion.md
*/

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"

	"github.com/dekarrin/tunaq/tunascript/expfe/fetoken"
)

// Lexer returns the generated ictiobus Lexer for TQTextExpansion.
func Lexer(lazy bool) lex.Lexer {
	var lx lex.Lexer
	if lazy {
		lx = ictiobus.NewLazyLexer()
	} else {
		lx = ictiobus.NewLexer()
	}

	// default state, shared by all
	lx.RegisterClass(fetoken.TCText, "")
	lx.RegisterClass(fetoken.TCFlag, "")
	lx.RegisterClass(fetoken.TCIf, "")
	lx.RegisterClass(fetoken.TCElseif, "")
	lx.RegisterClass(fetoken.TCEndif, "")
	lx.RegisterClass(fetoken.TCElse, "")

	lx.AddPattern(`(?:[^\\\$]|\\.)+`, lex.LexAs(fetoken.TCText.ID()), "", 0)
	lx.AddPattern(`\$[A-Za-z0-9_]+`, lex.LexAs(fetoken.TCFlag.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ii][Ff](?:\s+(?:[^\\\]]|\][^\]]|\\.)*)?\]\]`, lex.LexAs(fetoken.TCIf.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Ll](?:[Ss][Ee]\s*)?[Ii][Ff](?:\s+(?:[^\\\]]|\][^\]]|\\.)*)?\]\]`, lex.LexAs(fetoken.TCElseif.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Nn][Dd]\s*[Ii][Ff]\s*\]\]`, lex.LexAs(fetoken.TCEndif.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Ll][Ss][Ee]\s*\]\]`, lex.LexAs(fetoken.TCElse.ID()), "", 0)

	return lx
}