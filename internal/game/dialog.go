package game

import (
	"fmt"
	"strings"

	"github.com/dekarrin/rosed"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type DialogAction int

const (
	DialogEnd DialogAction = iota
	DialogLine
	DialogChoice
)

func (dst DialogAction) String() string {
	switch dst {
	case DialogEnd:
		return "END"
	case DialogLine:
		return "LINE"
	case DialogChoice:
		return "CHOICE"
	default:
		return fmt.Sprintf("DialogAction(%d)", int(dst))
	}
}

// DialogActionsByString is a map indexing string values to their corresponding
// DialogAction.
var DialogActionsByString map[string]DialogAction = map[string]DialogAction{
	DialogEnd.String():    DialogEnd,
	DialogLine.String():   DialogLine,
	DialogChoice.String(): DialogChoice,
}

// DialogStep is a single step of a Dialog tree. It instructions a Dialog as to
// what should happen in it and how to do it. A step either specifies an end,
// a line, or a choice.
type DialogStep struct {
	Action DialogAction

	// The line of dialog. Not used if action is END.
	Content string

	// How the player responds to the line. Only used if action is LINE, else
	// use choices.
	Response string

	// Choices and the label of dialog step they map to. If one isn't given, it
	// is assumed to end the conversation.
	Choices map[string]string

	// Label of the DialogStep. If not set it will just be the index within the
	// conversation tree.
	Label string
}

// Copy returns a deeply-copied DialogStep.
func (ds DialogStep) Copy() DialogStep {
	dsCopy := DialogStep{
		Action:   ds.Action,
		Label:    ds.Label,
		Response: ds.Response,
		Choices:  make(map[string]string, len(ds.Choices)),
		Content:  ds.Content,
	}

	for k, v := range ds.Choices {
		dsCopy.Choices[k] = v
	}

	return dsCopy
}

func (ds DialogStep) String() string {
	str := fmt.Sprintf("DialogStep<%q", ds.Action)

	switch ds.Action {
	case DialogEnd:
		return str + ">"
	case DialogLine:
		if ds.Label != "" {
			str += fmt.Sprintf(" (%q)", ds.Label)
		}
		return str + fmt.Sprintf(" %q>", ds.Content)
	case DialogChoice:
		str += " ("
		choiceCount := len(ds.Choices)
		gotChoices := 0
		for text, dest := range ds.Choices {
			str += fmt.Sprintf("%q->%q", text, dest)
			gotChoices++
			if gotChoices+1 < choiceCount {
				str += ", "
			}
		}
		str += ")>"
		return str
	default:
		return str + " (UNKNOWN TYPE)>"
	}
}

// Conversation includes the dialog tree and the current position within it. It
// can be created simply by manually creating a Conversation and assigning a
// sequence of steps to Dialog.
type Conversation struct {
	Dialog  []DialogStep
	cur     int
	aliases map[string]int
}

// NextStep gets the next DialogStep in the conversation. If it returns a
// DialogStep with an Action of DialogEnd, the dialog is over and there is
// nothing further to do. If it returns a DialogStep with an Action of
// DialogLine, NextStep can be safely called to go to the next DialogStep after
// that one is processed. If it returns a DialogStep with an Action of
// DialogChoice, JumpTo should be used to jump to the given alias before calling
// NextStep again.
func (convo *Conversation) NextStep() DialogStep {
	if convo.cur >= len(convo.Dialog) {
		return DialogStep{Action: DialogEnd}
	}

	current := convo.Dialog[convo.cur]
	convo.cur++
	return current
}

// JumpTo makes the Conversation JumpTo the step with the given label, so that
// the next call to NextStep returns that DialogStep. If an invalid label is
// given, the next call to NextStep will return an END step. This is a valid
// option for moving the convo to the appropriate position after the user has
// entered a choice.
func (convo *Conversation) JumpTo(label string) {
	if convo.aliases == nil {
		convo.buildAliases()
	}

	pos, ok := convo.aliases[label]

	if !ok {
		convo.buildAliases()
		pos, ok = convo.aliases[label]
	}

	convo.cur = len(convo.Dialog)

	if ok {
		convo.cur = pos
	}
}

func (convo *Conversation) buildAliases() {
	convo.aliases = make(map[string]int)
	for i := range convo.Dialog {
		if convo.Dialog[i].Label == "" {
			convo.Dialog[i].Label = fmt.Sprintf("%d", i)
		}

		convo.aliases[convo.Dialog[i].Label] = i
	}
}

func (gs *State) RunConversation(npc *NPC) error {
	var output string
	if len(npc.Dialog) < 1 {
		nomPro := cases.Title(language.AmericanEnglish).String(npc.Pronouns.Nominative)
		doesNot := "doesn't"
		if npc.Pronouns.Plural {
			doesNot = "don't"
		}
		if err := gs.io.Output("%s %s have much to say.\n", nomPro, doesNot); err != nil {
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

				ed := rosed.Edit("\n"+strings.ToUpper(npc.Name)+":\n").
					CharsFrom(rosed.End).
					Insert(rosed.End, "\""+line+"\"").
					Wrap(gs.io.Width).
					Insert(rosed.End, "\n")
				if resp != "" {
					ed = ed.
						Insert(rosed.End, "\nYOU:\n").
						Insert(rosed.End, rosed.Edit("\""+resp+"\"").Wrap(gs.io.Width).String()).
						Insert(rosed.End, "\n")
				}
				output = ed.String()

				if err := gs.io.Output(output); err != nil {
					return err
				}
				if _, err := gs.io.Input("==> "); err != nil {
					return err
				}
			case DialogChoice:
				line := step.Content
				ed := rosed.Edit("\n"+strings.ToUpper(npc.Name)+":\n").
					CharsFrom(rosed.End).
					Insert(rosed.End, "\""+line+"\"").
					Wrap(gs.io.Width).
					Insert(rosed.End, "\n\n").
					CharsFrom(rosed.End)

				var choiceOut = make([]string, len(step.Choices))
				choiceIdx := 0
				for ch := range step.Choices {
					chDest := step.Choices[ch]
					choiceOut[choiceIdx] = chDest

					ed = ed.Insert(rosed.End, fmt.Sprintf("%d) \"%s\"\n", choiceIdx+1, ch))

					choiceIdx++
				}
				ed = ed.Apply(func(idx int, line string) []string {
					return []string{rosed.Edit(line).Wrap(gs.io.Width).String()}
				})

				var err error

				err = gs.io.Output(ed.String())
				if err != nil {
					return err
				}

				var validNum bool
				var choiceNum int
				for !validNum {
					choiceNum, err = gs.io.InputInt("==> ")
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

				dest := choiceOut[choiceNum-1]
				npc.Convo.JumpTo(dest)
			default:
				// should never happen
				panic("unknown line type")
			}
		}
	}
}
