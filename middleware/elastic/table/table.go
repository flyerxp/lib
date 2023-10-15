package table

type Table struct {
	Name             string
	Version          float32
	IsAllowLocalCurd bool
	Routing          string
	MaxWindowResult  int
	TrackTotalHits   uint64
}
