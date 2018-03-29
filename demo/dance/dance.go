package main

import (
	"os"
	"github.com/ansoni/termination"
	"github.com/nsf/termbox-go"
)

var kirbyShape = termination.Shape {
    "default": []string {
      "  (>'-')>",
      "  ('-')",
      " ('-')",
      "<('-'<) ",
      " ('-')",
      "  ('-')",
    },
}

type KirbyData struct {
	GotoX int
	GotoY int
}

func kirbyMovement(t *termination.Termination, e *termination.Entity, position termination.Position) termination.Position {
	data, _ := e.Data.(KirbyData)
	if data.GotoX < position.X {
		position.X-=1
	} else if data.GotoX > position.X {
		position.X+=1
	}

	if data.GotoY < position.Y {
		position.Y-=1
	} else if data.GotoY > position.Y {
		position.Y+=1
	}
	return position
}



func main() {
	term := termination.New()
        term.FramesPerSecond = 10
	kirby := term.NewEntity(termination.Position{20,20,0})
        kirby.Shape = kirbyShape
        kirby.MovementCallback = kirbyMovement
	kirby.Data = KirbyData{ GotoX: 20, GotoY: 20 }
	go term.Animate()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	
	// Termbox has great support for inputs
	for {
		switch ev := termbox.PollEvent(); ev.Type {
                case termbox.EventKey:
                        if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
        			term.Close()
				os.Exit(0)
			}
                case termbox.EventMouse:
			data,_ := kirby.Data.(KirbyData)
			data.GotoX = ev.MouseX
			data.GotoY = ev.MouseY
			kirby.Data = data
		}
	}

}
