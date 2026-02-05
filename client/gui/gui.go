package gui

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func (v *Vector) Add(w Vector) {
	v.X += w.X
	v.Y += w.Y
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
	Position Vector
	Sprite   *ebiten.Image
	Target   Vector
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

const targetSpriteScreenFraction = 0.25

type Game struct {
	Pokemon      []PokeLocation
	C            *Core
	Done         bool
	Cfg          *gl.Config
	CommandLine  []rune
	Commands     map[string]func(g *Game, args ...string) error
	TitleFont    text.Face
	CommandFont  text.Face
	User         string
	Screen       *ebiten.Image
	NextOpenSpot Vector
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
		if r == ' ' || r == '_' || r == '-' {
			g.CommandLine = append(g.CommandLine, r)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		command := string(g.CommandLine)
		fmt.Println("Executing: ", command)
		g.CommandLine = []rune{}
		fmt.Println(command)

		cmd := strings.Fields(command)
		if len(cmd) == 0 {
			return nil
		}
		f, ok := g.Commands[cmd[0]]
		if !ok {
			return nil
		}
		err := f(g, cmd[1:]...)
		if err != nil {
			fmt.Println("Error executing %s: %v\n", cmd[0], err)
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		l := len(g.CommandLine)
		if l == 0 {
			return nil
		}
		g.CommandLine = g.CommandLine[:l-1]
	}

	if g.Done {
		return ebiten.Termination
	}

	for i := range g.Pokemon {
		speed := 10.0
		poke := &g.Pokemon[i]
		if poke.Position != poke.Target {
			dx, dy := gl.MoveToTarget(poke.Target.X, poke.Target.Y, poke.Position.X, poke.Position.Y, speed)
			move := Vector{
				X: dx,
				Y: dy,
			}
			poke.Position.Add(move)
		}
	}

	return nil
}

func init() {

}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Screen = screen
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
	if g.CommandFont != nil {
		w := screen.Bounds().Dx()
		h := screen.Bounds().Dy()

		t := fmt.Sprintf("COMMAND:  %v", string(g.CommandLine))

		x := int(w / 15)
		y := int(14 * h / 15)

		var options text.DrawOptions
		options.GeoM.Translate(float64(x), float64(y))
		options.ColorScale.ScaleWithColor(color.White)

		text.Draw(screen, t, g.CommandFont, &options)
	}

	sh := float64(screen.Bounds().Dy())

	for i, loc := range g.Pokemon {
		if loc.Sprite == nil {
			fmt.Println("Draw: nil sprite at index", i)
			continue
		}

		ph := float64(loc.Sprite.Bounds().Dy())
		if ph == 0 {
			fmt.Println("Draw: zero-height sprite at index", i)
			continue
		}

		spriteScale := (sh * targetSpriteScreenFraction) / ph

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(spriteScale, spriteScale)
		op.GeoM.Translate(loc.Position.X, loc.Position.Y)

		screen.DrawImage(loc.Sprite, op)
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

func Quit(g *Game, args ...string) error {
	g.Done = true
	fmt.Println("Exiting gui")
	return nil
}

func Catch(g *Game, args ...string) error {
	if len(args) == 0 {
		return nil
	}
	if args[0] == "random" || args[0] == "r" {
		poke, err := g.C.CatchRandom(*g.Cfg)
		if err != nil {
			fmt.Println("Error catching pokemon: ", err)
			return err
		}
		fmt.Println("Caught: ", poke.Name)
		err = g.drawCaughtPokemon(poke)
		return err
	}
	poke, err := g.C.Catch(*g.Cfg, strings.ToLower(args[0]))
	if err != nil {
		fmt.Printf("Error catching %s: %v\n", args[0], err)
		return err
	}
	fmt.Println("Caught ", poke.Name)
	err = g.drawCaughtPokemon(poke)
	return err
}

func (g *Game) drawCaughtPokemon(poke gl.Pokemon) error {
	sprite, err := ensureSprite(poke)
	if err != nil {
		fmt.Println("Error fetching sprite: ", err)
		return err
	}
	p := ebiten.NewImageFromImage(sprite)

	sw := float64(g.Screen.Bounds().Dx())
	sh := float64(g.Screen.Bounds().Dy())

	pw := float64(p.Bounds().Dx())
	ph := float64(p.Bounds().Dy())

	// position BEFORE scaling; treat Position as top-left of the *unscaled* sprite
	x := sw/2 - pw/2 // center horizontally
	y := sh - ph     // touch bottom
	spriteScale := (sh * targetSpriteScreenFraction) / ph
	wScaled := pw * spriteScale

	if g.NextOpenSpot.X >= float64(g.Screen.Bounds().Dx())-wScaled {
		g.ScrollRight(g.NextOpenSpot.X - float64(g.Screen.Bounds().Dx()) + wScaled)
	}

	loc := PokeLocation{
		Sprite:   p,
		Position: Vector{X: x, Y: y},
		Target:   g.NextOpenSpot,
	}

	g.NextOpenSpot.X = g.NextOpenSpot.X + wScaled + .005*sw
	g.Pokemon = append(g.Pokemon, loc)
	return nil
}

func ensureSprite(poke gl.Pokemon) (image.Image, error) {
	file := "assets/sprites/" + poke.Name + "_" + "front" + ".png"
	_, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			sprite, _, err := GetSprite(poke)
			if err != nil {
				fmt.Println("Error getting sprite: ", err)
				return nil, err
			}
			f, err := os.Create(file)
			if err != nil {
				fmt.Printf("Error creating file: ", err)
				return nil, err
			}
			defer f.Close()
			err = png.Encode(f, sprite)
			if err != nil {
				fmt.Println("Error encoding sprite to png: ", err)
				return nil, err
			}
			return sprite, nil
		} else {
			fmt.Println("Error retrieving file from path: ", err)
			return nil, err
		}
	}
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return nil, err
	}
	defer f.Close()

	sprite, _, err := image.Decode(f)
	if err != nil {
		fmt.Println("Error decoding sprite")
		return nil, err
	}
	return sprite, nil
}

func GetSprite(poke gl.Pokemon) (img image.Image, format string, err error) {
	var url string
	url = poke.Sprites["front"]

	client := http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating http request: %v\n", err)
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error getting response from http request: ", err)
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Response status not ok: ", resp.Status)
		err = fmt.Errorf("Response status not ok")
		return nil, "", err
	}

	outImg, format, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Println("Error decoding image from response: ", err)
		return nil, "", err
	}
	return outImg, format, nil
}

func (g *Game) ScrollLeft(distance float64) {
	speedX := Vector{
		X: distance,
		Y: 0,
	}
	g.NextOpenSpot.Add(speedX)
	for i := range g.Pokemon {
		p := &g.Pokemon[i]
		p.Target.Add(speedX)
	}
}

func (g *Game) ScrollRight(distance float64) {
	speedX := Vector{
		X: -1 * distance,
		Y: 0,
	}
	g.NextOpenSpot.Add(speedX)
	for i := range g.Pokemon {
		p := &g.Pokemon[i]
		p.Target.Add(speedX)
	}
}
