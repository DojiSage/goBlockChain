package main

import (
	"fmt"
	"goblockchain/utils"
)


func main() {
	fmt.Println(utils.IsFoundHost("127.0.0.1", 5000))
}