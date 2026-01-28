package utils

import (
	"fmt"
)

func PrintWelcome(module string) {
	// Clear screen
	fmt.Print("\033[2J")
	fmt.Print("\033[H")

	fmt.Println("\033[1;36m") // blue
	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║                                                       ║")
	fmt.Println("║   	    █████╗ ██╗   ██╗██████╗ ██╗ ██████╗          ║")
	fmt.Println("║   	   ██╔══██╗██║   ██║██╔══██╗██║██╔═══██╗         ║")
	fmt.Println("║   	   ███████║██║   ██║██║  ██║██║██║   ██║         ║")
	fmt.Println("║   	   ██╔══██║██║   ██║██║  ██║██║██║   ██║         ║")
	fmt.Println("║   	   ██║  ██║╚██████╔╝██████╔╝██║╚██████╔╝         ║")
	fmt.Println("║   	   ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚═╝ ╚═════╝          ║")
	fmt.Println("║                                                       ║")
	fmt.Println("║  ███████╗████████╗██████╗ ███████╗ █████╗ ███╗   ███╗ ║")
	fmt.Println("║  ██╔════╝╚══██╔══╝██╔══██╗██╔════╝██╔══██╗████╗ ████║ ║")
	fmt.Println("║  ███████╗   ██║   ██████╔╝█████╗  ███████║██╔████╔██║ ║")
	fmt.Println("║  ╚════██║   ██║   ██╔══██╗██╔══╝  ██╔══██║██║╚██╔╝██║ ║")
	fmt.Println("║  ███████║   ██║   ██║  ██║███████╗██║  ██║██║ ╚═╝ ██║ ║")
	fmt.Println("║  ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝ ║")
	fmt.Println("║                                                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")
	fmt.Println("\033[0m") // reset color

	// INFO
	fmt.Println()
	fmt.Println("\033[1;33m» LOW-LATENCY AUDIO STREAMER\033[0m")
	fmt.Println("  ────────────────────────────────────────────────")
	fmt.Println("  Version 1.0 | ASIO | LINUX | WINDOWS | WebSocket")
	fmt.Println()
	fmt.Println("  MODULE: ", module)
	fmt.Println()
}
