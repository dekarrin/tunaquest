package expfe

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --slr -l TQTextExpansion -v 1.0 -d tte --sim-off --ir github.com/dekarrin/tunaq/tunascript/syntax.Template --hooks ./tunascript/syntax --hooks-table TmplHooksTable --dest ./tunascript/expfe --pkg expfe tunascript/expansion.md
*/

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"

	"github.com/dekarrin/tunaq/tunascript/expfe/expfetoken"
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
	lx.RegisterClass(expfetoken.TCText, "")
	lx.RegisterClass(expfetoken.TCFlag, "")
	lx.RegisterClass(expfetoken.TCIf, "")
	lx.RegisterClass(expfetoken.TCElseif, "")
	lx.RegisterClass(expfetoken.TCEndif, "")
	lx.RegisterClass(expfetoken.TCElse, "")

	lx.AddPattern(`(?:[^\\\$]|\\.|\$(?:[^A-Za-z0-9_[]|\[[^[]|$))+`, lex.LexAs(expfetoken.TCText.ID()), "", 0)
	lx.AddPattern(`\$[A-Za-z0-9_]+`, lex.LexAs(expfetoken.TCFlag.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ii][Ff](?:\s+(?:[^\\\]]|\][^\]]|\\.)*)?\]\]`, lex.LexAs(expfetoken.TCIf.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Ll](?:[Ss][Ee]\s*)?[Ii][Ff](?:\s+(?:[^\\\]]|\][^\]]|\\.)*)?\]\]`, lex.LexAs(expfetoken.TCElseif.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Nn][Dd]\s*[Ii][Ff]\s*\]\]`, lex.LexAs(expfetoken.TCEndif.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Ll][Ss][Ee]\s*\]\]`, lex.LexAs(expfetoken.TCElse.ID()), "", 0)

	return lx
}
