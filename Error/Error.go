package Error

type Level uint8
const (
	Debug Level = 0
	Info Level = 1
	Warning Level = 2
	Low Level = 3
	High Level = 4
	Fatal Level = 5
)

type Error struct {
	Lvl Level
	Text string
}

func New(lvl Level, text string) Error {
	return Error{
		Lvl: lvl,
		Text: text,
	}
}

func Err(lvl Level, err error) Error {
	return Error{
		Lvl:  lvl,
		Text: err.Error(),
	}
}