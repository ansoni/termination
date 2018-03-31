package termination

import (
	_ "container/list"
	"time"
	"fmt"
	"os"
	"log"
	"sort"

	"github.com/nsf/termbox-go"
	rtree "github.com/dhconnelly/rtreego"
)

type Position struct {
	X int
	Y int
	Z int
}

type DeathCallback func(*Termination, *Entity)
type MovementCallback func(*Termination, *Entity, Position) Position
type CollisionCallback func(term *Termination, me *Entity, them Entity, position Position)

type Shape map[string][]string

type Termination struct {
	Debug            string /* set this to file path to output Debug to it */
	Width            int /* RO - the Width of the terminal */
	Height           int /* RO - the Height of the terminal */
	FramesPerSecond  int /* RW - Defaults to 60.  This is the overall Animation speed */
	FrameNum         int /* RO - the current frame number */ 
	TransparencyChar rune /* RW - What character to use for Transparency in our models */ 
	DefaultColor     rune /* RW - The default color for everything */ 

	/* Internal State */
	entities    []*Entity /* All entities added are here */
	_frameStart int64 /* nanosecond snapshot for timing */
	_frameStop  int64 /* nanosecond snapshot for timing */
	_debugFile  *os.File /* our debug output file */
	entityId    int /* incremented counter for handing out ids to entities */
	rt          *rtree.Rtree /* collision detection! */
}

type Entity struct {
	Shape            map[string][]string /* RW - Our Animation Shape.  This contains one or more animation paths */
	ColorMask        map[string][]string /* RW - Our Animation Shape.  Should match the Shape exactly, but instead of doing ascii art, you use characters for colors.  b = blue, c = cyan, etc */
	DefaultColor     rune /* Default Color for the Shape */
	DeathOnLastFrame bool /* Does the Shape die after a single run? */
	DeathOnOffScreen bool /* Does the Shape die when it can't be seen/ */
	Data             interface{} /* RW - This is your state, write what you want here */
	FramesPerSecond  int  /* RW - Speed of the Animation.  This number cannot be greater than the overall Termination::FramesPerSecond */
	MovesPerSecond  int  /* RW - Defaults to Speed of the Animation. This allows you to adjust the movement vs animation*/
	Height           int /* RO - Height of your Shape */
	Width            int /* RO - Width of your Shape */
	ShapePath        string /* RW - Defaults to "default".  This is what path to execute in your shape.  Shape[ShapePath] => []string */

	/* Event Callbacks */
	MovementCallback        MovementCallback /* Tell us how to move the thing */
	CollisionCallback        CollisionCallback /* Tell us what to do when you hit something */
	DeathCallback            DeathCallback /* Tell us how to die */

	/* Internal State */
	position         Position /* where we are */
	_visible bool /* are we visible.  This is mainly used to ensure we don't die if we start off-screen */
	term     *Termination /* Where we live */
	frame    int /* What frame we are on */
	id       int /* our id! */
	bounds   *rtree.Rect /* our cached bounds for collision detection */
}

/** Our Collision Callback.
  * TODO: We should hide this implementation
  */
func (entity *Entity) Bounds() *rtree.Rect {

	return entity.bounds;
}

/* Create a new termination Object that will fill the entire screen */
func New() *Termination {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	term := new(Termination)
	w, h := termbox.Size()
	term.Debug = ""
	term.Height = h
	term.Width = w
	term.entityId = 0
	term.FramesPerSecond = 60
	term.FrameNum = 0
	term.TransparencyChar = '?'
	term.DefaultColor = 'w'
	term.rt = rtree.NewTree(2, 5, 30)
	return term
}

/* Create a new entity.
 * TODO: We have a race condition that allows a half instantiated object to be drawn.
 */
func (term *Termination) NewEntity(position Position) *Entity {
	myEntity := new(Entity)
	myEntity.term = term
	myEntity.ShapePath = "default"
	myEntity.frame = 0 /* We increment before first draw */
	myEntity.position = position
	myEntity.FramesPerSecond = term.FramesPerSecond
	myEntity.MovesPerSecond = -1
	myEntity.Height = 1
	myEntity.Width = 1
	myEntity.DefaultColor = 'w'
	myEntity.DeathOnLastFrame = false
	myEntity.DeathOnOffScreen = false
	myEntity._visible = false
	myEntity.id = term.entityId
	term.entityId += 1

	/* term.entities gets sorted by Position.Z */
	term.entities = append(term.entities, myEntity)
	return myEntity
}

func (term *Termination) Close() {
	termbox.Close()
}

func (term *Termination) frameStart() {
	now := time.Now()
	term._frameStart = now.UnixNano()
	term.FrameNum += 1
	if term.FrameNum > term.FramesPerSecond {
		term.FrameNum = 1
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (term *Termination) debug(format string, a ...interface{}) {
	if term.Debug != "" {
		if term._debugFile == nil {
			f, err := os.Create(term.Debug)
			term._debugFile = f
			check(err)
			log.SetOutput(f)
		}
		output := fmt.Sprintf(format, a...)
		log.Println(output)
	}
}

func (term *Termination) frameStop() {
	now := time.Now()
	term._frameStop = now.UnixNano()
	nanoWait := (int64(time.Second/time.Nanosecond) / int64(term.FramesPerSecond))
	nanoSleepTime := nanoWait - (int64(term._frameStop - term._frameStart))
	time.Sleep(time.Duration(nanoSleepTime) * time.Nanosecond)
}

func colorForRune(char rune) (termbox.Attribute, bool) {
	switch char {
	case '#':
		return termbox.ColorBlack, true
	case 'b':
		return termbox.ColorBlue, true
	case 'B':
		return termbox.ColorBlue | termbox.AttrBold, true
	case 'w':
		return termbox.ColorWhite, true
	case 'W':
		return termbox.ColorWhite | termbox.AttrBold, true
	case 'g':
		return termbox.ColorGreen, true
	case 'G':
		return termbox.ColorGreen | termbox.AttrBold, true
	case 'y':
		return termbox.ColorYellow, true
	case 'Y':
		return termbox.ColorYellow | termbox.AttrBold, true
	case 'm':
		return termbox.ColorMagenta, true
	case 'M':
		return termbox.ColorMagenta | termbox.AttrBold, true
	case 'r':
		return termbox.ColorRed, true
	case 'R':
		return termbox.ColorRed | termbox.AttrBold, true
	case 'c':
		return termbox.ColorCyan, true
	case 'C':
		return termbox.ColorCyan | termbox.AttrBold, true
	}
	return termbox.ColorWhite, false
}

func (term *Termination) sortEntities() {
	sort.Slice(term.entities, func(i, j int) bool {
		return term.entities[i].position.Z > term.entities[j].position.Z
	})
}

func (term *Termination) Animate() {
	for {
		term.frameStart() // <--- helps us with timing
		/* Clear Everything in the buffer */
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

		/* process every entity and draw it! 
		 * these are sorted by zIndex
                 */
		for e := 0; e < len(term.entities); e++ {
			anEntity := term.entities[e]

			/* Execute Movement Functions */
			shouldMove := anEntity.shouldMove()
			shouldStepFrame := anEntity.shouldStepFrame()
			if anEntity.MovementCallback != nil && shouldMove {
				newPosition := anEntity.MovementCallback(term, anEntity, anEntity.position)
				if anEntity.position.Z != newPosition.Z {
					term.sortEntities()
				}
				anEntity.position = newPosition
			}

			if shouldStepFrame {
				/* step our frames */
				pathLen := len(anEntity.Shape[anEntity.ShapePath])
				if anEntity.frame >= (pathLen - 1) {
					if (anEntity.frame != -1) {
						// second loop
						if anEntity.DeathOnLastFrame {
							anEntity.Die()
							continue
						}
					}
					anEntity.frame = 0
				} else {
					anEntity.frame += 1
				}
			}

			/* our frame to draw! */
			drawData := []rune(anEntity.Shape[anEntity.ShapePath][anEntity.frame])

			/* our mask to color it */
			var colorData []rune
			colorDataLength := 0
			if anEntity.ColorMask != nil && len(anEntity.ColorMask[anEntity.ShapePath]) > anEntity.frame {
				term.debug("Has a ColorMask!")
				colorData = []rune(anEntity.ColorMask[anEntity.ShapePath][anEntity.frame])
				colorDataLength = len(colorData)
			}

			/* Figure out the new position */
			originalPosition := anEntity.position
			cursor := anEntity.position
			width := 1
			height := 1
			i := 0
			j := 0

			termDefaultColor, termDefaultColorOk := colorForRune(term.DefaultColor)
			entityDefaultColor, entityDefaultColorOk := colorForRune(anEntity.DefaultColor)
			var defaultColor termbox.Attribute
			if (entityDefaultColorOk) {
				defaultColor = entityDefaultColor
			} else if (termDefaultColorOk) {
				defaultColor = termDefaultColor
			} else {
				panic(fmt.Sprintf("No Default Color defined - Term %v and Entity %v are invalid", termDefaultColor, entityDefaultColor))
			}

			ignoreWhitespace := true
			for _, char := range drawData {
				j+=1
				color := defaultColor

				/* figure out color */
				if colorData != nil {
					term.debug("isColored")
					if colorDataLength > i {
						maskColor, new := colorForRune(colorData[i])
						if (new) {
							color = maskColor
							term.debug("Selected Color: %v", color)
						}
					}
				}
				i += 1

				if (char != term.TransparencyChar) {
					if ignoreWhitespace {
						if char != ' ' {
							ignoreWhitespace = false
						}
					}

					// set the location
					termbox.SetCell(cursor.X, cursor.Y, char, color, termbox.ColorBlack)
				}

				// newlines are... newlines
				if char == '\n' {
					height += 1

					// we only take the largest width
					if (j > width) {
						width = j
					}
					j = 0
					cursor.Y += 1
					cursor.X = originalPosition.X
					ignoreWhitespace = true
					continue
				} else {
					cursor.X += 1
				}
			}

			anEntity.Width = width
			anEntity.Height = height

			if anEntity.DeathOnOffScreen {
				visible := true
				minX := anEntity.position.X
				maxX := anEntity.Width + anEntity.position.X
				minY := anEntity.position.Y
				maxY := anEntity.Height + anEntity.position.Y
				term.debug("[%v,%v]x[%v,%v]\n", minX, minY, maxX, maxY)

				if 0 > minX && 0 > maxX {
					visible = false
				} else if term.Width < minX && term.Width < maxX {
					visible = false
				} else if 0 > minY && 0 > minY {
					visible = false
				} else if term.Height < maxY && term.Height < maxY {
					visible = false
				}
				// we have to be visible first
				if ! anEntity._visible {
					term.debug("Its not visible! %v\n", visible)
					anEntity._visible = visible
				} else {
					//are we off-screen yet?
					if ! visible {
						term.debug("kill it")
						anEntity.Die()
					}
				}
			}

			if anEntity.bounds == nil {
				//initialize bounds
				anEntity.bounds, _ = rtree.NewRect(rtree.Point{float64(anEntity.position.X), float64(anEntity.position.Y)}, []float64{float64(anEntity.Height), float64(anEntity.Width)})

				// we could of added this earlier
				// but we wouldn't have our bounds yet
				term.rt.Insert(anEntity)

			} else {
				//update our bounds
				anEntity.bounds, _ = rtree.NewRect(rtree.Point{float64(anEntity.position.X), float64(anEntity.position.Y)}, []float64{float64(anEntity.Height), float64(anEntity.Width)})
			}

			term.detectCollisions(anEntity)
		}

		/* Flush buffer to screen after we modify everything (thus creating an animation) */
		termbox.Flush()
		term.frameStop()
	}
}

func (term *Termination) detectCollisions(entity *Entity) {
	results := term.rt.SearchIntersect(entity.Bounds())
	if (len(results) > 1) {
		//fmt.Printf("\n\ncollision")
	}
	//fmt.Printf("\n%v",len(results))

}

func (entity *Entity) shouldStepFrame() bool {
	term := entity.term
	if entity.FramesPerSecond != term.FramesPerSecond && entity.FramesPerSecond > 0 {
		updateEveryFrame := int(term.FramesPerSecond / entity.FramesPerSecond)
		if term.FrameNum == 0 {
			panic(fmt.Sprintf("Frame Number should never be 0 - %v/%v=0", term.FramesPerSecond, entity.FramesPerSecond))
		}

		if updateEveryFrame == 0 {
			panic(fmt.Sprintf("Asked to update every 0 frames - %v/%v=0", term.FramesPerSecond, entity.FramesPerSecond))
		}

		if term.FrameNum%updateEveryFrame != 0 {
			term.debug("Should not Move - Terminal.FramesPerSecond: %v, Entity.FramesPerSecond: %v, Frame: %v", term.FramesPerSecond, entity.FramesPerSecond, term.FrameNum)
			return false
		} else {
			term.debug("Should Move - Terminal.FramesPerSecond: %v, Entity.FramesPerSecond: %v, Frame: %v", term.FramesPerSecond, entity.FramesPerSecond, term.FrameNum)
		}
	}
	return true
}

func (entity *Entity) shouldMove() bool {
	term := entity.term
	if entity.FramesPerSecond != term.FramesPerSecond && (entity.FramesPerSecond > 0 || entity.MovesPerSecond > 0) {
		speedBaseline := term.FramesPerSecond
		if entity.MovesPerSecond > 0 {
			speedBaseline = entity.MovesPerSecond
		}
		updateEveryFrame := int(term.FramesPerSecond / speedBaseline)
		if term.FrameNum == 0 {
			panic(fmt.Sprintf("Frame Number should never be 0 - %v/%v=0", term.FramesPerSecond, entity.FramesPerSecond))
		}

		if updateEveryFrame == 0 {
			panic(fmt.Sprintf("Asked to update every 0 frames - %v/%v=0", term.FramesPerSecond, entity.FramesPerSecond))
		}

		if term.FrameNum%updateEveryFrame != 0 {
			term.debug("Should not Move - Terminal.FramesPerSecond: %v, Entity.FramesPerSecond: %v, Frame: %v", term.FramesPerSecond, entity.FramesPerSecond, term.FrameNum)
			return false
		} else {
			term.debug("Should Move - Terminal.FramesPerSecond: %v, Entity.FramesPerSecond: %v, Frame: %v", term.FramesPerSecond, entity.FramesPerSecond, term.FrameNum)
		}
	}
	return true
}

func (entity *Entity) Die() {
	if entity.DeathCallback != nil {
		entity.DeathCallback(entity.term, entity)
	}
	entities := entity.term.entities
	for i := 0; i < len(entities); i++ {
		anEntity := entities[i]
		if entity.id == anEntity.id {
			entity.term.entities = append(entity.term.entities[:i], entities[i+1:]...)
			break
		}
	}
}
