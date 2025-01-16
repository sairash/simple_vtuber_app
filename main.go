package main

import (
	"context"
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/urfave/cli/v3"
)

type Game struct {
	Image         *ebiten.Image
	PivotX        float64
	PivotY        float64
	Amplitude     float64
	AmplitudeChan chan float64
}

func StartAudioStream(amplitudeChan chan float64) {
	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, 64, func(in []float32) {
		var sum float64
		for _, v := range in {
			sum += math.Abs(float64(v))
		}
		averageAmplitude := sum / float64(len(in))
		amplitudeChan <- averageAmplitude
	})
	if err != nil {
		log.Fatalf("Error opening stream: %v", err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Fatalf("Error starting stream: %v", err)
	}

	for {
		time.Sleep(time.Millisecond * 100)
	}
}

func (g *Game) Update() error {
	select {
	case amplitude := <-g.AmplitudeChan:

		if amplitude > 0.05 {
			amplitude = 0.05
		} else if amplitude <= 0.025 {
			amplitude = 0
		}
		g.Amplitude = amplitude
	default:
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0xff, 0, 0xff})

	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(-g.PivotX, -g.PivotY)

	op.GeoM.Scale(0.5+g.Amplitude, 0.5+g.Amplitude)

	op.GeoM.Translate(210, 500)

	screen.DrawImage(g.Image, op)
	// ebitenutil.DebugPrint(screen, fmt.Sprintf("%f", g.Amplitude))

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 400, 500
}

func run_program(image_path string) {
	img, _, err := ebitenutil.NewImageFromFile(image_path)
	if err != nil {
		log.Fatalf("Error loading image: %v", err)
	}

	pivot := img.Bounds().Size()

	amplitudeChan := make(chan float64)

	go StartAudioStream(amplitudeChan)

	ebiten.SetWindowSize(400, 500)
	ebiten.SetWindowTitle("Vocapp")

	game := &Game{
		Image:         img,
		AmplitudeChan: amplitudeChan,
		PivotX:        float64(pivot.X) / 2,
		PivotY:        float64(pivot.Y),
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "character",
				Aliases: []string{"c"},
				Value:   "./leohead.png",
				Usage:   "add a character image path",

				Action: func(ctx context.Context, c *cli.Command, s string) error {
					c.HideHelp = true
					run_program(s)
					return nil
				},
			},
		},
		Usage:       "You Speak, Character go boom boom.",
		Description: "Simple Vtubing App where the character moves when you speak.",
		Authors:     []any{"Sairash Sharma Gautam"},
		Action: func(ctx context.Context, c *cli.Command) error {
			if !c.HideHelp {
				cli.ShowAppHelp(c)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

}
