package game

type Monster struct {
	Charecter
}

func NewRat(x, y int32) *Monster {
	monstr := &Monster{}
	monstr.Pos = Pos{x, y}
	monstr.Symbol = 'R'
	monstr.Name = "Rat"
	monstr.Hitpoints = 5
	monstr.Strength = 5
	monstr.Speed = 5
	return monstr
}
func NewSpider(x, y int32) *Monster {
	monstr := &Monster{}
	monstr.Pos = Pos{x, y}
	monstr.Symbol = 'S'
	monstr.Name = "Spider"
	monstr.Hitpoints = 10
	monstr.Strength = 10
	monstr.Speed = 2
	return monstr
}