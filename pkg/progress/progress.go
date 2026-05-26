package progress

import (
	"fmt"
	"strings"
	"time"
)

type Bar struct {
	total    int64
	current  int64
	width    int
	start    time.Time
	lastDraw time.Time
}

func NewBar(total int64) *Bar {
	return &Bar{
		total:    total,
		width:    50,
		start:    time.Now(),
		lastDraw: time.Now(),
	}
}

func (b *Bar) Update(current int64) {
	b.current = current

	// Only redraw every 100ms to avoid flicker
	if time.Since(b.lastDraw) < 100*time.Millisecond {
		return
	}
	b.lastDraw = time.Now()
	b.draw()
}

func (b *Bar) draw() {
	if b.total == 0 {
		return
	}

	percent := float64(b.current) / float64(b.total)
	filled := int(float64(b.width) * percent)
	empty := b.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	elapsed := time.Since(b.start).Seconds()
	var eta float64
	if b.current > 0 && percent > 0 {
		eta = elapsed/percent - elapsed
	}

	fmt.Printf("\r[%s] %.1f%% (%d/%d) ETA: %.0fs",
		bar, percent*100, b.current, b.total, eta)
}

func (b *Bar) Finish() {
	fmt.Println()
}

// For indeterminate operations (when total size unknown)
type Spinner struct {
	frames   []string
	current  int
	message  string
	stop     chan bool
	lastDraw time.Time
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message:  message,
		stop:     make(chan bool),
		lastDraw: time.Now(),
	}
}

func (s *Spinner) Start() {
	go func() {
		for {
			select {
			case <-s.stop:
				fmt.Print("\r" + strings.Repeat(" ", len(s.message)+10) + "\r")
				return
			default:
				if time.Since(s.lastDraw) > 80*time.Millisecond {
					fmt.Printf("\r%s %s", s.frames[s.current], s.message)
					s.current = (s.current + 1) % len(s.frames)
					s.lastDraw = time.Now()
				}
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.stop <- true
}

func (s *Spinner) StopAndPrint(message string) {
	s.Stop()
	fmt.Printf("\r✅ %s\n", message)
}
