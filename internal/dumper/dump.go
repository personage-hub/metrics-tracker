package dumper

type Dumper interface {
	SaveData(gaugeMap map[string]float64, counterMap map[string]int64) error
	RestoreData() (gaugeMap map[string]float64, counterMap map[string]int64, err error)
	CheckHealth() bool
}
