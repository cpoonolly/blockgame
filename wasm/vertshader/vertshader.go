package vertshader

import (
	"fmt"
	"syscall/js"
)

// VertexShader represents a vertex shader program
type VertexShader struct {
	id      js.Value
	srcCode string
}

// New contstruct a new VertexShader
func (verShader *VertexShader) New() {
	fmt.Println("hello world")
}
