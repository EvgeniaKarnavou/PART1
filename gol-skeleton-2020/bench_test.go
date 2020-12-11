
package main

import (
"fmt"
"os"
"testing"
"uk.ac.bris.cs/gameoflife/gol"
)
// Benchmark applies the Run to the ship.png b.N times.
// The time taken is carefully measured by go.
// The b.N  repetition is needed because benchmark results are not always constant.
//testing.B defines the number of times we should perform the operation we’re benchmarking
func BenchmarkRun(b *testing.B) {
	os.Stdout= nil
	p := gol.Params{
		Turns:       100,
		ImageWidth:  128,
		ImageHeight: 128,
	}
	for i:=0; i<b.N;i++{
		for threads:=1;threads<=16;threads++{
			p.Threads=threads
			Name := fmt.Sprintf("%dx%dx%d-%d",p.ImageWidth,p.ImageHeight,p.Turns,p.Threads)
			b.Run(fmt.Sprintf(Name), func(b *testing.B) {
				events := make(chan gol.Event,1000)
				gol.Run(p,events,nil)
				for range events{
				}
			})
		}
	}
}
