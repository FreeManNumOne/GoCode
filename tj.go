package main

import "fmt"

func main() {
	str := []string{"how", "do", "you", "do", "how", "do", "you", "do", "how", "do", "you", "do", "how", "do", "you", "do", "how", "do", "you", "do"}
	var tj map[string]int
	tj = make(map[string]int, 10)
	for _, v := range str {
		fmt.Printf("%s\n", v)
		_, ok := tj[v]
		if !ok {
			tj[v] = 1
		} else {
			tj[v] += 1
		}
	}
	fmt.Println(tj)
}
