package main

//
// 包名可以和路径不一致, 路径仅引导编译器加载源文件, 
// 程序中使用包中 package 声明的包名称作为前缀引用包中的 api.
//
import (
	"witness-enterprise/witness"
	"witness-enterprise/web"
)

func main() {
	witness.StartWitnessProgram()
	witness_web.StartWebService()
	witness.StartHttpServer()
}