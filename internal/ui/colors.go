package ui

import "github.com/fatih/color"

// Theme colors based on mushak brand palette
var (
	// Primary brand colors (blues)
	Brand1 = color.New(color.FgHiCyan).SprintFunc()    // #48cae4 - light cyan
	Brand2 = color.New(color.FgCyan).SprintFunc()      // #0077b6 - cyan blue
	Brand3 = color.New(color.FgBlue).SprintFunc()      // #023e8a - dark blue

	// Semantic colors
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Warning = color.New(color.FgYellow).SprintFunc()
	Info    = color.New(color.FgCyan).SprintFunc()
	Muted   = color.New(color.FgHiBlack).SprintFunc()

	// Text styles
	Bold      = color.New(color.Bold).SprintFunc()
	BoldCyan  = color.New(color.Bold, color.FgCyan).SprintFunc()
	BoldGreen = color.New(color.Bold, color.FgGreen).SprintFunc()
	BoldRed   = color.New(color.Bold, color.FgRed).SprintFunc()
)

// ASCII art for mushak with gradient effect
const ASCIIArt = `
    ███╗   ███╗██╗   ██╗███████╗██╗  ██╗ █████╗ ██╗  ██╗
    ████╗ ████║██║   ██║██╔════╝██║  ██║██╔══██╗██║ ██╔╝
    ██╔████╔██║██║   ██║███████╗███████║███████║█████╔╝
    ██║╚██╔╝██║██║   ██║╚════██║██╔══██║██╔══██║██╔═██╗
    ██║ ╚═╝ ██║╚██████╔╝███████║██║  ██║██║  ██║██║  ██╗
    ╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝
`

// PrintBanner prints the mushak ASCII art banner in brand colors
func PrintBanner() {
	cyan := color.New(color.FgCyan)
	cyan.Println(ASCIIArt)
	color.New(color.FgHiBlack).Println("    Zero-config, zero-downtime deployments to your Linux server")
	println()
}

// PrintSuccess prints a success message with checkmark
func PrintSuccess(message string) {
	color.New(color.FgGreen).Printf("✓ %s\n", message)
}

// PrintError prints an error message
func PrintError(message string) {
	color.New(color.FgRed).Printf("✗ %s\n", message)
}

// PrintInfo prints an info message with arrow
func PrintInfo(message string) {
	color.New(color.FgCyan).Printf("→ %s\n", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	color.New(color.FgYellow).Printf("⚠ %s\n", message)
}

// PrintHeader prints a section header
func PrintHeader(message string) {
	println()
	color.New(color.Bold, color.FgCyan).Println(message)
	color.New(color.FgHiBlack).Println("────────────────────────────────────────")
}

// PrintKeyValue prints a key-value pair
func PrintKeyValue(key, value string) {
	color.New(color.FgHiBlack).Printf("  %s: ", key)
	color.New(color.FgCyan).Println(value)
}

// PrintSeparator prints a visual separator
func PrintSeparator() {
	color.New(color.FgHiBlack).Println("════════════════════════════════════════")
}

// PrintBox prints a message in a box
func PrintBox(lines []string) {
	PrintSeparator()
	for _, line := range lines {
		color.New(color.FgCyan).Println(line)
	}
	PrintSeparator()
}
