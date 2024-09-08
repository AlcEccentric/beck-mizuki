package param

import "fmt"

type ExecutionMode int

const (
	ColdStartMode ExecutionMode = iota
	RegularUpdateMode
	DirectlyExitMode
)

func CrawlerModeFromString(modeStr string) (mode ExecutionMode, err error) {
	switch modeStr {
	case "cs":
		return ColdStartMode, nil
	case "regular":
		return RegularUpdateMode, nil
	case "exit":
		return DirectlyExitMode, nil
	default:
		return -1, fmt.Errorf("mode %s is not supported", modeStr)
	}
}

func (mode ExecutionMode) String() string {
	switch mode {
	case ColdStartMode:
		return "cs"
	case RegularUpdateMode:
		return "regular"
	case DirectlyExitMode:
		return "exit"
	default:
		return ""
	}
}
