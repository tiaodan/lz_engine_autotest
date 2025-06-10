/**
 * 功能：处理http请求，响应
 */
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type queryDrone map[string]Any

type queryData map[string][]queryDrone

type queryResult map[string]queryData

var (
	queryIp       string
	queryParam    = "{query: { drone\n {\n id \n name\n seen_sensor\n {\n detected_freq_khz\n }\n }\n}}"
	postParam     = bytes.NewBuffer([]byte(queryParam))
	queryExcellen int

	countDownTimes int // 检测到后，多少次没检测到
	findTimes      = 0 // 检测到飞机次数
)

func queryTask() {
	fmt.Println("----进入方法: 开启查询任务, queryTask()")
	// 初始化
	countDownTimes = noQueryTimes2NextSig // 初始化计数器

	<-sendIsStart
	if len(queryIp) == 0 {
		queryIp = "http://" + devIp + ":3200/graphql"
	}

	ticker := time.NewTicker(time.Duration(queryDroneInterval) * time.Second)
	defer ticker.Stop()
	var res *http.Response
	var err error
	// timeout := time.After(time.Duration(queryDroneInterval+5) * time.Second) // 设置超时时间为 queryDroneInterval + 5 秒 - no use
	// var timeoutNum float64
	for {
		logrus.Debug("queryTask(), 查询任务, for循环")
		// fmt.Println("-------------------------------------------------queryTask(), 查询任务, for循环--阻塞")
		select {
		case <-sendIsEnd:
			fmt.Println("-------------------------------------------------停止查询, sendIsEnd")
			logrus.Info("停止查询, sendIsEnd")
			// 关闭文件
			queryHistroyTxtFile.Close() // 直接写应该没有问题
			return
		case <-userEndQuery:
			fmt.Println("-------------------------------------------------停止查询, userEndQuery")
			logrus.Info("停止查询, userEndQuery")
			queryHistroyTxtFile.Close() // 直接写应该没有问题
			return
		// case <-time.After(time.Duration(queryDroneInterval+5) * time.Second):
		// 	logrus.Error("超时, 停止查询")
		// 	return
		case <-ticker.C: // 这里代码有问题，如果和 case <-sendIsEnd: case <-userEndQuery: 并列写，有可能ticker.C 阻塞那几秒内，收到了终止信息，又无法进入那个case进行处理
			// fmt.Println("-------------------------------------------------继续查询, ticker.C")
			logrus.Info("-------------------------------------------------继续查询, ticker.C")
			// 重连5次
			for j := 0; ; j++ {
				q := strings.NewReader(`{"query":"{drone\n {\n id \n name \n seen_sensor\n {\n detected_freq_khz\n }\n }\n}"}`)
				res, err = http.Post(queryIp, "application/json;charset=utf-8", q)
				if err != nil {
					if j > 2 {
						fmt.Println("graph连接失败，本次运行结束")
						logrus.Error("graph连接失败，本次运行结束")
						errorPanic(err)
					}
					fmt.Printf("graph连接失败，重试第%d次\n", j+1)
					// addlog(err)
					time.Sleep(time.Second)
				} else if res.StatusCode != 200 {
					if j > 2 {
						fmt.Println("graph连接失败，本次运行结束")
						logrus.Error("graph连接失败，本次运行结束")
						errorPanic(errors.New(res.Status))
					}
					fmt.Printf("graph连接失败，重试第%d次\n", j+1)
					// addlog(errors.New(res.Status))
					time.Sleep(time.Second)
				} else {
					break
				}
			}

			respBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				logrus.Error("queryTask, 进行到这, 报错了。。")
				errorPanic(err)
			}
			res.Body.Close()
			result := new(queryResult)
			json.Unmarshal(respBytes, result)
			logrus.Debug("查到的 result = ", result)
			list := (*result)["data"]["drone"]
			logrus.Debug("查到的drone list = ", list)
			droneList := []Drone{}
			// 制作查询飞机列表
			for _, v := range list {
				ids := []string{}
				id := v["id"].(string)
				ids = append(ids, id)
				s := v["name"].(string)
				d := Drone{Id: ids, Name: s, FreqList: 0}
				f := v["seen_sensor"].([]interface{})
				// logrus.Debug("查到的 s = ", s)
				// logrus.Debug("查到的 d = ", d)
				// logrus.Debug("查到的 f = ", f)
				for _, w := range f {
					w1 := w.(map[string]interface{})
					w2 := w1["detected_freq_khz"]
					freq := w2.(float64)
					d.FreqList = int(freq)
				}
				droneList = append(droneList, d)
			}

			logrus.Info("查询到无人机= ", droneList)
			logrus.Debug("查询任务中,目标飞机preQueryDrone= ", currentQueryTargetDrone)                             // [{3DR Solo [2462000] 8adc9635291e}] 原来写法: preQueryDrone := Drone{id: currentSigPkgDroneId, name: "3DR Solo", freqList: []int{2462000}}
			queryResultHasMistake := checkAlgorithmWhereFreqHasMistake(droneList, currentQueryTargetDrone) // 判断是否查询到 ,true / false, 后面再写逻辑,判断true/false
			queryResultNoMistake := checkAlgorithmWhereFreqNoMistake(droneList, currentQueryTargetDrone)   // 判断是否查询到 ,true / false, 后面再写逻辑,判断true/false
			currentTime := time.Now()
			droneNameIsEqual := checkAlgorithmWhereDroneNameIsEqual(droneList, currentQueryTargetDrone)                                                        // 判断是否查询到 ,true / false, 后面再写逻辑,判断true/false
			writeQueryExcel(currentQueryTargetDrone, droneList, queryResultHasMistake, queryResultNoMistake, currentSigDirPath, currentTime, droneNameIsEqual) // 最早的参数-3个,// 表头: 时间, 飞机, 时间戳, 要查询的飞机id, 查询结果-有没有(true/ false)
			// writeQuery2Txt(currentQueryTargetDrone, droneList, queryResultHasMistake, queryResultNoMistake, currentSigDirPath, currentTime)                    // 最早的参数-3个,// 表头: 时间, 飞机, 时间戳, 要查询的飞机id, 查询结果-有没有(true/ false)
			// v0.0.0.1 为了提高效率新增，一旦判断 queryResultNoMistake==true 或者queryResultNoMistake==false && queryResultHasMistake== true了，就不继续查了。发送个信号
			// 最新逻辑，发一次信号就好了，不多发信号
			if droneNameIsEqual && (queryResultNoMistake || queryResultHasMistake) && changeFolderFlagNum == 0 { // 如果有问题，用这个
				// if droneNameIsEqual && (queryResultNoMistake || queryResultHasMistake) { // 切换文件夹期间，查到数据，会影响文件夹切换逻辑
				logrus.Info("查询到正确数据, queryResultNoMistake==true, 发送信号: 切换文件夹")
				// userEndQuery <- any // 发送信号：用户停止当此查询 - 发这个信号不对，只能查1个
				// userEndSend <- any // 发送信号：用户停止当此查询 - 发这个信号也不对，只能查1个
				// userChangeQuerySigFolder <- any // 发送信号：换信号文件夹
				changeFolderFlag = true
				changeFolderFlagNum = 1
				// changeFolderFlagNum += 1
				logrus.Infof("切换信号文件夹标志 changeFolderFlag =%v, changeFolderFlagNum=%v", changeFolderFlag, changeFolderFlagNum)
			}

			if droneNameIsEqual && (queryResultNoMistake || queryResultHasMistake) {
				countDownTimes = noQueryTimes2NextSig // 检测到，重置计数器
				findTimes += 1
			} else {
				countDownTimes -= 1 // 没检测到
			}

			logrus.Infof("countDownTimes=%v, changeFolderFlag=%v, findTimes=%v, changeFolderFlagNum=%v", countDownTimes, changeFolderFlag, findTimes, changeFolderFlagNum)
			// 发送信号条件：每次都没测到6次，或者 测到过等消失后4次。这样效率高
			// if (findTimes == 0 && countDownTimes == 0) || (findTimes >= 1 && countDownTimes <= 5) {
			if (countDownTimes == 0) || (findTimes >= 1 && countDownTimes <= noQueryTimes2NextSig-afterQueriedWaitTimes) {
				msgChangeFolder <- any // 发送切换文件夹信号
				// 跳过当前循环标志 start,防止上一步阻塞
				changeFolderFlag = true
				changeFolderFlagNum = 1
				// 跳过当前循环标志 end,防止上一步阻塞
				countDownTimes = noQueryTimes2NextSig // 检测到，重置计数器
				findTimes = 0
				logrus.Info("------------- 发送完信号,并重置完计数器: 换文件夹, msgChangeFolder <- any")
			}

			// default:
			// 	fmt.Println("-------------------------------------------------判断超时, 查询,for 默认分支")
			// 	logrus.Error("查询,for 默认分支，timeoutNum = ", timeoutNum)
			// 	timeoutNum += 0.5
			// 	time.Sleep(time.Millisecond * 500) // 暂停 0.5 秒
			// 	if timeoutNum >= float64(queryDroneInterval+5) {
			// 		logrus.Error("超时, 停止查询")
			// 		return
			// 	}
		}
	}
}

// 发送请求，参数graphql
// 参数 url 请求链接 string
// 参数 method 请求方式 string
// 参数 graphqlJsonStr 转成json的接口 string
func SendRequestByGraphql(url string, method string, graphqlStr string) string {
	// url := "https://192.168.84.248/rf/graphql"
	// url := "https://192.168.85.93/rf/graphql"
	// method := "POST"
	fmt.Println("SendRequestByGraphql, graphqlStr = ", graphqlStr)
	payload := strings.NewReader(graphqlStr)

	// 禁用 TLS 证书验证
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	// client := &http.Client{}  // 原始写法
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		log.Println("请求失败，错误=", err)
		return ""
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
	req.Header.Add("Content-Type", "application/graphql")
	// 为什么必须加上 Bearer? 84.248用的
	// req.Header.Add("Authorization", "Bearer "+config.Token)  // ------ 这段代码，copy过来不适配，待修改
	req.Header.Add("Authorization", "Bearer "+"config.Token") // 临时随便写了一个token
	// 85.93 用的
	// req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6Ijg4MjhjYTdlLTkyMWQtMTFlYi05NDhkLTlmMjQ2NjMyMzk2OCIsImlhdCI6MTcyMTM1ODkzMiwiZXhwIjoxNzUyODk0OTMyfQ.hlKKSQRHT2XeGsfCzc-UdxeWW3StnQ14oEYW-QC0VJ0")

	res, err := client.Do(req)
	if err != nil {
		log.Println("响应失败，错误=", err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("读取响应失败，错误=", err)
		return ""
	}
	log.Println("响应内容= ", string(body))
	return string(body)
}

// 检测算法-频率可以有误差 (频率误差<=10 Mhz)
// 参数:已查到的飞机列表, 目标飞机
// 频率单位 Hz 2462000 - 2452000 = 10000
func checkAlgorithmWhereFreqHasMistake(queriedDrones []Drone, targetDrones []Drone) bool {

	// 只有一个id的写法
	// for _, queriedDrone := range queriedDrones {
	// 	// 判断查询到的飞机列表,id是否相等
	// 	if queriedDrone.Id == targetDrone.Id {
	// 		// 判断频率,误差是否 <=10 Mhz
	// 		freqMistake := math.Abs(float64(queriedDrone.FreqList) - float64(targetDrone.FreqList))
	// 		if freqMistake <= 10000 {
	// 			return true
	// 		}
	// 	}
	// }
	// return false

	// 多个id的写法 funk.Contains() arg1: list arg2: string (要是是个数组，不行)
	if len(targetDrones) == 1 { // targetDrone 数组只有一个数据
		for _, queriedDrone := range queriedDrones {
			// 分id.txt 内容是 随机 和具体id情况
			if funk.Contains(targetDrones[0].Id, "随机") {
				// 判断频率,误差是否 <=10 Mhz
				freqMistake := math.Abs(float64(queriedDrone.FreqList) - float64(targetDrones[0].FreqList))
				if freqMistake <= float64(mistakeFreqConfig*1000) {
					return true
				}
			} else {
				// 判断查询到的飞机列表,id是否相等
				if funk.Contains(targetDrones[0].Id, queriedDrone.Id[0]) {
					// 判断频率,误差是否 <=10 Mhz
					freqMistake := math.Abs(float64(queriedDrone.FreqList) - float64(targetDrones[0].FreqList))
					if freqMistake <= float64(mistakeFreqConfig*1000) {
						return true
					}
				}
			}
		}
	} else if len(targetDrones) > 1 {
		for _, queriedDrone := range queriedDrones {
			for _, targetDrone := range targetDrones {
				// 分id.txt 内容是 随机 和具体id情况
				if funk.Contains(targetDrone.Id, "随机") {
					// 判断频率,误差是否 <=10 Mhz
					freqMistake := math.Abs(float64(queriedDrone.FreqList) - float64(targetDrone.FreqList))
					if freqMistake <= float64(mistakeFreqConfig*1000) {
						return true
					}
				} else {
					// 判断查询到的飞机列表,id是否相等
					if funk.Contains(targetDrone.Id, queriedDrone.Id[0]) {
						// 判断频率,误差是否 <=10 Mhz
						freqMistake := math.Abs(float64(queriedDrone.FreqList) - float64(targetDrone.FreqList))
						if freqMistake <= float64(mistakeFreqConfig*1000) {
							return true
						}
					}
				}
			}
		}
	}
	return false

}

// 检测算法-频率可以有误差 (频率误差<=10 Mhz)
// 参数:已查到的飞机列表, 目标飞机
// 频率单位 Hz 2462000 - 2452000 = 10000
func checkAlgorithmWhereFreqNoMistake(queriedDrones []Drone, targetDrones []Drone) bool {
	/*
		//  只有一个id的写法
			for _, queriedDrone := range queriedDrones {
				// 判断查询到的飞机列表,id是否相等
				if queriedDrone.Id == targetDrone.Id {
					// 判断频率,误差是否 <=0 Mhz
					if queriedDrone.FreqList == targetDrone.FreqList {
						return true
					}
				}
			}
			return false
	*/

	// 多个id的写法 funk.Contains() arg1: list arg2: string (要是是个数组，不行)
	if len(targetDrones) == 1 { // targetDrone 数组只有一个数据
		for _, queriedDrone := range queriedDrones {
			// 分id.txt 内容是 随机 和具体id情况
			if funk.Contains(targetDrones[0].Id, "随机") {
				// 判断频率,误差是否 <=0 Mhz
				if queriedDrone.FreqList == targetDrones[0].FreqList {
					return true
				}
			} else {
				// 判断查询到的飞机列表,id是否相等
				if funk.Contains(targetDrones[0].Id, queriedDrone.Id[0]) {
					// 判断频率,误差是否 <=0 Mhz
					if queriedDrone.FreqList == targetDrones[0].FreqList {
						return true
					}
				}
			}
		}
	} else if len(targetDrones) > 1 {
		for _, queriedDrone := range queriedDrones {
			for _, targetDrone := range targetDrones {
				// 分id.txt 内容是 随机 和具体id情况
				if funk.Contains(targetDrone.Id, "随机") {
					// 判断频率,误差是否 <=0 Mhz
					if queriedDrone.FreqList == targetDrone.FreqList {
						return true
					}
				} else {
					// 判断查询到的飞机列表,id是否相等
					if funk.Contains(targetDrone.Id, queriedDrone.Id[0]) {
						// 判断频率,误差是否 <=0 Mhz
						if queriedDrone.FreqList == targetDrone.FreqList {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// 检测算法-droneName 是否相等
// 参数:已查到的飞机列表, 目标飞机
func checkAlgorithmWhereDroneNameIsEqual(queriedDrones []Drone, targetDrones []Drone) bool {
	for _, queriedDrone := range queriedDrones {
		for _, targetDrone := range targetDrones {
			// 判断查询到的飞机列表,id是否相等。
			/*
				// 区分大小写写法
				if queriedDrone.Name == targetDrone.Name {
					return true
				}
			*/
			// 不区分大小写写法
			if strings.EqualFold(queriedDrone.Name, targetDrone.Name) {
				return true
			}
		}
	}
	return false
}

// func writeQueryExcel(ts string, res []Drone, t time.Time) {
// 之前的参数-6个,// 表头: 时间, 飞机, 时间戳, 要查询的飞机id, 查询结果(带误差)-有没有(true/ false), 查询结果(无误差)-有没有(true/ false)
// 现在的参数-?个,// 表头: 要查询的飞机id, 查询到的飞机, 查询结果(带误差)-有没有(true/ false), 查询结果(无误差)-有没有(true/ false), 信号文件夹路径,查询时间(用于统计总用时), 机型名称是否相等
func writeQueryExcel(preQueryDrone []Drone, res []Drone, queryResultHasMistake bool, queryResultNoMistake bool, sigFolderPath string, currentTime time.Time, droneNameIsEqual bool) {
	logrus.Debug("写入查询excel, 写入内容 == ", preQueryDrone, res, queryResultHasMistake, queryResultNoMistake)
	// 打开文件,不存在就创建
	queryHistroyFile, err = createOrOpenExcelFile(queryHistroyFilePath)
	// index, _ := e.NewSheet("发送记录")
	// 切为活动窗口
	// e.SetActiveSheet(index)
	queryHistroyFile.SetColWidth("sheet1", "A", "E", 30)

	// 写入表头
	// 表头: 时间, 飞机, 时间戳, 要查询的飞机id, 查询结果-有没有(true/ false), 信号文件夹路径
	err = queryHistroyFile.SetSheetRow("sheet1", "A1", &[]Any{"要查询的飞机id", "查询到的飞机", "查询结果(有id并且频率误差<10)-有没有(true/false)", "查询结果(有id并且频率相等)-有没有(true/false)", "信号文件夹路径", "当前时间(用于计算总时间)", "机型名称是否相等"})

	// 写入行内容
	queryExcellen++

	tableRow := &[]Any{preQueryDrone, res, queryResultHasMistake, queryResultNoMistake, sigFolderPath, time2stringforFilename(currentTime), droneNameIsEqual}
	logrus.Infof("写入查询记录表, 当前drone=%v, 信号包路径=%v", preQueryDrone, sigFolderPath)
	err = queryHistroyFile.SetSheetRow("Sheet1", "A"+strconv.Itoa(queryExcellen+1), tableRow)
	errorPanic(err)
	logrus.Debug("写入查询文件, filePath= ", queryHistroyFilePath)
	err = queryHistroyFile.SaveAs(queryHistroyFilePath)
	errorPanic(err)
}

// 写入查询结果到txt文件
// 现在的参数-?个,// 表头: 要查询的飞机id, 查询到的飞机, 查询结果(带误差)-有没有(true/ false), 查询结果(无误差)-有没有(true/ false), 信号文件夹路径,查询时间(用于统计总用时)
func writeQuery2Txt(preQueryDrone []Drone, res []Drone, queryResultHasMistake bool, queryResultNoMistake bool, sigFolderPath string, currentTime time.Time) {
	// logrus.Debug("写入查询结果到txt文件, 写入内容 == ", preQueryDrone, res, queryResultHasMistake, queryResultNoMistake)
	// logrus.Info("创建查询txt文件")
	// 1. 创建或者打开文件
	queryHistroyTxtFile, err = createOrOpenTxtFile(queryHistroyFileTxtPath)
	errorPanic(err)

	// 5. 创建表头
	_, err = queryHistroyTxtFile.WriteString("要查询的飞机id, 查询到的飞机, 查询结果(有id并且频率误差<10)-有没有(true/false), 查询结果(有id并且频率相等)-有没有(true/false), 信号文件夹路径, 当前时间(用于计算总时间)\n")
	if err != nil {
		logrus.Error("写入查询txt文件表头失败, err=", err)
	}

	// 6. 写入行内容
	rowStr := fmt.Sprintf("%v, %v, %v, %v, %v, %v \n", preQueryDrone, res, queryResultHasMistake, queryResultNoMistake, sigFolderPath, currentTime)
	_, err = queryHistroyTxtFile.WriteString(rowStr)
	if err != nil {
		logrus.Error("写入查询txt文件表头失败, err=", err)
	}
}
