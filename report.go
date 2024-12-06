/**
 * 功能：生成报告
 */
package main

import (
	"strconv"

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
	tableHeaders := []Any{"厂家", "具体机型路径", "要查询的机型", "查询结果", "异常原因", "总用时"}
	err = reportFile.SetSheetRow("分析报告", "A1", &tableHeaders)
	errorPanic(err)

	// 2. 判断查询列表 最后2列结果 (有误差的结果 + 无误差的结果)
	// 查询最后2列
	// ?是否要openFile? - 不用
	rows, err := queryHistroyFile.GetRows("Sheet1")
	errorPanic(err)
	var boolResultNoMistakeList []string
	var boolResultHasMistakeList []string
	currentDroneStr := ""    // 用于区分？？，怎么说
	writeReportRowIndex := 2 // 写入报告表index
	// for index := 1; index <= len(rows); index++ {
	for index, row := range rows { // 不好处理的写法, index 和row同步？
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
		}
		// currentDroneStr 不一样的时候，写入报告一条数据, 或者是最后一条数据时
		if currentDroneStr != row[0] || index+1 == len(rows) {
			logrus.Debug("boolResultNoMistakeList = ", boolResultNoMistakeList)
			logrus.Debug("boolResultNoMistakeList = ", boolResultHasMistakeList)
			oneSigReportResult := checkAlgorithmWhereQueryResult(boolResultNoMistakeList, boolResultHasMistakeList)
			logrus.Infof("report 单个信号包,结果。currentSigFolderDir=%v, currentDroneStr=%v, oneSigReportResult =%v ", currentSigFolderDir, currentDroneStr, oneSigReportResult)

			// 写入行内容
			tableRow := []Any{"厂家??", currentSigFolderDir, currentDroneStr, oneSigReportResult, "异常原因??"}
			logrus.Debug("---------------- 打算写入文件,currentDroneStr", currentDroneStr)
			logrus.Debug("---------------- 打算写入文件", &tableRow)
			err = reportFile.SetSheetRow("分析报告", "A"+strconv.Itoa(writeReportRowIndex), &tableRow)
			errorPanic(err)

			// 写入后, 重置变量
			boolResultNoMistakeList = []string{}
			boolResultHasMistakeList = []string{}
			currentDroneStr = row[0]
			boolResultNoMistakeList = append(boolResultNoMistakeList, row[3])
			boolResultHasMistakeList = append(boolResultHasMistakeList, row[2])
			writeReportRowIndex++
		}
	}

	// 3. 写入表内容
	err = reportFile.SaveAs(reportFilePath)
	errorPanic(err)

}

// 处理查询结果的算法 - 写到 报告的表里
// 算法:
// - 步骤1:无误差结果,只要有TRUE (string类型),就返回true
// - 步骤2:有误差结果,只要有TRUE (string类型),就返回true
// 参数: 1 没有误差的bool列表 2 有误差的结果bool列表
func checkAlgorithmWhereQueryResult(boolResultNoMistakeList []string, boolResultHasMistakeList []string) bool {
	logrus.Info("report检测算法, 无误差结果 boolResultNoMistakeList= ", boolResultNoMistakeList)
	logrus.Info("report检测算法, 有误差结果 boolResultHasMistakeList= ", boolResultHasMistakeList)
	// 先判断 id相等 & 频率相等的 结果
	for _, boolResultNoMistake := range boolResultNoMistakeList {
		// 只要有TRUE (string类型),就返回true
		if boolResultNoMistake == "TRUE" {
			return true
		}
	}

	// 再判断 id相等 & 频率有误差的 结果
	for _, boolResultHasMistake := range boolResultHasMistakeList {
		// 只要有TRUE (string类型),就返回true
		if boolResultHasMistake == "TRUE" {
			return true
		}
	}

	return false
}
