package blocks

type Blocks struct {
	Message string `json:"message"`
	// Overlay string `json:"overlay"`
	Prompt  string `json:"prompt"`
	Input   string `json:"input"`
	Lines   []Line `json:"lines"`
	ActEntr int    `json:"active entry"`
}

type Line struct {
	Text string `json:"text"`
	// Markup bool   `json:"markup"`
	Icon string `json:"icon"`
	Data string `json:"data"`
}

type BlocksIn struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

type Select struct {
	Id       string
	Action   string
	Selected int
	Message  string
}
