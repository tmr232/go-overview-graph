package demo

import "fmt"

func Demo(flag bool, n int) {
	if flag {
		fmt.Println("Hello, World!")
	}
	fmt.Println("Oh no!")

	for _, x := range "Hello, World!" {
		fmt.Println(x)
	}

	if flag {
		fmt.Println("A")
	} else if n == 3 {
		fmt.Println("B")
	} else if n == 2 {
		fmt.Println("C")
	} else {
		fmt.Println("D")
	}

	switch n {
	case 1:
		return
	case 2:
		return
	case 3:
		return
	}
	fmt.Println("Yo")
}

type Thing struct{}

func (t *Thing) Action() {
	func() {
		fmt.Println("H")
	}()
	if t == nil {
		fmt.Println("Oh no!")
	} else {
		fmt.Println("Yes!")
	}
}

func f() {
	(&Thing{}).Action()
}
