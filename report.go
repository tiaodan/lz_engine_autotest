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

// 功能：已生成的报告，整合到 机型库里:机型库(已经回放信号的)
/*
	思路：
	1. 准备报告文件( 不存在即创建)
	2. 创建表头
	3. 写入表内容
	4. 保存文件
*/
func createReportRelateSigReplayDronesDb() {
	logrus.Info("分析报告-关联 机型库(已经回放信号的)")
	// 步骤1: 打开文件
	reportFile, err = createOrOpenExcelFile(reportFilePath)
	errorPanic(err)

	// 步骤2:  创建表头
	// 创建表 - 分析报告
	sheetIndex, _ := reportFile.NewSheet("分析报告-关联机型库(已回放信号)")
	// 设置表为 活动窗口
	reportFile.SetActiveSheet(sheetIndex)
	// 设置列宽
	reportFile.SetColWidth("分析报告-关联机型库(已回放信号)", "A", "P", 15)
	// 写入表头
	tableHeaders := []Any{"ID", "厂家", "品牌", "型号", "协议(drones.csv)", "协议子类型(drones.csv)", "频段",
		"详细频率", "信号文件夹名称(品牌-型号-频段-详细频率)", "信号文件夹路径",
		"信号文件夹路径是否存在", "机型.txt内容", "id.txt内容", "信号文件夹路径重复数量",
		"要查询的机型", "查询结果", "异常原因", "总用时(单位: 分钟)",
		"seafile链接", "信号回放次数"}
	err = reportFile.SetSheetRow("分析报告-关联机型库(已回放信号)", "A1", &tableHeaders)
	errorPanic(err)

	// 步骤3:  写入表内容
	// 把分析报告 - 内容装到map里
	sigPathMap := make(map[string]string)     // 具体机型路径 map key value 类型 key 都是 sigPath，因为它唯一
	queryDroneMap := make(map[string]string)  // 要查询的机型 map, key 都是 sigPath，因为它唯一
	queryResultMap := make(map[string]string) // 查询结果 map, key 都是 sigPath，因为它唯一
	errorReasonMap := make(map[string]string) // 异常原因 map, key 都是 sigPath，因为它唯一
	totalTimeMap := make(map[string]string)   // 总时长 map, key 都是 sigPath，因为它唯一
	// for 循环 读取分析报告
	file, err := createOrOpenExcelFile(reportFilePath)
	errorPanic(err)

	// 获取工作表所有列
	sheetName := "分析报告"
	rows, err := file.Rows(sheetName)
	errorPanic(err)
	logrus.Infof("func=createReportRelateSigReplayDronesDb(), path= %v, sheetName=%v", reportFilePath, sheetName)

	index := 2
	for rows.Next() {
		sigPath, err := file.GetCellValue(sheetName, "B"+strconv.Itoa(index)) // 具体机型路径
		errorPanic(err)
		if sigPath == "" { // 如果不判断发送最后一条空数据时，会报错。因为有 rows.Next()。会获取到最后一条数据，下一行的空数据
			logrus.Info("createReportRelateSigReplayDronesDb(), 最后一条数据，退出")
			break
		}
		sigPathMap[sigPath] = sigPath

		queryDrone, err := file.GetCellValue(sheetName, "C"+strconv.Itoa(index)) // 要查询的机型
		errorPanic(err)
		queryDroneMap[sigPath] = queryDrone

		queryResult, err := file.GetCellValue(sheetName, "D"+strconv.Itoa(index)) // 查询结果
		errorPanic(err)
		queryResultMap[sigPath] = queryResult

		errorReason, err := file.GetCellValue(sheetName, "E"+strconv.Itoa(index)) // 查询结果
		errorPanic(err)
		errorReasonMap[sigPath] = errorReason

		totalTime, err := file.GetCellValue(sheetName, "F"+strconv.Itoa(index)) // 查询结果
		errorPanic(err)
		totalTimeMap[sigPath] = totalTime

		index++
	}

	logrus.Info("sigPathMap = ", sigPathMap)
	// for写入 sheet: 分析报告-关联机型库(已回放信号) 每一行 for dronesDb
	/*
		tableHeaders := []Any{"ID", "厂家", "品牌", "型号", "协议(drones.csv)","协议子类型(drones.csv)" "频段",
		"详细频率", "信号文件夹名称(品牌-型号-频段-详细频率)", "信号文件夹路径",
		"信号文件夹路径是否存在", "机型.txt内容", "id.txt内容", "信号文件夹路径重复数量",
		"要查询的机型", "查询结果", "异常原因", "总用时(单位: 分钟)",
		"seafile链接", "信号重复回放次数"}
	*/
	logrus.Info("dronesDb.SigFolderPath = ", dronesDb.SigFolderPath)
	for index, sigPath := range dronesDb.SigFolderPath {
		// logrus.Infof("写入 sheet: 分析报告-关联机型库(已回放信号) , index= %v, sigPath= %v", index, sigPath)
		tableRow := []Any{dronesDb.Id[index], dronesDb.Manufacture[index], dronesDb.Brand[index], dronesDb.Model[index],
			dronesDb.Protocol[index], dronesDb.Subtype[index], dronesDb.FreqBand[index], dronesDb.Freq[index],
			dronesDb.SigFolderName[index], dronesDb.SigFolderPath[index], dronesDb.SigFolderPathExist[index],
			dronesDb.DroneTxt[index], dronesDb.DroneIdTxt[index], dronesDb.SigFolderPathRepeatNum[index],
			queryDroneMap[sigPath], queryResultMap[sigPath], errorReasonMap[sigPath], totalTimeMap[sigPath],
			dronesDb.SeaFilePath[index], dronesDb.SigFolderReplayNum[index]}
		err = reportFile.SetSheetRow("分析报告-关联机型库(已回放信号)", "A"+strconv.Itoa(index+2), &tableRow)
		errorPanic(err)
	}

	// 步骤4：保存文件
	err = reportFile.SaveAs(reportFilePath)
	errorPanic(err)
}

// 功能：已生成的报告，整合到 机型库里:机型库(最全机型库)
/*
	思路：
	1. 准备报告文件( 不存在即创建)
	2. 对比信息
	3. 写入表内容
*/
func createReportRelateAllDronesDb() {
	logrus.Info("分析报告-关联 机型库(最全机型库)")
	// 步骤1: 打开文件
	reportFile, err = createOrOpenExcelFile(reportFilePath)
	errorPanic(err)

	// 步骤2:  创建表头
	// 创建表 - 分析报告
	reportFile.NewSheet("分析报告-关联机型库(最全机型库)")
	// sheetIndex, _ := reportFile.NewSheet("分析报告-关联机型库(最全机型库)")
	// 设置表为 活动窗口 - 不设置，活动窗口是-回放信号表
	// reportFile.SetActiveSheet(sheetIndex)
	// 设置列宽
	reportFile.SetColWidth("分析报告-关联机型库(最全机型库)", "A", "P", 15)
	// reportFile.SetColWidth("分析报告", "A", "P", 15) // 没有自适应写法
	// 写入表头
	tableHeaders := []Any{"ID", "厂家", "品牌", "型号", "协议(drones.csv)", "协议子类型(drones.csv)", "频段",
		"详细频率", "信号文件夹名称(品牌-型号-频段-详细频率)", "信号文件夹路径",
		"信号文件夹路径是否存在", "机型.txt内容", "id.txt内容", "信号文件夹路径重复数量",
		"要查询的机型", "查询结果", "异常原因", "总用时(单位: 分钟)",
		"seafile链接", "信号重复回放次数"}
	err = reportFile.SetSheetRow("分析报告-关联机型库(最全机型库)", "A1", &tableHeaders)
	errorPanic(err)

	// 步骤3:  写入表内容
	// 把分析报告 - 内容装到map里
	sigPathMap := make(map[string]string)     // 具体机型路径 map key value 类型 key 都是 sigPath，因为它唯一
	queryDroneMap := make(map[string]string)  // 要查询的机型 map, key 都是 sigPath，因为它唯一
	queryResultMap := make(map[string]string) // 查询结果 map, key 都是 sigPath，因为它唯一
	errorReasonMap := make(map[string]string) // 异常原因 map, key 都是 sigPath，因为它唯一
	totalTimeMap := make(map[string]string)   // 总时长 map, key 都是 sigPath，因为它唯一
	// for 循环 读取分析报告
	file, err := createOrOpenExcelFile(reportFilePath)
	errorPanic(err)

	// 获取工作表所有列
	sheetName := "分析报告"
	rows, err := file.Rows(sheetName)
	errorPanic(err)
	logrus.Infof("func=createReportRelateAllDronesDb(), path= %v, sheetName=%v", reportFilePath, sheetName)

	index := 2
	for rows.Next() {
		sigPath, err := file.GetCellValue(sheetName, "B"+strconv.Itoa(index)) // 具体机型路径
		errorPanic(err)
		if sigPath == "" { // 如果不判断发送最后一条空数据时，会报错。因为有 rows.Next()。会获取到最后一条数据，下一行的空数据
			logrus.Info("createReportRelateAllDronesDb(), 最后一条数据，退出")
			break
		}
		sigPathMap[sigPath] = sigPath

		queryDrone, err := file.GetCellValue(sheetName, "C"+strconv.Itoa(index)) // 要查询的机型
		errorPanic(err)
		queryDroneMap[sigPath] = queryDrone

		queryResult, err := file.GetCellValue(sheetName, "D"+strconv.Itoa(index)) // 查询结果
		errorPanic(err)
		queryResultMap[sigPath] = queryResult

		errorReason, err := file.GetCellValue(sheetName, "E"+strconv.Itoa(index)) // 查询结果
		errorPanic(err)
		errorReasonMap[sigPath] = errorReason

		totalTime, err := file.GetCellValue(sheetName, "F"+strconv.Itoa(index)) // 查询结果
		errorPanic(err)
		totalTimeMap[sigPath] = totalTime

		index++
	}

	logrus.Info("sigPathMap = ", sigPathMap)
	// for写入 sheet: 分析报告-关联机型库(已回放信号) 每一行 for allDronesDb
	/*
		tableHeaders := []Any{"ID", "厂家", "品牌", "型号", "协议(drones.csv)","协议子类型(drones.csv)" "频段",
		"详细频率", "信号文件夹名称(品牌-型号-频段-详细频率)", "信号文件夹路径",
		"信号文件夹路径是否存在", "机型.txt内容", "id.txt内容", "信号文件夹路径重复数量",
		"要查询的机型", "查询结果", "异常原因", "总用时(单位: 分钟)"}
	*/
	logrus.Infof("allDronesDb. len(allDronesDb.SigFolderPath)=%v, len(allDronesDb.SeaFilePath=%v)", len(allDronesDb.SigFolderPath), len(allDronesDb.SeaFilePath))
	for index, sigPath := range allDronesDb.SigFolderPath {
		// logrus.Infof("写入 sheet: 分析报告-关联机型库(最全机型库) , index=%v, , len(allDronesDb.SigFolderPath)=%v, allDronesDb.SigFolderPath[index]= %v", index, len(allDronesDb.SeaFilePath), allDronesDb.SeaFilePath[index])
		tableRow := []Any{allDronesDb.Id[index], allDronesDb.Manufacture[index], allDronesDb.Brand[index], allDronesDb.Model[index],
			allDronesDb.Protocol[index], allDronesDb.Subtype[index], allDronesDb.FreqBand[index], allDronesDb.Freq[index],
			allDronesDb.SigFolderName[index], allDronesDb.SigFolderPath[index], allDronesDb.SigFolderPathExist[index],
			allDronesDb.DroneTxt[index], allDronesDb.DroneIdTxt[index], allDronesDb.SigFolderPathRepeatNum[index],
			queryDroneMap[sigPath], queryResultMap[sigPath], errorReasonMap[sigPath], totalTimeMap[sigPath],
			allDronesDb.SeaFilePath[index], allDronesDb.SigFolderReplayNum[index]}
		// logrus.Infof("写入 sheet: 分析报告-关联机型库(最全机型库) , index= %v, sigPath= %v, tableRow= %v", index, sigPath, &tableRow)
		err = reportFile.SetSheetRow("分析报告-关联机型库(最全机型库)", "A"+strconv.Itoa(index+2), &tableRow)
		errorPanic(err)
	}

	// 步骤4：保存文件
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
	droneNameNotEqualNum := 0
	boolResultNoMistakeNum := 0
	boolResultHasMistakeNum := 0
	for index, boolResultDroneNameEqual := range boolResultDroneNameEqualList {
		if boolResultDroneNameEqual == "TRUE" && (boolResultNoMistakeList[index] == "TRUE" || boolResultHasMistakeList[index] == "TRUE") {
			return true, errorReason
		}
		// 用于写 errorReason
		if boolResultDroneNameEqual == "FALSE" {
			droneNameNotEqualNum += 1
		}
		if boolResultNoMistakeList[index] == "FALSE" {
			boolResultNoMistakeNum += 1
		}
		if boolResultHasMistakeList[index] == "FALSE" {
			boolResultHasMistakeNum += 1
		}
	}
	if droneNameNotEqualNum == len(boolResultDroneNameEqualList) {
		errorReason = "未检测到该机型"
	} else if boolResultNoMistakeNum == len(boolResultNoMistakeList) && boolResultHasMistakeNum == len(boolResultHasMistakeList) {
		errorReason = "id不匹配, 或者频率误差<10M"
	} else {
		errorReason = "频率误差<10M"
	}
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
