package tunascript

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

	// TODO: add the rest

	return inter
}

func ParseText(s string) {

}
