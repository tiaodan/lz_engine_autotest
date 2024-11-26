/*
*
功能：处理错误
*/
package main

import "github.com/sirupsen/logrus"

// 终端程序，打印错误
func errorPanic(err error) {
	if err != nil {
		logrus.Panic("出异常了, 程序退出! err=", err)
	}
}

// 纯打印错误
func errorEcho(err error) {
	if err != nil {
		logrus.Error("出异常了, 打印错误。err=", err)
	}
}
