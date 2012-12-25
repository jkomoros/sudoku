package dokugen

type Cell struct {
	grid   *Grid
	Number int
	Row    int
	Col    int
	simpleCellList
}
