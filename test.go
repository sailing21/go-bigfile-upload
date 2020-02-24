package main

import (
	"fmt"
)

func t1() {
	fmt.Println("t1 start")
	defer fmt.Println("t1 end")
	fmt.Println("t1 process")
}
func main() {
	//t1()
	//fmt.Println("main end")
	//arr := make([] int,5)
	//arr = []int{1,5,4,6,0}
	//sort.Ints(arr)
	//for i:=0; i<5; i++ {
	//	fmt.Println(arr[i])
	//}
	//fmt.Println(time.Unix(0,0*int64(time.Millisecond)).Format("2006-01-02 15:04:05"))
	//p := "./upload/统计学习方法1.pdf"
	//os.Chtimes(p,time.Unix(0,1582260876652*int64(time.Millisecond)),time.Unix(0,1582260876652*int64(time.Millisecond)))
	//os.RemoveAll("./upload/2b3b210a58be4fdef030a0c289fb46fa")
	//f,e := os.Stat(p)
	//fmt.Println(f)
	//fmt.Println(e)
	//fmt.Println(os.IsExist(e))

}
