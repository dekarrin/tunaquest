package fe

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --slr -l TunaScript -v 1.0 -d tsi --ir int --hooks ./syntax ./tunascript.md
*/

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/trans"

	"fmt"
	"strings"
)

// SDTS returns the generated ictiobus syntax-directed translation scheme for
// TunaScript.
func SDTS() trans.SDTS {
	sdts := ictiobus.NewSDTS()

	sdtsBindTCTunascript(sdts)

	return sdts
}

func sdtsBindTCTunascript(sdts trans.SDTS) {
	var err error
	err = sdts.Bind(
		"TUNASCRIPT", []string{"EXPR"},
		"value",
		"test_const",
		nil,
	)
	if err != nil {
		prodStr := strings.Join([]string{"EXPR"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "TUNASCRIPT", prodStr, err.Error()))
	}
}
