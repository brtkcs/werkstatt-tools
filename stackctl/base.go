package main

import (
	"fmt"
	"os/exec"
)

func main() {
	// exec.Command – mint bash-ban a $(command)
	// Ez futtatja: podman ps --format "{{.Names}} {{.Status}}"
	cmd := exec.Command("podman", "ps", "--format", "{{.Names}} {{.Status}}")

	// Output() – lefuttatja és visszaadja az eredményt
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("hiba:", err)
		return
	}

	fmt.Println(string(output))
}
