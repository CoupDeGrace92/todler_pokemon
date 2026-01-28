package gamelogic

type RecievedPokemon struct {
	Id      int32  `json:"id"`
	Name    string `json:"name"`
	Types   string `json:"types"`
	Sprites string `json:"sprite_url"`
	Url     string `json:"url"`
	Xp      int32  `json:"base_xp"`
}

type Pokemon struct {
	Id      int
	Name    string
	Types   []string
	Sprites []string
	Url     string
	Xp      int
}
