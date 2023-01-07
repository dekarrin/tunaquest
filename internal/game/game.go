package game

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/dekarrin/rosed"
	"github.com/dekarrin/tunaq/internal/command"
	"github.com/dekarrin/tunaq/internal/tqerrors"
	"github.com/dekarrin/tunaq/internal/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var commandHelp = [][2]string{
	{"HELP", "show this help"},
	{"DROP/PUT", "put down an object in the room"},
	{"DEBUG NPC", "print info on all NPCs, or a single NPC with label LABEL if 'DEBUG NPC LABEL' is typed, or steps all NPCs if 'DEBUG NPC @STEP' is typed."},
	{"DEBUG ROOM", "print info on the current room, or teleport to room with label LABEL if 'DEBUG ROOM LABEL' is typed."},
	{"EXITS", "show the names of all exits from the room"},
	{"GO/MOVE", "go to another room via one of the exits"},
	{"INVENTORY/INVEN", "show your current inventory"},
	{"LOOK", "show the description of the room"},
	{"QUIT/BYE", "end the game"},
	{"TAKE/GET", "pick up an object in the room"},
	{"TALK/SPEAK", "talk to someone/something in the room [WIP]"},
	{"USE", "use an object in your inventory [WIP]"},
}

var textFormatOptions = rosed.Options{
	PreserveParagraphs: true,
}

// State is the game's entire state.
type State struct {
	// World is all rooms that exist and their current state.
	World map[string]*Room

	// CurrentRoom is the room that the player is in.
	CurrentRoom *Room

	// Inventory is the objects that the player currently has.
	Inventory Inventory

	// npcLocations is a map of an NPC's label to the label of the room that the
	// NPC is currently in.
	npcLocations map[string]string

	// width is how wide to make output
	io IODevice
}

type IODevice struct {
	// The width of each line of output.
	Width int

	// a function to send output. If s is empty, an empty line is sent.
	Output func(s string, a ...interface{}) error

	// a function to use to get string input. If prompt is blank, no prompt is
	// sent before the input is read.
	Input func(prompt string) (string, error)

	// a function to use to get int input. If prompt is blank, no prompt is
	// sent before the input is read. If invalid input is received, keeps
	// prompting until a valid one is entered.
	InputInt func(prompt string) (int, error)
}

// New creates a new State and loads the list of rooms into it. It performs
// basic sanity checks to ensure that a valid world is being passed in and
// normalizes them as needed.
//
// startingRoom is the label of the room to start with.
// ioDev is the input/output device to use when the user needs to be prompted
// for more info, or for showing to the user.
// io.Width is how wide the output should be. State will try to make all\
// output fit within this width. If not set or < 2, it will be automatically
// assumed to be 80.
func New(world map[string]*Room, startingRoom string, ioDev IODevice) (State, error) {
	if ioDev.Width < 2 {
		ioDev.Width = 80
	}
	if ioDev.Input == nil {
		return State{}, fmt.Errorf("io device must define an Input function")
	}
	if ioDev.InputInt == nil {
		return State{}, fmt.Errorf("io device must define an InputInt function")
	}
	if ioDev.Output == nil {
		return State{}, fmt.Errorf("io device must define an Output function")
	}

	gs := State{
		World:        world,
		Inventory:    make(Inventory),
		npcLocations: make(map[string]string),
		io:           ioDev,
	}

	// now set the current room
	var startExists bool
	gs.CurrentRoom, startExists = gs.World[startingRoom]
	if !startExists {
		return gs, fmt.Errorf("starting room with label %q does not exist in passed-in rooms", startingRoom)
	}

	// read current npc locations and prep them for movement
	for _, r := range gs.World {
		for _, npc := range r.NPCs {
			npc.ResetRoute()
			gs.npcLocations[npc.Label] = r.Label
		}
	}

	return gs, nil
}

// MoveNPCs applies all movements on NPCs that are in the world.
func (gs *State) MoveNPCs() {
	newLocs := map[string]string{}

	for npcLabel, roomLabel := range gs.npcLocations {
		room := gs.World[roomLabel]
		npc := room.NPCs[npcLabel]

		next := npc.NextRouteStep(room)

		if next != "" {
			nextRoom := gs.World[next]
			nextRoom.NPCs[npc.Label] = npc
			delete(room.NPCs, npc.Label)
			newLocs[npc.Label] = nextRoom.Label
		} else {
			newLocs[npc.Label] = room.Label
		}
	}

	gs.npcLocations = newLocs
}

func (gs *State) RunConversation(npc *NPC) error {
	var output string
	if len(npc.Dialog) < 1 {
		nomPro := cases.Title(language.AmericanEnglish).String(npc.Pronouns.Nominative)
		if err := gs.io.Output("%s doesn't have much to say.", nomPro); err != nil {
			return err
		}
		return nil
	} else {
		for {
			step := npc.Convo.NextStep()
			switch step.Action {
			case DialogEnd:
				npc.Convo = nil
				return nil
			case DialogLine:
				line := step.Content
				resp := step.Response

				ed := rosed.Edit("\""+line+"\"").
					Wrap(gs.io.Width).
					Insert(0, npc.Name+":\n")
				if resp == "" {
					ed = ed.
						Insert(ed.CharCount(), "==>\n")
				} else {
					ed = ed.Insert(ed.CharCount(), "\n\nYOU:\n")
					ed = ed.Insert(ed.CharCount(), rosed.Edit(resp).Wrap(gs.io.Width).String())
					ed = ed.Insert(ed.CharCount(), "==>")
				}
				output = ed.String()

				if err := gs.io.Output(output); err != nil {
					return err
				}
				if _, err := gs.io.Input(""); err != nil {
					return err
				}
			case DialogChoice:
				line := step.Content
				ed := rosed.Edit("\""+line+"\"").
					Wrap(gs.io.Width).
					Insert(0, npc.Name+":\n")
				ed = ed.Insert(ed.CharCount(), "\n\n")
				ed = ed.CharsFrom(ed.CharCount())

				var choiceOut = make([]string, len(step.Choices))
				choiceIdx := 0
				for ch := range step.Choices {
					chDest := step.Choices[ch]
					choiceOut[choiceIdx] = chDest

					ed = ed.Insert(ed.CharCount(), fmt.Sprintf("%d) ", choiceIdx+1))
					ed = ed.Insert(ed.CharCount(), "\n")

					choiceIdx++
				}
				ed = ed.Apply(func(idx int, line string) []string {
					return []string{rosed.Edit(line).Wrap(gs.io.Width).String()}
				})

				var validNum bool
				var choiceNum int
				var err error
				for !validNum {
					choiceNum, err = gs.io.InputInt("> ")
					if err != nil {
						return err
					} else {
						if choiceNum < 1 || len(step.Choices) < choiceNum {
							err = gs.io.Output("Please enter a number between 1 and %d\n", len(step.Choices))
							if err != nil {
								return err
							}
						} else {
							validNum = true
						}
					}
				}

				dest := choiceOut[choiceNum]
				npc.Convo.JumpTo(dest)
			default:
				// should never happen
				panic("unknown line type")
			}
		}
	}
}

// Advance advances the game state based on the given command. If there is a
// problem executing the command, it is given in the error output and the game
// state is not advanced. If it is, the result of the command is written to the
// provided output stream.
//
// Invalid commands will be returned as non-nil errors as opposed to writing
// directly to the IO stream; the caller can decide whether to do this themself.
//
// Note that for this, QUIT is not considered a valid command is it would be on
// a controlling engine to end the game state based on that.
//
// TODO: differentiate syntax errors from io errors
func (gs *State) Advance(cmd command.Command) error {
	var output string

	switch cmd.Verb {
	case "QUIT":
		return tqerrors.Interpreterf("I can't QUIT; I'm not being executed by a quitable engine")
	case "GO":
		egress := gs.CurrentRoom.GetEgressByAlias(cmd.Recipient)
		if egress == nil {
			return tqerrors.Interpreterf("%q isn't a place you can go from here", cmd.Recipient)
		}

		gs.CurrentRoom = gs.World[egress.DestLabel]

		gs.MoveNPCs()

		output += rosed.Edit(egress.TravelMessage).WrapOpts(gs.io.Width, textFormatOptions).String()
	case "EXITS":
		exitTable := ""

		for _, eg := range gs.CurrentRoom.Exits {
			exitTable += strings.Join(eg.Aliases, "/")
			exitTable += " -> "
			exitTable += eg.Description
			exitTable += "\n"
		}

		output = exitTable
	case "TAKE":
		item := gs.CurrentRoom.GetItemByAlias(cmd.Recipient)
		if item == nil {
			return tqerrors.Interpreterf("I don't see any %q here", cmd.Recipient)
		}

		// first remove the item from the room
		gs.CurrentRoom.RemoveItem(item.Label)

		// then add it to inventory.
		gs.Inventory[item.Label] = *item

		output = fmt.Sprintf("You pick up the %s and add it to your inventory.", item.Name)
	case "DROP":
		item := gs.Inventory.GetItemByAlias(cmd.Recipient)
		if item == nil {
			return tqerrors.Interpreterf("You don't have a %q", cmd.Recipient)
		}

		// first remove item from inven
		delete(gs.Inventory, item.Label)

		// add to room
		gs.CurrentRoom.Items = append(gs.CurrentRoom.Items, *item)

		output = fmt.Sprintf("You drop the %s onto the ground", item.Name)
	case "LOOK":
		if cmd.Recipient != "" {
			return tqerrors.Interpreterf("I can't LOOK at particular things yet")
		}

		output = gs.CurrentRoom.Description
		if len(gs.CurrentRoom.Items) > 0 {
			var itemNames []string

			for _, it := range gs.CurrentRoom.Items {
				itemNames = append(itemNames, it.Name)
			}

			output += "\n\n"
			output += "On the ground, you can see "

			output += util.MakeTextList(itemNames) + "."
		}

		output = rosed.Edit(output).WrapOpts(gs.io.Width, textFormatOptions).String()
	case "INVENTORY":
		if len(gs.Inventory) < 1 {
			output = "You aren't carrying anything"
		} else {
			var itemNames []string
			for _, it := range gs.Inventory {
				itemNames = append(itemNames, it.Name)
			}

			output = "You currently have the following items:\n"
			output += util.MakeTextList(itemNames) + "."
		}

		output = rosed.Edit(output).WrapOpts(gs.io.Width, textFormatOptions).String()
	case "TALK":
		roomLabel, ok := gs.npcLocations[cmd.Recipient]
		if !ok {
			return tqerrors.Interpreterf("There doesn't seem to be any NPCs with label %q in this world", cmd.Recipient)
		}

		room := gs.World[roomLabel]
		npc := room.NPCs[cmd.Instrument]

		if npc.Convo == nil {
			npc.Convo = &Conversation{Dialog: npc.Dialog}
		}

		err := gs.RunConversation(npc)
		if err != nil {
			return err
		}

		output = fmt.Sprintf("You stop talking to %s.", strings.ToLower(npc.Pronouns.Objective))
	case "DEBUG":
		if cmd.Recipient == "ROOM" {
			if cmd.Instrument == "" {
				output = gs.CurrentRoom.String() + "\n\n(Type 'DEBUG ROOM label' to teleport to that room)"
			} else {
				if _, ok := gs.World[cmd.Instrument]; !ok {
					return tqerrors.Interpreterf("There doesn't seem to be any rooms with label %q in this world", cmd.Instrument)
				}

				gs.CurrentRoom = gs.World[cmd.Instrument]

				output = fmt.Sprintf("Poof! You are now in %q\n", cmd.Instrument)
			}
		} else if cmd.Recipient == "NPC" {
			if cmd.Instrument == "" {
				// info on all NPCs and their locations
				data := [][]string{{"NPC", "Movement", "Room"}}

				// we need to ensure a consistent ordering so need to sort all
				// keys first
				orderedNPCLabels := make([]string, len(gs.npcLocations))
				var orderedIdx int
				for npcLabel := range gs.npcLocations {
					orderedNPCLabels[orderedIdx] = npcLabel
					orderedIdx++
				}
				sort.Strings(orderedNPCLabels)

				for _, npcLabel := range orderedNPCLabels {
					roomLabel := gs.npcLocations[npcLabel]
					room := gs.World[roomLabel]
					npc := room.NPCs[npcLabel]

					infoRow := []string{npc.Label, npc.Movement.Action.String(), room.Label}
					data = append(data, infoRow)
				}

				footer := "Type \"DEBUG NPC\" followed by the label of an NPC for more info on that NPC.\n"
				footer += "Type \"DEBUG NPC @STEP\" to move all NPCs forward by one turn."

				tableOpts := rosed.Options{
					TableHeaders: true,
				}

				output = rosed.Edit("\n"+footer).
					InsertTableOpts(0, data, 80, tableOpts).
					String()
			} else if strings.HasPrefix(cmd.Instrument, "@") {
				if cmd.Instrument == "@STEP" {
					// check original locations so we can tell how many moved
					originalLocs := make(map[string]string)
					for k, v := range gs.npcLocations {
						originalLocs[k] = v
					}

					gs.MoveNPCs()

					// count how many moved and how many stayed
					var moved, stayed int
					for k := range gs.npcLocations {
						if originalLocs[k] != gs.npcLocations[k] {
							moved++
						} else {
							stayed++
						}
					}

					pluralNPCs := ""
					if stayed+moved != 1 {
						pluralNPCs = "s"
					}

					output = fmt.Sprintf("Applied movement to %d NPC%s; %d moved, %d stayed", stayed+moved, pluralNPCs, moved, stayed)
				} else {
					return tqerrors.Interpreterf("There is no NPC DEBUG action called @%q; you can only use the @STEP action with NPCs")
				}
			} else {
				roomLabel, ok := gs.npcLocations[cmd.Instrument]
				if !ok {
					return tqerrors.Interpreterf("There doesn't seem to be any NPCs with label %q in this world", cmd.Instrument)
				}

				room := gs.World[roomLabel]
				npc := room.NPCs[cmd.Instrument]

				npcInfo := [][2]string{
					{"Name", npc.Name},
					{"Pronouns", npc.Pronouns.Short()},
					{"Room", room.Label},
					{"Movement Type", npc.Movement.Action.String()},
					{"Start Room", npc.Start},
				}

				if npc.Movement.Action == RoutePatrol {
					routeInfo := ""
					for i := range npc.Movement.Path {
						if npc.routeCur != nil && (((*npc.routeCur)+1)%len(npc.Movement.Path) == i) {
							routeInfo += "==> "
						} else {
							routeInfo += "--> "
						}
						routeInfo += npc.Movement.Path[i]
						if i+1 < len(npc.Movement.Path) {
							routeInfo += " "
						}
					}
					npcInfo = append(npcInfo, [2]string{"Route", routeInfo})
				} else if npc.Movement.Action == RouteWander {
					allowed := strings.Join(npc.Movement.AllowedRooms, ", ")
					forbidden := strings.Join(npc.Movement.AllowedRooms, ", ")

					if forbidden == "" {
						if allowed == "" {
							forbidden = "(none)"
						} else {
							forbidden = "(any not in Allowed list)"
						}
					}
					if allowed == "" {
						allowed = "(all)"
					}

					npcInfo = append(npcInfo, [2]string{"Allowed Rooms", allowed})
					npcInfo = append(npcInfo, [2]string{"Forbidden Rooms", forbidden})
				}

				diaStr := "(none defined)"
				if len(npc.Dialog) > 0 {
					node := "step"
					if len(npc.Dialog) != 1 {
						node += "s"
					}
					diaStr = fmt.Sprintf("%d %s in dialog tree", len(npc.Dialog), node)
				}
				npcInfo = append(npcInfo, [2]string{"Dialog", diaStr})

				npcInfo = append(npcInfo, [2]string{"Description", npc.Description})

				// build at width + 2 then eliminate the left margin that
				// InsertDefinitionsTable always adds to remove the 2 extra
				// chars
				tableOpts := rosed.Options{ParagraphSeparator: "\n"}
				output = rosed.Edit("NPC Info for "+npc.Label+"\n"+
					"\n",
				).
					InsertDefinitionsTableOpts(math.MaxInt, npcInfo, gs.io.Width+2, tableOpts).
					LinesFrom(2).
					Apply(func(idx int, line string) []string {
						line = strings.Replace(line[2:], "  -", "  :", 1)
						return []string{line}
					}).
					String()
			}
		} else {
			return tqerrors.Interpreterf("I don't know how to debug %q", cmd.Recipient)
		}
	case "HELP":
		ed := rosed.
			Edit("").
			WithOptions(rosed.Options{ParagraphSeparator: "\n"}).
			InsertDefinitionsTable(0, commandHelp, 80)
		output = ed.
			Insert(0, "Here are the commands you can use (WIP commands do not yet work fully):\n").
			String()
	default:
		return tqerrors.Interpreterf("I don't know how to %q", cmd.Verb)
	}

	// IO to give output:
	return gs.io.Output(output + "\n\n")
}
