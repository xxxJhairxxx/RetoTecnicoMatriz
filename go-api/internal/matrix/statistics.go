package matrix

type Statistics struct {
	Max         float64         `json:"max"`
	Min         float64         `json:"min"`
	Sum         float64         `json:"sum"`
	Average     float64         `json:"average"`
	Count       int             `json:"count"`
	Diagonal    map[string]bool `json:"diagonal"`
	AnyDiagonal bool            `json:"anyDiagonal"`
}
