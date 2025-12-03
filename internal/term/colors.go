package term

import "fmt"

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

func Success(format string, a ...interface{}) {
	fmt.Printf(Green+"✓ "+format+Reset+"\n", a...)
}

func Info(format string, a ...interface{}) {
	fmt.Printf(Blue+"ℹ "+format+Reset+"\n", a...)
}

func Step(format string, a ...interface{}) {
	fmt.Printf(Bold+"==> "+Reset+format+"\n", a...)
}

func Warn(format string, a ...interface{}) {
	fmt.Printf(Yellow+"⚠ "+format+Reset+"\n", a...)
}

func Error(format string, a ...interface{}) {
	fmt.Printf(Red+"✗ "+format+Reset+"\n", a...)
}

func Print(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

func Section(title string) {
	fmt.Println()
	fmt.Println(Bold + title + Reset)
	fmt.Println(Dim + "───────────────────────────────────────────────────" + Reset)
}

func Banner() {
	fmt.Println("\033[38;5;196m________\033[0m            \033[38;5;202m_____\033[0m")
	fmt.Println("\033[38;5;196m___  __/\033[38;5;208m_____ \033[38;5;214m________\033[38;5;220m  /_\033[0m")
	fmt.Println("\033[38;5;208m__  /_ \033[38;5;214m_  __ `\033[38;5;220m/__  /\033[38;5;226m_  __/\033[0m")
	fmt.Println("\033[38;5;214m_  __/ \033[38;5;220m/ /_/ /\033[38;5;226m__  /_\033[38;5;228m/ /_\033[0m")
	fmt.Println("\033[38;5;220m/_/    \033[38;5;226m\\__,_/ \033[38;5;228m_____/\033[38;5;231m\\__/\033[0m")
}
