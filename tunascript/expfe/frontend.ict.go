// Package expfe contains the frontend for analyzing TQTextExpansion
// code. The function [Frontend] is the primary entrypoint for callers, and will
// return a full TQTextExpansion frontend ready for immediate use.
package expfe

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --slr -l TQTextExpansion -v 1.0 -d tte --sim-off --ir github.com/dekarrin/tunaq/tunascript/syntax.Template --hooks ./tunascript/syntax --hooks-table TmplHooksTable --dest ./tunascript/expfe --pkg expfe tunascript/expansion.md
*/

import (
	"fmt"
	"os"
	"strings"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"

	"github.com/dekarrin/tunaq/tunascript/syntax"
)

// FrontendOptions allows options to be set on the compiler frontend returned by
// [Frontend]. It allows setting of debug flags and other optional
// functionality.
type FrontendOptions struct {

	// LexerEager is whether the Lexer should immediately read all input the
	// first time it is called. The default is lazy lexing, where the minimum
	// number of tokens are read when required by the parser.
	LexerEager bool

	// LexerTrace is whether to add tracing functionality to the lexer. This
	// will cause the tokens to be printed to stderr as they are lexed. Note
	// that with LexerEager set, this implies that they will all be lexed and
	// therefore printed before any parsing occurs.
	LexerTrace bool

	// ParserTrace is whether to add tracing functionality to the parser. This
	// will cause parsing events to be printed to stderr as they occur. This
	// includes operations such as token or symbol stack manipulation, and for
	// LR parsers, shifts and reduces.
	ParserTrace bool

	// SDTSTrace is whether to add tracing functionality to the translation
	// scheme. This will cause translation events to be printed to stderr as
	// they occur. This includes operations such as parse tree annotation and
	// hook execution.
	SDTSTrace bool
}

// Frontend returns the complete compiled frontend for the TQTextExpansion langauge.
// The hooks map must be provided as it is the interface between the translation
// scheme in the frontend and the external code executed in the backend. The
// opts parameter allows options to be set on the frontend for debugging and
// other purposes. If opts is nil, it is treated as an empty FrontendOptions.
func Frontend(hooks trans.HookMap, opts *FrontendOptions) ictiobus.Frontend[syntax.Template] {
	if opts == nil {
		opts = &FrontendOptions{}
	}

	fe := ictiobus.Frontend[syntax.Template]{

		Language:    "TQTextExpansion",
		Version:     "1.0",
		IRAttribute: "ast",
		Lexer:       Lexer(!opts.LexerEager),
		Parser:      Parser(),
		SDTS:        SDTS(),
	}

	// Add traces if requested

	if opts.LexerTrace {
		fe.Lexer.RegisterTraceListener(func(t lex.Token) {
			fmt.Fprintf(os.Stderr, "Token: %s\n", t)
		})
	}

	if opts.ParserTrace {
		fe.Parser.RegisterTraceListener(func(s string) {
			fmt.Fprintf(os.Stderr, "Parser: %s\n", s)
		})
	}

	if opts.SDTSTrace {
		fe.SDTS.RegisterListener(func(e trans.Event) {
			switch e.Type {
			case trans.EventAnnotation:
				fmt.Fprintf(os.Stderr, "SDTS: Annotated parse tree:\n%s\n", e.Tree)
			case trans.EventHookCall:
				relNode, ok := e.Hook.Node.RelativeNode(e.Hook.Target.Rel)
				if !ok {
					panic("bad relative node in SDTS hook call event")
				}

				var argSb strings.Builder
				for i := range e.Hook.Args {
					argSb.WriteString(fmt.Sprintf("%#v", e.Hook.Args[i].Value))
					if i+1 < len(e.Hook.Args) {
						argSb.WriteString(", ")
					}
				}

				message := fmt.Sprintf("SDTS: Set (%d: %s).%s = %s(%s) = ", relNode.ID(), relNode.Symbol, e.Hook.Target.Name, e.Hook.Name, argSb.String())
				if e.Hook.Result.Error != nil {
					message += fmt.Sprintf("ERROR(%v) (with value %#v)", e.Hook.Result.Error.Error(), e.Hook.Result.Value)
				} else {
					message += fmt.Sprintf("%#v", e.Hook.Result.Value)
				}

				fmt.Fprintf(os.Stderr, "%s\n", message)

			default:
				fmt.Fprintf(os.Stderr, "SDTS: %v\n", e)
			}
		})
	}

	// Set the hooks
	fe.SDTS.SetHooks(hooks)

	return fe
}
