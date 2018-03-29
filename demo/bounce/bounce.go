package main

import (
   "github.com/ansoni/termination"
   _ "time"
   _ "fmt"
   "math/rand" 
   "github.com/nsf/termbox-go"
)
var ballShape = termination.Shape {
    "default": []string { 
      "<-0->",
      "<0-->",
      "<-0->",
      "<--0>",
    },
    "left": []string {
       "|0-->",
    },
    "right": []string {
       "<--0|",
    },
}

var ballMask = map[string][]string {
    "default": []string { 
      "rwgwr",
      "rgwwr",
      "rwgwr",
      "rwwgr",
    },
    "left": []string {
       "ybbby",
    },
    "right": []string {
       "ybbby",
    },
}

/**
 * We don't use the mouse... but we should use it for collision/Death in the future
 */
var mouseShape = map[string][]string {
	"default": []string {
	 "0",
	},
	"mouseLeft": []string {
	"<",
	},
	"mouseRight": []string {
	">",
	},
	"mouseRelease": []string {
	"*",
	},
	"mouseWheelDown": []string {
	"V",
	},
	"mouseWheelUp": []string {
	"^",
	},
}

var ballMask2 = map[string][]string {
    "default": []string { 
      "rrrrr",
      "rrrrr",
      "rrrrr",
      "rrrrr",
    },
    "left": []string {
       "yyyyy",
    },
    "right": []string {
       "yyyyy",
    },
}

func ballCollision(term *termination.Termination, me *termination.Entity, them termination.Entity) {

}


func ballMovement(t *termination.Termination, e *termination.Entity, position termination.Position) termination.Position {
	direction := e.Data.(string)
	if (direction == "right") {
		if ((position.X + e.Width) >= t.Width) {
			e.ShapePath = "right"
			e.Data = "left"
			return position
		}
		e.ShapePath = "default"	
		position.X+=1
		return position
	} else { //left
		if (position.X == 0) {
			e.ShapePath = "left"	
			e.Data = "right"
			return position
		} 
		e.ShapePath = "default"	
		position.X-=1
		return position
	}
}

func mouseMovement(t *termination.Termination, mouse *termination.Entity, position termination.Position) termination.Position {
	if mouse.Data == nil {
		return position
	}
	ev := mouse.Data.(*termbox.Event)
	switch ev.Key {
	case termbox.MouseLeft:
		mouse.ShapePath="mouseLeft"		
	case termbox.MouseMiddle:
		mouse.ShapePath="mouseMiddle"		
	case termbox.MouseRight:
		mouse.ShapePath="mouseRight"		
	case termbox.MouseWheelUp:
		mouse.ShapePath="mouseWheelUp"		
	case termbox.MouseWheelDown:
		mouse.ShapePath="mouseWheelDown"		
	case termbox.MouseRelease:
		mouse.ShapePath="mouseRelease"		
	default:
		mouse.ShapePath="default"		
	}
	position.X=ev.MouseX
	position.Y=ev.MouseY
	return position;
}

func update_mouse(mouse *termination.Entity, ev *termbox.Event) {
	mouse.Data=ev
}

func addBall(term * termination.Termination, position termination.Position, mask1 bool) {
	ball := term.NewEntity(position)
	ball.Shape = ballShape
	if mask1 == true {
        	ball.ColorMask = ballMask
	} else {
  		ball.ColorMask = ballMask2
	}	
	ball.MovementCallback = ballMovement
	rand := random(1,10)
	if rand < 5 {
		ball.Data="right"
	} else {
		ball.Data="left"
	}
}

func random(min int, max int) int {
    return rand.Intn(max-min) + min
}

func main() {
	term := termination.New()
	term.FramesPerSecond = 5
	defer term.Close()

	maskSelection := []bool{true, false}
	for y := 0; y < term.Height;y++ {
		x := random(0,term.Width);
		addBall(term, termination.Position{x,y,0}, maskSelection[random(0,2)])
	}

	go term.Animate()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	mouse := term.NewEntity(termination.Position{-1,-1,0})
	mouse.Shape = mouseShape
	mouse.MovementCallback = mouseMovement


/* user termbox inputs */
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				break mainloop
			}
		case termbox.EventMouse:
			update_mouse(mouse, &ev)
		}
	}
}
