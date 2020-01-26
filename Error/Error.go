package Error

type Level uint8

const (
	Debug   Level = 0
	Info    Level = 1
	Warning Level = 2
	Low     Level = 3
	High    Level = 4
	Fatal   Level = 5
)

type Error struct {
	Lvl  Level
	Text string
}

func New(lvl Level, text string, flag ...string) *Error {
	if len(flag) > 0 {
		var allFlags string
		for _, flagStr := range flag {
			allFlags += flagStr + " "
		}
		text = allFlags + text
	}
	return &Error{
		Lvl:  lvl,
		Text: text,
	}
}

func Err(lvl Level, err error, flag ...string) *Error {
	var text string
	for _, flagStr := range flag {
		text += flagStr + " "
	}
	text += err.Error()
	return &Error{
		Lvl:  lvl,
		Text: text,
	}
}
