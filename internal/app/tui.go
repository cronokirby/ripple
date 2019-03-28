package app

import (
	"fmt"
	"log"

	"github.com/cronokirby/ripple/internal/network"
	"github.com/jroimartin/gocui"
)

// gui represents a graphical ui with a swarm handle as well
type gui struct {
	*gocui.Gui
	swarm *network.SwarmHandle
	nick  string
}

func (g *gui) ReceiveContent(user, content string) {
	msg, err := g.View("messages")
	if err != nil {
		return
	}
	fmt.Fprintf(msg, "%s: %s\n", user, content)
}

func wrapSwarm(swarm *network.SwarmHandle, f func(*gui) error) func(*gocui.Gui) error {
	return func(g *gocui.Gui) error {
		wrapped := &gui{g, swarm, ""}
		return f(wrapped)
	}
}

var defaultEditor = gocui.EditorFunc(simpleEditor)

func simpleEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && ch != 10 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
}

func inputView(g *gui, maxX, maxY int) error {
	if v, err := g.SetView("input", 0, 5*maxY/6+2, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		v.Editor = defaultEditor
	}
	sendContent := func(gui *gocui.Gui, v *gocui.View) error {
		content := v.Buffer()
		if content == "" {
			return nil
		}
		// trim the newline
		content = content[:len(content)-1]
		v.Clear()
		if err := v.SetCursor(0, 0); err != nil {
			return err
		}
		var name string
		_, err := fmt.Sscanf(content, "!nick %s", &name)
		if err == nil {
			g.nick = name
			g.swarm.ChangeNickname(name)
		} else {
			g.ReceiveContent("(me) "+g.nick, content)
			g.swarm.SendContent(content)
		}
		return nil
	}
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, sendContent); err != nil {
		return err
	}
	return nil
}

// RunTUI starts the terminal ui for the app
func RunTUI(swarm *network.SwarmHandle) {
	under, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer under.Close()
	g := &gui{Gui: under, swarm: nil}
	swarm.SetReceiver(g)
	g.Cursor = true
	g.SetManagerFunc(wrapSwarm(swarm, layout))
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatal(err)
	}
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatal(err)
	}
}

func layout(g *gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("messages", 0, 0, maxX-1, 5*maxY/6+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "messages"
		v.Autoscroll = true
		v.Wrap = true
	}
	if err := inputView(g, maxX, maxY); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("input"); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
