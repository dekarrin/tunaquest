package fetmpl

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --slr -l TunaQuest Template -v 1.0 -d tte --sim-off --ir github.com/dekarrin/tunaq/tunascript/syntax.Template --hooks ./tunascript/syntax --hooks-table TmplHooksTable --dest ./tunascript/fetmpl --pkg fetmpl tunascript/expansion.md
*/

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"

	"github.com/dekarrin/tunaq/tunascript/fetmpl/fetmpltoken"
)

// Lexer returns the generated ictiobus Lexer for TunaQuest Template.
func Lexer(lazy bool) lex.Lexer {
	var lx lex.Lexer
	if lazy {
		lx = ictiobus.NewLazyLexer()
	} else {
		lx = ictiobus.NewLexer()
	}

	// default state, shared by all
	lx.RegisterClass(fetmpltoken.TCText, "")
	lx.RegisterClass(fetmpltoken.TCFlag, "")
	lx.RegisterClass(fetmpltoken.TCIf, "")
	lx.RegisterClass(fetmpltoken.TCElseif, "")
	lx.RegisterClass(fetmpltoken.TCEndif, "")
	lx.RegisterClass(fetmpltoken.TCElse, "")

	lx.AddPattern(`(?:[^\\\$]|\\.|\$(?:[^A-Za-z0-9_[]|\[[^[]|$))+`, lex.LexAs(fetmpltoken.TCText.ID()), "", 0)
	lx.AddPattern(`\$[A-Za-z0-9_]+`, lex.LexAs(fetmpltoken.TCFlag.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ii][Ff](?:\s+(?:[^\\\]]|\][^\]]|\\.)*)?\]\]`, lex.LexAs(fetmpltoken.TCIf.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Ll](?:[Ss][Ee]\s*)?[Ii][Ff](?:\s+(?:[^\\\]]|\][^\]]|\\.)*)?\]\]`, lex.LexAs(fetmpltoken.TCElseif.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Nn][Dd]\s*[Ii][Ff]\s*\]\]`, lex.LexAs(fetmpltoken.TCEndif.ID()), "", 0)
	lx.AddPattern(`\$\[\[\s*[Ee][Ll][Ss][Ee]\s*\]\]`, lex.LexAs(fetmpltoken.TCElse.ID()), "", 0)

	return lx
}