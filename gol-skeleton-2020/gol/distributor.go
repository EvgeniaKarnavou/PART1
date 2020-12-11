package gol

import (
	"fmt"
	"time"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events    chan<- Event
	ioCommand chan<- ioCommand
	ioIdle    <-chan bool
	ioFilename chan <- string
	ioInput <- chan uint8
	ioOutput chan<- uint8
}
const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}


func calculateNeighbours(p Params, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}


func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	aliveCells := []util.Cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	return aliveCells
}

func CellsCount(p Params, world [][]byte) int {
	count := 0
	for Y := 0; Y < p.ImageHeight; Y++ {
		for X := 0; X < p.ImageWidth; X++ {
			if world[Y][X] == alive {
				count++
			}
		}
	}
	return count
}

func calculateNextState(start int, end int, p Params, world [][]byte,events chan <- Event, turn int) [][]byte {
	val := end- start
	newWorld := make([][]byte, val )
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	i:=0
	for y := start; y < end; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := calculateNeighbours(p, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[i][x] = alive
				} else {
					newWorld[i][x] = dead
					events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				}
			} else {
				if neighbours == 3 {
					newWorld[i][x] = alive
					events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				} else {
					newWorld[i][x] = dead
				}
			}
		}
		i++
	}
	return newWorld
}

func worker(start int, end int, world [][]byte, p Params, events chan <- Event, turn int,out chan<- [][]byte){
	//execute the calculateNextState function
	worldPart :=calculateNextState(start, end, p, world, events, turn)
	//send the resulting slice back to the out channel
	out<- worldPart
}


func keyPress(c distributorChannels, p Params, world[][]byte){
	c.ioCommand <- ioOutput
	filename := fmt.Sprintf("%dx%dx%d.pgm", p.ImageWidth, p.ImageHeight, p.Turns)
	c.ioFilename <- filename
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- world[y][x]
		}
	}
}


// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels,keyPresses <-chan rune) {
	// TODO: Create a 2D slice to store the world.
	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	// Request the io goroutine to read in the image with the given filename.

	c.ioCommand <- ioInput
	filename := fmt.Sprintf("%dx%d",p.ImageHeight,p.ImageWidth)
	c.ioFilename <- filename

	// The io goroutine sends the requested image byte by byte, in rows.
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			val := <-c.ioInput
			if val != 0 {
				newWorld[y][x] = val
			}
		}
	}
	// TODO: For all initially alive cells send a CellFlipped Event.

	AliveCells := []util.Cell{}
	AliveCells = calculateAliveCells(p, newWorld)
	for _,cell:= range AliveCells{
		c.events<- CellFlipped{0, cell}
		//fmt.Println(p.Turns)
	}



	turn:=0
	//AiveCells := CellsCount(p, newWorld)
	//fmt.Println(AiveCells)


	workerHeight := p.ImageHeight/p.Threads
	remaining:=p.ImageHeight%p.Threads
	//create a channel to store the slices of the world for each worker
	out:= make ([]chan [][]byte,p.Threads)
	for i:= range out{
		out[i]= make(chan [][]byte)
	}

	ticker := time.NewTicker(2 * time.Second)

	for turn := 1; turn <= p.Turns; turn++ {

		if p.Threads > 1 {


			for i := 0; i < p.Threads; i++ {
				//t := turn + 1
				//if the number of threads doesnt divide well with the image, then
				//if the current thread is the last one, give it the remaining lines to calculate
				if (remaining > 0) && ((i + 1) == p.Threads) {
					go worker(i*workerHeight, ((i+1)*workerHeight)+remaining, newWorld, p, c.events, turn, out[i])
				} else { //else, just give each thread the appointed workerHeight
					go worker(i*workerHeight, (i+1)*workerHeight, newWorld, p, c.events, turn, out[i])
				}
			}
			//make a temprorary world to append all the slices to.
			tempOut := make([][]byte, 0)
			for i := 0; i < p.Threads; i++ {
				part := <-out[i]
				tempOut = append(tempOut, part...)
			}

			for y := 0; y < p.ImageHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					// Replace placeholder tempOut[y][x] with the real newWorld[y][x]
					newWorld[y][x] = tempOut[y][x]
				}
			}

			c.events <- TurnComplete{CompletedTurns: turn}

		}else{
			//if the number of worker threads is one,
			//give the lone worker the whole image to calculate

			start:=0
			end:=p.ImageHeight

			newWorld= calculateNextState(start, end, p, newWorld, c.events, turn)
			c.events <- TurnComplete{CompletedTurns: turn}
		}
		select {
		case k:= <- keyPresses:
			if k == 's' {
				keyPress(c,p,newWorld)

			}else if k == 'q'{
				keyPress(c,p,newWorld)
				c.events <- StateChange{turn, Quitting}
				return

			}else if k == 'p'{
				fmt.Println(turn)
				c.events<-StateChange{CompletedTurns: turn, NewState:Paused}
				for {
					tempKey := <-keyPresses
					if tempKey == 'p' {
						c.events<-StateChange{CompletedTurns: turn, NewState: Executing}
						break
					}
				}
			}


		case <-ticker.C:
			AliveCells := CellsCount(p, newWorld)
			fmt.Println(AliveCells)
			c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: AliveCells}
		default:
		}
	}




	c.ioCommand <- ioOutput
	filename1 := fmt.Sprintf("%dx%dx%d", p.ImageWidth, p.ImageHeight, p.Turns)
	c.ioFilename <- filename1
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			//out := newWorld[y][x]
			c.ioOutput <- newWorld[y][x]
		}
	}
	//var final []util.Cell
	final := calculateAliveCells(p, newWorld)
	c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: final}

	c.events <- ImageOutputComplete{CompletedTurns: p.Turns, Filename: filename}
	c.ioCommand <- ioCheckIdle
	<- c.ioIdle
	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}