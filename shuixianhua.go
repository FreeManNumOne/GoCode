package main

import "fmt"

func test(x int ) {
	var a int = x / 100
	var b int = (x/10)%10
	var c int = x % 10
	var d = a*a*a + b*b*b +c*c*c
	if d == x {
		fmt.Println("水仙花数",x)
	}else {
		fmt.Println("不是水仙花数",x)
	}
}

func main(){
	for i :=0 ; i < 1000 ;i++{
		test(i)
	}
}
