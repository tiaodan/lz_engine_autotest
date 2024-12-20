/**
 * 功能：发送信号
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

var addrTCP *net.TCPAddr
var connTCP *net.TCPConn

// 发送初始化
func sendInit() {
	fmt.Println("----进入方法: 开启发送任务 init , sendInit()")
	var err error
	addrTCP, err = net.ResolveTCPAddr("tcp", devIp+":8000")
	errorPanic(err)
	fmt.Println("正在连接tcp")
	for i := 0; i < 5; i++ {
		connTCP, err = net.DialTCP("tcp", nil, addrTCP)
		if err != nil {
			fmt.Printf("tcp连接失败，重试第%d次\n", i+1)
			// addlog(err)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			sendIsStart <- any // 发送信号：发送程序开始
			fmt.Println("tcp拨号成功")
			return
		}
	}
	fmt.Println("tcp连接失败，本次运行结束")
	errorPanic(err)
}

func sendTask() {
	fmt.Println("----进入方法: 开启发送任务, sendTask()")

	// 获取信号列表， from 待发送列表excel
	fmt.Println("----进入方法: 开启发送任务, sendTask(), preSendHistoryFilePath=", preSendHistoryFilePath)
	sigpkgList = getSigpkgListFromPreSendHistoryFile(preSendHistoryFilePath, "待发送列表")
	logrus.Info("func=sendTask(), sigpkgList= ", sigpkgList)
	droneObjList = getQueryDroneFromPreSendHistoryFile(preSendHistoryFilePath, "待发送列表")
	logrus.Debug("func=sendTask(), droneObjList= ", droneObjList)
	sigFolderPathList = getSigFolderPathFromPreSendHistoryFile(preSendHistoryFilePath, "待发送列表")
	logrus.Debug("func=sendTask(), sigFolderPathList= ", sigFolderPathList)

	sendInit()
	// changeFolderFlag := false // 换文件夹标志
	for i, sigpkg := range sigpkgList {
		// 设置当前飞机，我自己加的代码
		if i == 0 {
			currentQueryTargetDrone = droneObjList[i] // 当前飞机，用于查询列表excel用
			currentSigDirPath = sigFolderPathList[i]  // 当前信号文件夹路径
		}

		// v0.0.0.1 新增，为优化查询效率。一旦检测到成功的，剩下的信号就不发了。就直接切换信号包
		/*
			这段写法有问题，还没研究
			if <-userChangeQuerySigFolder {
				continue
			}
		*/
		/*
			这段代码写在这，会阻塞，暂时没办法解决
			if !changeFolderFlag {
			select {
			case <-userChangeQuerySigFolder:
				logrus.Info("收到信号：切换信号文件夹标志, 修改flag")
				if i+1 < len(droneObjList) && sigFolderPathList[i] == sigFolderPathList[i+1] { // 不加这个，数组越界
					changeFolderFlag = true
				}
			}
			}
		*/
		logrus.Infof("index=%v, 切换信号文件夹标志 changeFolderFlag =%v, changeFolderFlagNum=%v", i, changeFolderFlag, changeFolderFlagNum)
		if sigpkg != "[换文件夹]" && changeFolderFlag && changeFolderFlagNum == 1 { // 切换文件夹消息数量==1时才认，只认收到的第一条消息
			logrus.Infof("index=%v, 切换信号文件夹标志=true, 跳过当前循环, sig=%v", i, sigpkg)
			continue
		}
		// else {
		// 	changeFolderFlagNum = 0 // 重置
		// }

		// copy过来的代码
		fmt.Printf("发送信号, index = %v, tasklist= %v \n", i, sigpkg)
		logrus.Infof("发送信号, index = %v, tasklist= %v ", i, sigpkg)
		if sigpkg == "[换文件夹]" {
			logrus.Infof("index=%v, 【换文件夹】切换信号文件夹标志 changeFolderFlag =%v, changeFolderFlagNum=%v", i, changeFolderFlag, changeFolderFlagNum)
			/*
				// 原来的写法
					// writeSendExcel(i, "[换文件夹]", time.Now()) //
					if i+1 < len(droneObjList) { // 不加这个，数组越界
						// currentQueryTargetDrone = droneObjList[i+1] // 当前飞机，用于查询列表excel用  - 这样写，在查询等待的时间里，还是当前飞机=下一个飞机了
						// currentSigDirPath = sigFolderPathList[i+1]  // 当前信号文件夹路径  - 这样写，在查询等待的时间里，还是当前飞机=下一个飞机了
						changeFolderFlag = false // 标志重置
					}
			*/
			fmt.Println("[换文件夹]，等待", cdFolderInterval, "秒后发送")
			logrus.Info("[换文件夹]，等待", cdFolderInterval, "秒后发送")
			logrus.Info("[换文件夹]，等待期间, 当前飞机currentDrone= ", currentQueryTargetDrone)

			select {
			case <-time.After(time.Duration(cdFolderInterval) * time.Second):
				if i+1 < len(droneObjList) { // 不加这个，数组越界
					currentQueryTargetDrone = droneObjList[i+1] // 当前飞机，用于查询列表excel用 - 这样写，在查询等待的时间里，还是当前飞机，而不是下一个飞机了
					currentSigDirPath = sigFolderPathList[i+1]
				}
				// 等待实际结束后，再重置变量。 代码放这里，解决：切换文件夹期间，查到数据，会影响文件夹切换逻辑
				changeFolderFlagNum = 0  // 重置
				changeFolderFlag = false // 标志重置
			case <-userEndSend: // 匹配到信号，用户终止发送
				connTCP.Close()
				fmt.Println("sigpkg == [换文件夹】分支, userEndSend, 关闭tcp")
				userEndQuery <- any
				return
			}
		} else {
			count, err := send(sigpkg)
			if err != nil {
				t := time.Now()
				ts := fmt.Sprintf("%d.%02d.%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
				// writeSendExcel(i, "发送失败", t)
				fmt.Println(ts, "发送失败", sigpkg)
			} else {
				t := time.Now()
				ts := fmt.Sprintf("%d.%02d.%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
				// writeSendExcel(i, ts, t)
				fmt.Println(ts, "发送", sigpkg)
			}
			select {
			case <-time.After(time.Duration(sigPkgSendInterval * count)):
			case <-userEndSend:
				connTCP.Close()
				fmt.Println("关闭tcp")
				userEndQuery <- any // 发送信号：用户停止查询
				return
				// case <-userChangeQuerySigFolder:
				// 	logrus.Info("收到信号：切换信号文件夹标志, 修改flag")
				// 	if i+1 < len(droneObjList) && sigFolderPathList[i] == sigFolderPathList[i+1] { // 不加这个，数组越界
				// 		changeFolderFlag = true
				// 	}
			}
		}
	}
	connTCP.Close()
	fmt.Println("发送完毕")
	sendIsEnd <- any // 发送信号：发送程序结束
	logrus.Info("发送完毕，发送信号, SendIsEnd")
}

func send(url string) (int, error) {
	var count int
	bytes, err := os.ReadFile(url)
	errorPanic(err)
	count, err = connTCP.Write(bytes)
	if err != nil {
		err2 := connTCP.Close()
		errorPanic(err2)
		fmt.Println(url)
		fmt.Println("tcp发送失败，等待3秒后重试")
		sendReset()
		count, err = connTCP.Write(bytes)
		if err != nil {
			return 0, errors.New("发送失败")
		}
		return 1000000, nil
	}
	return count, nil
}

func sendReset() {
	var err error
	connTCP.Close()
	time.Sleep(3 * time.Second)
	for i := 0; i < 5; i++ {
		connTCP, err = net.DialTCP("tcp", nil, addrTCP)
		if err != nil {
			fmt.Printf("tcp连接失败，重试第%d次\n", i+1)
			// addlog(err)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			fmt.Println("tcp拨号成功")
			return
		}
	}
	fmt.Println("tcp连接失败，本次运行结束")
	errorPanic(err)
}

func closeSend() {
	fmt.Println("----进入方法, 关闭发送, closeSend()")
	if connTCP != nil {
		logrus.Debug("connTCP 空了, 关闭连接")
		connTCP.Close()
	}
}

/*
功能：从excel获取信号列表
参数：
1. path 文件路径, 建议绝对路径
2. sheetName 表名

返回值：
1. sigpkgList []string
*/
func getSigpkgListFromPreSendHistoryFile(path string, sheetName string) []string {
	sigpkgList = make([]string, 0) // 重置信号列表为空
	file, err := createOrOpenExcelFile(path)
	logrus.Debugf("func=getSigpkgListFromPreSendHistoryFil(), path===== %v, sheetName=%v", path, sheetName)
	if err != nil {
		logrus.Error("func=getSigpkgListFromPreSendHistoryFil(), 文件不存在, Error reading directory= ", err)
		return sigpkgList
	}

	// 获取工作表 信号 D列
	rows, err := file.Rows(sheetName)
	errorPanic(err)

	index := 2
	for rows.Next() {
		value, err := preSendHistoryFile.GetCellValue(preSendHistoryFileSheetName, "D"+strconv.Itoa(index))
		errorPanic(err)
		if value != "" { // 不判断发送最后一条空数据时，报错
			sigpkgList = append(sigpkgList, value)
		}
		index++
	}
	return sigpkgList
}

/*
功能：从excel获取要查询的飞机（信号列表前一列）
参数：
1. path 文件路径, 建议绝对路径
2. sheetName 表名

返回值：
1. []Drone
*/
func getQueryDroneFromPreSendHistoryFile(path string, sheetName string) [][]Drone {
	droneList := make([][]Drone, 0) // 重置目标飞机列表为空. 二维数组
	file, err := createOrOpenExcelFile(path)
	logrus.Debugf("func=getQueryDroneFromPreSendHistoryFile(), path===== %v, sheetName=%v", path, sheetName)
	if err != nil {
		logrus.Error("func=getQueryDroneFromPreSendHistoryFile(), 文件不存在, Error reading directory= ", err)
		return droneList
	}

	// 获取工作表 要查询的飞机 C列
	rows, err := file.Rows(sheetName)
	errorPanic(err)

	rowCount, err := getExcelRowsCount(path, sheetName) // 有数据的总行数
	errorPanic(err)
	index := 2
	for rows.Next() {
		// 避免获取最后一条是空数据
		if index > rowCount {
			break
		}
		value, err := preSendHistoryFile.GetCellValue(preSendHistoryFileSheetName, "C"+strconv.Itoa(index))
		logrus.Debug("value==========", value)
		// 可能会获取到空数据
		errorPanic(err)
		var droneObj []Drone
		err = json.Unmarshal([]byte(value), &droneObj)
		errorPanic(err)

		if value != "" { // 不判断发送最后一条空数据时，报错
			droneList = append(droneList, droneObj)
		}
		index++
	}
	return droneList
}

/*
功能：从excel获取信号文件夹列表（信号列表前2列）
参数：
1. path 文件路径, 建议绝对路径
2. sheetName 表名

返回值：
1. sigFolderPathList []string
*/
func getSigFolderPathFromPreSendHistoryFile(path string, sheetName string) []string {
	folderList := make([]string, 0) // 重置信号文件夹列表为空
	file, err := createOrOpenExcelFile(path)
	logrus.Debugf("func=getSigFolderPathFromPreSendHistoryFile(), path===== %v, sheetName=%v", path, sheetName)
	if err != nil {
		logrus.Error("func=getSigFolderPathFromPreSendHistoryFile(), 文件不存在, Error reading directory= ", err)
		return folderList
	}

	// 获取工作表 信号 B列
	rows, err := file.Rows(sheetName)
	errorPanic(err)

	index := 2
	for rows.Next() {
		value, err := preSendHistoryFile.GetCellValue(preSendHistoryFileSheetName, "B"+strconv.Itoa(index))
		errorPanic(err)
		if value != "" { // 不判断发送最后一条空数据时，报错
			folderList = append(folderList, value)
		}
		index++
	}
	return folderList
}
