/**
 * 功能：生成报告
 */
package main

import (
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// 创建报告
/*
	思路：
	1. 准备报告文件( 不存在即创建)
	2. 对比信息
	3. 写入表内容
*/
func createReport() {
	// 0. 读取，查询列表，信号包目录变量：currentSigFolderDir
	// currentSigFolderDirList := getSigpkgDirFromQueryHistoryFile(queryHistroyFilePath, "Sheet1")
	// logrus.Debug("func=createReport(), currentSigFolderDirList= ", currentSigFolderDirList)
	currentSigFolderDir := "12344" // 假设它= 这个，后面再改

	logrus.Debug("----进入方法, 创建发送记录excel, newSendExcel()")
	// 1. 准备报告文件( 不存在即创建)
	reportFile, err = createOrOpenExcelFile(reportFilePath) // 先注释,等在查询列表 写好后,再写到reportFile里
	// queryHistroyFile, err = createExcelIfNotExists(queryHistroyFileDir) // 写入到 查询列表.xlsx
	errorPanic(err)

	// 创建表 - 分析报告
	index, _ := reportFile.NewSheet("分析报告")
	// 设置表为 活动窗口
	reportFile.SetActiveSheet(index)
	// 设置列宽
	reportFile.SetColWidth("分析报告", "A", "F", 35)
	// 写入表头
	tableHeaders := []Any{"厂家", "具体机型路径", "要查询的机型", "查询结果", "异常原因", "总用时(单位: 分钟)"}
	err = reportFile.SetSheetRow("分析报告", "A1", &tableHeaders)
	errorPanic(err)

	// 2. 判断查询列表 最后2列结果 (有误差的结果 + 无误差的结果)
	// 查询最后2列
	// ?是否要openFile? - 不用
	rows, err := queryHistroyFile.GetRows("Sheet1")
	errorPanic(err)
	var boolResultNoMistakeList []string
	var boolResultHasMistakeList []string
	var boolResultDroneNameEqualList []string
	currentDroneStr := ""    // 用于区分？？，怎么说
	writeReportRowIndex := 2 // 写入报告表index
	// for index := 1; index <= len(rows); index++ {

	// 计算总时长 单位：秒
	var queryStartTime time.Time
	var queryEndTime time.Time
	for index, row := range rows {
		if index == 1 {
			logrus.Info("feed总用时(单位: 分钟) queryStartTimeStr= ", (row[5]))
			queryStartTime = string2time(row[5])
		}
		if index == len(rows)-1 {
			logrus.Info("feed总用时(单位: 分钟) queryEndTimeStr= ", (row[5]))
			queryEndTime = string2time(row[5])
		}
	}
	// str1 := "20241209-094427"
	// str2 := "20241209-094627"
	// queryStartTime = string2time(str1)
	// queryEndTime = string2time(str2)
	totalTime := queryEndTime.Sub(queryStartTime).Minutes()    // 总用时, 其实是查询总用时
	totalTimeStr := strconv.FormatFloat(totalTime, 'f', 2, 64) // 总用时, 其实是查询总用时 string类型
	logrus.Info("feed总用时(单位: 分钟) queryStartTime= ", queryStartTime)
	logrus.Info("feed总用时(单位: 分钟) queryEndTime= ", queryEndTime)
	logrus.Info("feed总用时(单位: 分钟) = ", totalTimeStr)

	for index, row := range rows { // 不好处理的写法, index 和row同步？
		currentSigFolderDir = row[4]
		if index == 0 {
			continue
		}
		logrus.Debugf("------- index=%v, currentDroneStr=%v, row[0]=%v", index, currentDroneStr, row[0])
		if index == 1 {
			currentDroneStr = row[0] // 把有数据的第一行，赋值给它
			logrus.Debugf("------- index(1)=%v, currentDroneStr=%v, row[0]=%v", index, currentDroneStr, row[0])
		}

		if currentDroneStr == row[0] {
			boolResultNoMistakeList = append(boolResultNoMistakeList, row[3])
			boolResultHasMistakeList = append(boolResultHasMistakeList, row[2])
			boolResultDroneNameEqualList = append(boolResultDroneNameEqualList, row[6])
		}
		// currentDroneStr 不一样的时候，写入报告一条数据, 或者是最后一条数据时,或者是信号包路径不一样时
		// if currentDroneStr != row[0] || index+1 == len(rows) { // 原来的写法
		if (index+1 < len(rows) && currentSigFolderDir != rows[index+1][4]) || index+1 == len(rows) { // 现在的写法
			logrus.Debug("boolResultNoMistakeList = ", boolResultNoMistakeList)
			logrus.Debug("boolResultNoMistakeList = ", boolResultHasMistakeList)
			oneSigReportResult, errorReason := checkAlgorithmWhereQueryResult(boolResultDroneNameEqualList, boolResultNoMistakeList, boolResultHasMistakeList)
			logrus.Infof("report 单个信号包,结果。currentSigFolderDir=%v, currentDroneStr=%v, oneSigReportResult =%v ", currentSigFolderDir, currentDroneStr, oneSigReportResult)

			// 写入行内容
			tableRow := []Any{"厂家??", currentSigFolderDir, currentDroneStr, oneSigReportResult, errorReason, totalTimeStr}
			logrus.Debug("---------------- 打算写入文件,currentDroneStr", currentDroneStr)
			logrus.Debug("---------------- 打算写入文件", &tableRow)
			err = reportFile.SetSheetRow("分析报告", "A"+strconv.Itoa(writeReportRowIndex), &tableRow)
			errorPanic(err)

			// 写入后, 重置变量
			boolResultDroneNameEqualList = []string{}
			boolResultNoMistakeList = []string{}
			boolResultHasMistakeList = []string{}
			if index+1 < len(rows) { // 不加这个，数组越界
				currentDroneStr = rows[index+1][0]
			}
			// boolResultNoMistakeList = append(boolResultNoMistakeList, row[3])   // 为什么写这2句？
			// boolResultHasMistakeList = append(boolResultHasMistakeList, row[2]) // 为什么写这2句？
			writeReportRowIndex++
		}
	}

	// 3. 写入表内容
	err = reportFile.SaveAs(reportFilePath)
	errorPanic(err)

}

// 处理查询结果的算法 - 写到 报告的表里
// 算法:
// (有误差/ 无误差 频率) + 机型名字 ，全true,才return true。所以先判断机型
// - 步骤3:机型名字结果,只要有TRUE (string类型),就返回true
// - 步骤1:无误差结果,只要有TRUE (string类型),就返回true
// - 步骤2:有误差结果,只要有TRUE (string类型),就返回true
// 参数: 1 没有误差的bool列表 2 有误差的结果bool列表
// 返回值：
// arg1: 正确？ bool
// arg2: 异常原因 errorReason  string 。 正常填nil
func checkAlgorithmWhereQueryResult(boolResultDroneNameEqualList []string, boolResultNoMistakeList []string, boolResultHasMistakeList []string) (bool, string) {
	logrus.Info("report检测算法, 机型名称结果 boolResultDroneNameEqualList= ", boolResultDroneNameEqualList)
	logrus.Info("report检测算法, 无误差结果 boolResultNoMistakeList= ", boolResultNoMistakeList)
	logrus.Info("report检测算法, 有误差结果 boolResultHasMistakeList= ", boolResultHasMistakeList)

	// 判断query excel 每一行的数据
	errorReason := "" // 异常内容
	// 检查机型名称是否匹配
	for index, boolResultDroneNameEqual := range boolResultDroneNameEqualList {
		if boolResultDroneNameEqual == "TRUE" && (boolResultNoMistakeList[index] == "TRUE" || boolResultHasMistakeList[index] == "TRUE") {
			return true, errorReason
		}
	}
	errorReason = "id不匹配, 或者频率误差<10M"
	return false, errorReason

	/* // 原先的写法，有问题，是判断列表只有要true,就按true来算。而不是按每行判断的
	errorReason := "" // 异常内容
	nameMatch := false
	noMistakeMatch := false
	hasMistakeMatch := false

	// 检查机型名称是否匹配
	for i, boolResultDroneNameEqual := range boolResultDroneNameEqualList {
		if boolResultDroneNameEqual == "TRUE" {
			nameMatch = true
			break
		}
	}

	// 检查至少一个 boolResultNoMistake 为 true
	for _, boolResultNoMistake := range boolResultNoMistakeList {
		if boolResultNoMistake == "TRUE" {
			noMistakeMatch = true
			break
		}
	}

	// 检查至少一个 boolResultHasMistake 为 true
	for _, boolResultHasMistake := range boolResultHasMistakeList {
		if boolResultHasMistake == "TRUE" {
			hasMistakeMatch = true
			break
		}
	}

	// 检查条件是否满足，机型名称相等且至少一个 boolResultNoMistake 或boolResultHasMistake 为 true
	if nameMatch && (noMistakeMatch || hasMistakeMatch) {
		return true, ""
	}

	if !nameMatch {
		errorReason = "机型名称不符"
	} else if !noMistakeMatch && !hasMistakeMatch {
		errorReason = "频率不匹配, 误差>10M"
	}

	return false, errorReason
	*/
}
