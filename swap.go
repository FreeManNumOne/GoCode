package main

import "fmt"

func swap(a , b string)(string , string)  {
	return  b , a
}

func main()  {
	c ,d := swap("a" ,"b")
	fmt.Println(c ,d )
}
