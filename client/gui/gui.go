package gui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"strconv"
	"time"

	gl "coupdegrace92/pokemon_for_todlers/gamelogic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Vector struct {
	X float64
	Y float64
}

type Core struct {
	PokeDex    map[string]*gl.PokemonPlusCount
	NewPokemon map[string]int
}

func (c *Core) Reset(cfg gl.Config, user string) {
	cfg.Reset(user)
	c.PokeDex = make(map[string]*gl.PokemonPlusCount)
	c.NewPokemon = make(map[string]int)
}

func (c *Core) Catch(cfg gl.Config, name string) (gl.Pokemon, error) {
	poke, err := cfg.GetPokemon(name)
	if err != nil {
		fmt.Println(err)
		return gl.Pokemon{}, err
	}
	gl.AddToMap(poke, c.PokeDex)
	gl.AddToStringMap(poke, c.NewPokemon)
	return poke, nil
}

func (c *Core) CatchRandom(cfg gl.Config) (gl.Pokemon, error) {
	pokeID := gl.CatchRandom()
	strID := strconv.Itoa(pokeID)
	poke, err := cfg.GetPokemon(strID)
	if err != nil {
		return gl.Pokemon{}, err
	}
	gl.AddToMap(poke, c.PokeDex)
	gl.AddToStringMap(poke, c.NewPokemon)
	return poke, nil
}

func (c *Core) UpdateTeam(cfg gl.Config) error {
	err := cfg.UpdateTeam(c.NewPokemon)
	if err != nil {
		fmt.Println("Error saving team: ", err)
		return err
	}
	return nil
}

type PokeLocation struct {
	Positions []Vector
	Sprite    []*ebiten.Image
}

func LoadFontFace(path string, size float64) (text.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	src, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	//Method Measure has a pointer reciever so we need the tface to to be a pointer to a text-face like object

	tface := &text.GoTextFace{
		Source: src,
		Size:   size,
	}
	return tface, nil
}

type Timer struct {
	CurrentTicks int
	TargetTicks  int
}

func NewTimer(d time.Duration) *Timer {
	return &Timer{
		CurrentTicks: 0,
		TargetTicks:  int(d.Milliseconds()) * ebiten.TPS() / 1000,
	}
}

func (t *Timer) IsReady() bool {
	return t.CurrentTicks >= t.TargetTicks
}

func (t *Timer) Reset() {
	t.CurrentTicks = 0
}

type Game struct {
	Pokemon     []PokeLocation
	C           *Core
	Done        bool
	Cfg         *gl.Config
	CommandLine []rune
	Commands    []string
	TitleFont   text.Face
	CommandFont text.Face
	User        string
}

func (g *Game) Update() error {

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.Done = true
		err := g.C.UpdateTeam(*g.Cfg)
		if err != nil {
			fmt.Println("Error updating team to server: ", err)
		}
	}
	//We only want alpha runes:
	tmp := []rune{}
	tmp = ebiten.AppendInputChars(tmp)
	for _, r := range tmp {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			g.CommandLine = append(g.CommandLine, r)
		}
		if r == ' ' {
			g.CommandLine = append(g.CommandLine, r)
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		command := string(g.CommandLine)
		g.CommandLine = []rune{}
		fmt.Println(command)
	}
	if g.Done {
		return ebiten.Termination
	}
	return nil
}

func init() {

}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.TitleFont != nil {

		w := screen.Bounds().Dx()
		h := screen.Bounds().Dy()
		t := fmt.Sprintf("%s's Pokemon Adventure", g.User)

		width, _ := text.Measure(t, g.TitleFont, 0)

		x := int(w/2) - int(width/2)
		y := int(h / 15)

		var options text.DrawOptions
		options.GeoM.Translate(float64(x), float64(y))
		options.ColorScale.ScaleWithColor(color.White)

		text.Draw(screen, t, g.TitleFont, &options)
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func LoadNonEmbedded(name string) (*ebiten.Image, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}
