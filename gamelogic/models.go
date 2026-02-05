package gamelogic

type Config struct {
	JWT       string
	Refresh   string
	ServerUrl string
	Assets    string
}

type RecievedPokemon struct {
	Id      int32  `json:"id"`
	Name    string `json:"name"`
	Types   string `json:"types"`
	Sprites string `json:"sprite_url"`
	Url     string `json:"url"`
	Xp      int32  `json:"base_xp"`
	Count   int32  `json:"count"`
}

type Pokemon struct {
	Id      int
	Name    string
	Types   []string
	Sprites map[string]string
	Url     string
	Xp      int
	Count   int
}

type PokemonPlusCount struct {
	Id      int
	Name    string
	Types   []string
	Sprites map[string]string
	Url     string
	Xp      int
	Count   int
}

var TypeColors = map[string][]int{
	"normal":   {200, 200, 200},
	"fighting": {158, 96, 8},
	"flying":   {114, 187, 232},
	"poison":   {114, 61, 166},
	"ground":   {79, 48, 25},
	"rock":     {64, 50, 41},
	"bug":      {92, 145, 54},
	"ghost":    {32, 2, 36},
	"steel":    {80, 80, 80},
	"fire":     {255, 0, 0},
	"water":    {0, 0, 255},
	"grass":    {0, 255, 0},
	"electric": {222, 218, 0},
	"psychic":  {200, 76, 217},
	"ice":      {76, 217, 210},
	"dragon":   {36, 0, 145},
	"dark":     {23, 13, 5},
	"fairy":    {252, 10, 204},
}
