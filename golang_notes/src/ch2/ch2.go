// package name
package main

// import other packages
import (
	"fmt"
	"tempconv"
)

type student struct {
	name string
	age  int
}

var (
	x int     = 0
	y float64 = 1.1
	z bool    = false
)

const (
	dayOfWeek = 7
)

func main() {
	gexin := student{name: "gexin", age: 27}
	p := f1()
	var x int8
	x = -5 % (-3)

	fmt.Println(gexin, x, y, z, dayOfWeek, *p, tempconv.FreezingC)

	fmt.Printf("x: 0b %08b", x)
}

func f1() *int {
	v := 10
	return &v
}
