package gol

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

/*func gameOfLife(p Params, initialWorld [][]byte) [][]byte {

	world := initialWorld

	for turn := 0; turn < p.Turns; turn++ {
		world = calculateNextState(p, world)
	}

	return world
}*/

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	//create a filename channel
	ioFilename := make(chan string)
	ioInput := make(chan uint8)
	ioOutput := make(chan uint8)

	distributorChannels := distributorChannels{
		events,
		ioCommand,
		ioIdle,
		ioFilename,
		ioInput,
		ioOutput,

	}
	go distributor(p, distributorChannels, keyPresses)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: ioFilename,
		output:   ioOutput,
		input:    ioInput,
	}

	//form a filename
	//%dx%d bc images follow this specific format (e.g 16x16 ,512x512) and %d for base10
	//filename := fmt.Sprintf("%dx%d",p.ImageHeight,p.ImageWidth)

	go startIo(p, ioChannels)
}
