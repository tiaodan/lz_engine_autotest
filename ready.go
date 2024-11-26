/*
*
功能：处理 ready 阶段逻辑
*/
package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// 初始化
// 声明全局变量 sigpkgList 为字符串切片，preDirName 和 nowDirName 为字符串
var (
	sigpkgList                = make([]string, 0, 1000) // 信号采集列表
	droneObjList              = make([]Drone, 0, 1000)  // 目标飞机列表，与 sigpkgList信号采集列表 一一对应
	preSigDirPath             = ""                      // 之前信号包目录路径
	nowSigDirPath             = ""                      // 现在信号包目录路径
	currentSigDirPath         = ""                      // 当前信号包目录路径
	currentQueryTargetDrone   Drone                     // 当前查询目标飞机,查的是 信号包里的:机型.txt
	currentQueryTargetDroneId = "nil"                   // 当前无人机id
	currentDirSigNum          = 0                       // 当前目录，信号包数量
	currentSigCount           = 0                       // 当前查了几个信号
)

/*
功能：读取配置
参数：
1. 配置文件 名
2. 配置文件 后缀名
3. 配置文件 路径(相对路径)
*/
func readConfig(configName string, configSuffix string, configRelPath string) {
	// 读取配置
	viper.SetConfigName(configName)    // 设置 配置文件名 eg: viper.SetConfigName("config")
	viper.SetConfigType(configSuffix)  // 设置 配置文件后缀名 eg: viper.SetConfigType("ini")
	viper.AddConfigPath(configRelPath) // 设置 配置文件路径 eg: viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	errorPanic(err) // 出了异常就退出

	devIp = viper.GetString("network.devIp")                       // 设备ip
	sigDir = viper.GetString("signal.sigDir")                      // 信号包文件夹路径
	sigPkgSendInterval = viper.GetInt("signal.sigPkgSendInterval") // 发送间隔时间:毫秒/MB,按信号包大小
	cdFolderInterval = viper.GetInt("signal.cdFolderInterval")     // 换文件夹等待时间:秒
	queryDroneInterval = viper.GetInt("signal.queryDroneInterval") // 查询无人机间隔时间:秒

	// 打印配置
	logrus.Debug("配置 devIp (设备ip)= ", devIp)
	logrus.Debug("配置 sigDir (信号包根目录)= ", sigDir)
	logrus.Debug("配置 sigPkgSendInterval (发送间隔时间:毫秒/MB,按信号包大小)= ", sigPkgSendInterval)
	logrus.Debug("配置 cdFolderInterval (换文件夹等待时间:秒)= ", cdFolderInterval)
	logrus.Debug("配置 queryDroneInterval (查询无人机间隔时间:秒)= ", queryDroneInterval)
}

/*
功能：设置变量(全局+局部变量)
参数：
1.

思路：
- 设置 程序开始时间变量 初始值
- 设置 文件相关变量
*/
func setVar() {
	// 设置 程序开始时间变量 初始值
	startTime = time.Now()
	startTimeStr = time2stringforFilename(startTime)

	// 文件相关变量。设置成 文件-时间.xlsx
	preSendHistoryFilePath = "待发送列表-" + startTimeStr + ".xlsx"
	queryHistroyFilePath = "查询列表" + startTimeStr + ".xlsx"
	reportFilePath = "分析报告" + startTimeStr + ".xlsx"
	preSendHistoryFileSheetName = "待发送列表"

	// 打印变量
	logrus.Debug("全局变量 startTime (程序开始时间)= ", startTime)
	logrus.Debug("全局变量 startTimeStr (查程序开始时间str)= ", startTimeStr)
	logrus.Debug("全局变量 preSendHistoryFilePath (待发送信号记录文件路径)= ", preSendHistoryFilePath)
}

/*
功能：创建 待发送信号列表文件。
命名格式：待发送信号列表-20240102-103020.xlsx (名字+年月日-时分秒)
参数：
1. filePath string 文件路径

思路：
1. 创建或者打开文件
2. 创建sheet
3. 设置活动窗口为 新建sheet
4. 设置列宽
5. 创建表头
7. 保存文件
6. 循环读取信号包所有文件，写入行内容
*/
func createPreSendHistoryFile(filePath string) {
	// 1. 创建或者打开文件
	preSendHistoryFile, err = createOrOpenExcelFile(filePath)
	errorPanic(err)

	// 2. 创建sheet
	sheetName := "待发送列表"
	sheetIndex, err := preSendHistoryFile.NewSheet(sheetName)
	errorPanic(err)

	// 3. 设置活动窗口为 新建sheet
	preSendHistoryFile.SetActiveSheet(sheetIndex)

	// 4. 设置列宽
	preSendHistoryFile.SetColWidth(sheetName, "A", "F", 30)

	// 5. 创建表头
	// preSendHistoryFile.SetSheetRow(sheetName, "A1", &[]Any{"厂家", "信号包路径", "要查询的无人机", "待发送信号列表", "是否已发送 TRUE/FALSE"})
	preSendHistoryFile.SetSheetRow(sheetName, "A1", &[]Any{"厂家", "信号包路径", "要查询的无人机", "待发送信号列表"})

	// 7. 保存文件
	err = preSendHistoryFile.SaveAs(preSendHistoryFilePath)
	errorPanic(err)

	// 6. 循环读取信号包所有文件，写入行内容
	err = setPreSendHistoryFileSheetRow(sigDir)
	errorPanic(err)

}

/*
功能：写入《待发送信号列表文件》 row. 遍历所有信号文件，过程中将数据写入 《待发送信号列表》
参数：
1. sigDir string 信号包目录路径

思路：
1. 判断信号包路径能否读取到
2. 从信号包根目录，开始遍历 loopSigPkg()
3. 如果是目录
4. 如果是文件 判断: 机型.txt、id.txt、*.bvsp、*.信号后缀
*/
func setPreSendHistoryFileSheetRow(sigDir string) error {
	// 1. 判断信号包路径能否读取到
	path, err := os.Stat(sigDir)
	if err != nil {
		return err
	}

	// 2. 从信号包根目录，开始遍历 loopSigPkg()
	// 3. 如果是目录
	// 4. 如果是文件 判断: 机型.txt、id.txt、*.bvsp、*.信号后缀
	if path.IsDir() {
		loopDir(sigDir)
	} else {
		loopFile(sigDir)
	}
	logrus.Debug("读取的信号列表sigpkgList= ", sigpkgList)
	return nil
}

/*
功能：遍历目录
参数：
1. dirPath string 目录路径

思路：
1. 读取目录所有内容
2. 排序所有内容
3. 读取每个文件，并写入信息到 待发送列表
*/
func loopDir(dirPath string) {
	// 1. 读取目录所有内容
	files, _ := os.ReadDir(dirPath)
	logrus.Info("---------------------------------------------- 读取目录所有, files = ", files)
	// 2. 排序所有内容
	logrus.Error("------------------------------------- 排序还没写!!!")

	// 3. 读取每个文件，并写入信息到 待发送列表
	for _, file := range files {
		fileAbsPath := filepath.Join(dirPath, file.Name()) // 文件/目录 绝对路径
		// 如果是目录继续处理
		if file.IsDir() {
			logrus.Info("当前信号包路径: filepath.Join= ", fileAbsPath)
			// 如果是最后目录, 就不遍历了当前目录
			if dirIsEndDir(fileAbsPath) && !fileExist(fileAbsPath, "机型.txt") {
				continue
			}
			currentDirSigNum = dirSigNum(fileAbsPath)
			logrus.Debug("当前目录信号数量, currentDirSigNum= ", currentDirSigNum)

			// 1. 判断是否有文件: 机型.txt id.txt文件, 有了单独处理该文件
			if fileExist(fileAbsPath, "机型.txt") {
				logrus.Debug("匹配到 机型.txt, 设置全局变量 currentQueryTargetDrone, path= ", fileAbsPath)
				file, err := os.Open(filepath.Join(fileAbsPath, "机型.txt"))
				if err != nil {
					logrus.Error("无法打开文件 机型.txt:", err)
				}
				defer file.Close()

				contentBytes, err := io.ReadAll(file)
				content := string(contentBytes) // 转成string
				// 替换\r \n内容
				content = strings.ReplaceAll(content, "\r", "")
				content = strings.ReplaceAll(content, "\n", "")
				logrus.Debug("读取文件 机型.txt 内容= ", string(content))
				// 拆分内容
				parts := strings.Split(string(content), ":")

				if err != nil {
					logrus.Error("读取文件 机型.txt 内容出错:", err)
				}
				currentQueryTargetDrone.Name = strings.TrimSpace(parts[0])
				freq, err := strconv.Atoi(parts[1])
				errorPanic(err)
				currentQueryTargetDrone.FreqList = freq
				logrus.Info("匹配到 机型.txt, currentQueryTargetDrone = ", currentQueryTargetDrone)
			}

			// 1. 判断是否有文件:  id.txt文件, 有了单独处理该文件
			if fileExist(fileAbsPath, "id.txt") {
				logrus.Debug("匹配到id.txt, path= ", fileAbsPath)

				file, err := os.Open(filepath.Join(fileAbsPath, "id.txt"))
				if err != nil {
					logrus.Error("无法打开文件 id.txt:", err)
				}
				defer file.Close()

				content, err := io.ReadAll(file)
				logrus.Debug("读取文件id.txt 内容= ", string(content))
				if err != nil {
					logrus.Error("读取文件id.txt 内容出错:", err)
				}
				currentQueryTargetDroneId = strings.TrimSpace(string(content))
				currentQueryTargetDrone.Id = strings.TrimSpace(string(content))
				logrus.Info("匹配到id.txt, currentSigPkgDroneId = ", currentQueryTargetDroneId)
			}
			loopDir(fileAbsPath)
		} else { // 如果是文件，解析它、
			loopFile(fileAbsPath)
		}
	}
}

/*
功能：遍历目录
参数：
1. path string 文件路径

思路：
1. 判断是否有文件: 机型.txt id.txt文件, 有了单独处理该文件
2. 判断是否有文件: 后缀是 dat 或者 bvsp, 有了单独处理该文件
3. 有文件后缀是 dat 或者 bvsp,写入一条行数据; 写完毕后写: [换文件夹]
*/
func loopFile(path string) {
	// ???????什么时候写入一条数据 - loopDir 要一个标准位，判断是最后一个文件，才修改current 变量

	// 获取文件扩展名
	logrus.Debug("当前 sig = ", path)
	if fileExistExt(path, ".dat") || fileExistExt(path, ".bvsp") {
		nowSigDirPath = filepath.Dir(path)
		logrus.Debug("preSigDirPath= ", preSigDirPath)
		logrus.Debug("nowSigDirPath= ", nowSigDirPath)
		if preSigDirPath == "" {
			preSigDirPath = nowSigDirPath
			sigpkgList = append(sigpkgList, path)
			logrus.Debug("preSigDirPath == 空,判断了1次？？？")
			currentSigCount++ // 计数+1
		} else if preSigDirPath == nowSigDirPath {
			sigpkgList = append(sigpkgList, path)
			currentSigCount++ // 计数+1
			// 设置 当前信号包目录
			if currentSigDirPath != nowSigDirPath {
				currentSigDirPath = nowSigDirPath
			}
			logrus.Debug("preSigDirPath == nowSigDirPat,判断了几次？？？=信号数量")
			logrus.Debug("currentSigCount == ", currentSigCount)

			// 判断是否是最后一个文件
			if currentSigCount == currentDirSigNum {
				// 尝试在这里写入excel表，并重置 sigpkgList为空
				logrus.Info("loopFile 到最后一个信号文件")
				sigpkgList = append(sigpkgList, "[换文件夹]")
				logrus.Infof("写入待发送信号列表, 厂家=%v, 信号包路径=%v, 要查询的无人机=%v, 待发送信号列表=%v", sigDir, currentSigDirPath, currentQueryTargetDrone, sigpkgList)
				writeExcelTableRowByArgs(sigDir, currentSigDirPath, currentQueryTargetDrone, sigpkgList)
				sigpkgList = make([]string, 0) // 重置信号列表为空
			}
		} else {
			logrus.Debug("------------- 某个信号目录已经遍历完了,已经切到另一个目录")
			preSigDirPath = nowSigDirPath
			// sigpkgList = append(sigpkgList, "[换文件夹]")
			// sigpkgList = make([]string, 0)        // 重置信号列表为空
			currentSigCount = 0                   // 重置
			sigpkgList = append(sigpkgList, path) // 加入新目录的第一个信号
			currentSigCount++
		}
	}

}

/*
功能：当前目录下信号包数量。类型为 .dat .bvsp 认为是信号包
参数：
1. dirPath 目录路径, 建议绝对路径

返回值：
1. int 信号包数量
*/
func dirSigNum(dirPath string) int {
	num := 0
	files, err := os.ReadDir(dirPath)
	if err != nil {
		logrus.Error("func=dirSigNum(), 目录不存在, Error reading directory: ", err)
		return 0
	}

	absFilePath := ""
	for _, file := range files {
		absFilePath = filepath.Join(dirPath, file.Name())
		if fileExistExt(absFilePath, ".dat") || fileExistExt(absFilePath, ".bvsp") {
			num++
		}
	}
	return num

}

/*
功能：根据参数, 写入tablerow 到excel表
参数：
--- 1. excelObj excel文件对象  参数弃用
--- 2. sheetName 表名  名字=待发送列表, 参数弃用
3. &[]Any{"厂家", "信号包路径", "要查询的无人机", "待发送信号列表", "是否已发送 TRUE/FALSE"}
1. sigDir  厂家
2. currentSigDirPath  信号包路径
3. currentQueryTargetDrone  要查询的无人机
4. sigpkgList  待发送信号列表

返回值：
无
*/
func writeExcelTableRowByArgs(sigDir string, currentSigDirPath string, currentQueryTargetDrone Drone, sigpkgList []string) {
	// writeExcelTableRowByArgs(preSendHistoryFile, "待发送列表", &[]Any{sigDir, currentSigDirPath, currentQueryTargetDrone, sigpkgList})

	// 1. 创建或者打开文件
	preSendHistoryFile, err = createOrOpenExcelFile(preSendHistoryFilePath)
	errorPanic(err)

	// 2. 创建sheet
	sheetName := "待发送列表"

	// 3. 设置活动窗口为
	// preSendHistoryFile.SetActiveSheet(sheetIndex)

	logrus.Debug("-----------------------------------------------写入excel前，当前飞机=", currentQueryTargetDrone)
	for index, sig := range sigpkgList {
		// 判断当前有多少行
		rowCount, err := getExcelRowsCount(preSendHistoryFilePath, sheetName)

		errorPanic(err)
		logrus.Debugf("写入待发送列表.xlsx, rowCount=%v, index=%v, sig= %v, drone=%v", rowCount, index, sig, currentQueryTargetDrone)
		// 5. 写入
		// currentQueryTargetDrone 对象 转成json格式
		currentQueryTargetDroneJsonObj, err := json.Marshal(currentQueryTargetDrone)
		errorPanic(err)
		logrus.Debug("-----------------------------------------------写入excel前，当前飞机jsonObj=", string(currentQueryTargetDroneJsonObj))
		// 写入飞机 是jsonStr格式
		preSendHistoryFile.SetSheetRow(sheetName, "A"+strconv.Itoa(rowCount+1+index), &[]Any{sigDir, currentSigDirPath, string(currentQueryTargetDroneJsonObj), sig})
	}
	// 7. 保存文件
	err = preSendHistoryFile.SaveAs(preSendHistoryFilePath)
	errorPanic(err)
}
