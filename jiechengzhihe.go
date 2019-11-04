package main

import "fmt"

func sum(n int)  {
	var a int
	for i := 1; i <= n ; i++{
		for j := 1 ; j <= i ; j++{
			fmt.Println("i,j",i,j)
			a += i*j
			fmt.Println("aa",a)
		}
	}
	fmt.Println(a)
}

func main()  {
	sum(2)
}
