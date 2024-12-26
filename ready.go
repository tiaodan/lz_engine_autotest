/*
*
功能：处理 ready 阶段逻辑
*/
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// 初始化
// 声明全局变量 sigpkgList 为字符串切片，preDirName 和 nowDirName 为字符串
var (
	sigpkgList        = make([]string, 0, 1000)  // 信号采集列表
	droneObjList      = make([][]Drone, 0, 1000) // 目标飞机列表，与 sigpkgList信号采集列表 一一对应
	sigFolderPathList = make([]string, 0, 1000)  // 信号文件夹列表，与 sigpkgList信号采集列表 一一对应
	preSigDirPath     = ""                       // 之前信号包目录路径
	nowSigDirPath     = ""                       // 现在信号包目录路径
	currentSigDirPath = ""                       // 当前信号包目录路径
	// currentQueryTargetDroneList []Drone                   // 当前查询目标飞机列表,查的是 信号包里的:机型.txt。如果机型.txt 有多行，没行算一个飞机
	currentQueryTargetDrone    []Drone      // 当前查询目标飞机,查的是 信号包里的:机型.txt
	currentQueryTargetDroneIds = []string{} // 当前无人机id 数组列表, 可以是多个
	currentDirSigNum           = 0          // 当前目录，信号包数量
	currentSigCount            = 0          // 当前查了几个信号
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
	logLevel = viper.GetString("log.logLevel")                     // 日志级别 只认：debug 、info 、 error，不区分大小写。写其它的都按debug处理
	dronesDbEnable = viper.GetBool("dronesdb.dronesdbenable")      // 是否启用机型库
	dronesDbPath = viper.GetString("dronesdb.dronesdbpath")        // 机型库路径

	// 打印配置
	logrus.Info("配置 devIp (设备ip)= ", devIp)
	logrus.Info("配置 sigDir (信号包根目录)= ", sigDir)
	logrus.Info("配置 sigPkgSendInterval (发送间隔时间:毫秒/MB,按信号包大小)= ", sigPkgSendInterval)
	logrus.Info("配置 cdFolderInterval (换文件夹等待时间:秒)= ", cdFolderInterval)
	logrus.Info("配置 queryDroneInterval (查询无人机间隔时间:秒)= ", queryDroneInterval)
	logrus.Info("配置 logLevel (日志级别)= ", logLevel)
	logrus.Info("配置 dronesdbenable (是否启用机型库)= ", dronesDbEnable)
	logrus.Info("配置 dronesdbenable (是否启用机型库)= ", viper.GetString("dronesdb.dronesdbenable"))
	logrus.Info("配置 dronesDbPath (机型库路径)= ", dronesDbPath)
}

/*
功能：读取配置 - 小写的配置文件 (此文件为vippe.WriteConfig() 生成的)
参数：
1. 配置文件 名
2. 配置文件 后缀名
3. 配置文件 路径(相对路径)
*/
func readLowerConfig(configName string, configSuffix string, configRelPath string) {
	// 读取配置
	viper.SetConfigName(configName)    // 设置 配置文件名 eg: viper.SetConfigName("config")
	viper.SetConfigType(configSuffix)  // 设置 配置文件后缀名 eg: viper.SetConfigType("ini")
	viper.AddConfigPath(configRelPath) // 设置 配置文件路径 eg: viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	errorPanic(err) // 出了异常就退出

	devIp = viper.GetString("network.devip")                       // 设备ip
	sigDir = viper.GetString("signal.sigdir")                      // 信号包文件夹路径
	sigPkgSendInterval = viper.GetInt("signal.sigpkgsendinterval") // 发送间隔时间:毫秒/MB,按信号包大小
	cdFolderInterval = viper.GetInt("signal.cdfolderinterval")     // 换文件夹等待时间:秒
	queryDroneInterval = viper.GetInt("signal.querydroneinterval") // 查询无人机间隔时间:秒
	logLevel = viper.GetString("log.loglevel")                     // 日志级别 只认：debug 、info 、 error，不区分大小写。写其它的都按debug处理
	startTimeStr = viper.GetString("time.starttime")               // 开始时间str
	mistakeFreqConfig = viper.GetInt("query.mistakefreq")          // 查询无人机频率 最大误差值 单位：Mhz
	dronesDbEnable = viper.GetBool("dronesdb.dronesdbenable")      // 是否使用机型库，进行自动化测试
	dronesDbPath = viper.GetString("dronesdb.dronesdbpath")        // 机型库路径，一般用户回放部分筛选信号
	allDronesDbPath = viper.GetString("dronesdb.alldronesdbpath")  // all机型库路径

	// 读取配置开始时间
	preSendHistoryFilePath = viper.GetString("file.presendhistoryfilepath")
	preSendHistoryFileSheetName = "待发送列表"
	preSendHistoryFileTxtPath = "待发送列表-" + startTimeStr + ".txt" // 预发送记录文件txt 路径
	fmt.Println("startTimeStr= ", startTimeStr)

	fmt.Println("preSendHistoryFilePath= ", preSendHistoryFilePath)
	queryHistroyFilePath = viper.GetString("file.queryhistroyfilePath") // 查询文件路径
	queryHistroyFileTxtPath = "查询列表" + startTimeStr + ".txt"            // 查询记录文件txt 路径
	reportFilePath = viper.GetString("file.reportfilepath")             // 查询文件路径

	// 打印配置
	logrus.Info("配置 devIp (设备ip)= ", devIp)
	logrus.Info("配置 sigDir (信号包根目录)= ", sigDir)
	logrus.Info("配置 sigPkgSendInterval (发送间隔时间:毫秒/MB,按信号包大小)= ", sigPkgSendInterval)
	logrus.Info("配置 cdFolderInterval (换文件夹等待时间:秒)= ", cdFolderInterval)
	logrus.Info("配置 queryDroneInterval (查询无人机间隔时间:秒)= ", queryDroneInterval)
	logrus.Info("配置 logLevel (日志级别)= ", logLevel)
	logrus.Info("配置 开始时间str = ", startTimeStr)
	logrus.Info("配置 dronesDbEnable (是否启用机型库)= ", dronesDbEnable)
	logrus.Info("配置 dronesDbEnable (是否启用机型库)= ", viper.GetString("dronesdb.dronesdbenable"))
	logrus.Info("配置 dronesDbPath (回放信号机型库路径)= ", dronesDbPath)
	logrus.Info("配置 allDronesDbPath (机型库路径)= ", allDronesDbPath)
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

	// 配置日志相关
	// 配置日志等级
	var logrusLevel logrus.Level
	if strings.EqualFold(logLevel, "debug") {
		logrusLevel = logrus.DebugLevel
	} else if strings.EqualFold(logLevel, "info") {
		logrusLevel = logrus.InfoLevel
	} else if strings.EqualFold(logLevel, "error") {
		logrusLevel = logrus.ErrorLevel
	} else {
		logrusLevel = logrus.DebugLevel
	}

	logrus.SetLevel(logrusLevel)

	changeFolderFlag = false // 换文件夹标志 = false
	changeFolderFlagNum = 0

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
	// 不启用机型库 逻辑
	if !dronesDbEnable {
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

	// 启用机型库 逻辑
	if dronesDbEnable {
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
}

/*
功能：创建 待发送信号列表文件-txt 表头,方便用户查看，因为打开excel文件，程序会报错。
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
func createPreSendHistoryFileHeaderTxt(filePath string) {
	logrus.Info("创建预发送txt文件表头")
	// 1. 创建或者打开文件
	preSendHistoryTxtFile, err = createOrOpenTxtFile(filePath)
	errorPanic(err)

	// 5. 创建表头
	_, err = preSendHistoryTxtFile.WriteString("厂家, 信号包路径, 要查询的无人机, 待发送信号列表 \n")
	if err != nil {
		logrus.Error("写入txt文件失败, err=", err)
	}

	// 关闭文件
	preSendHistoryTxtFile.Close()
}

/*
功能：创建 待发送信号列表文件-txt,方便用户查看，因为打开excel文件，程序会报错。
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
func createPreSendHistoryFileTxt(filePath string) {
	logrus.Info("创建预发送txt文件")
	// 1. 创建或者打开文件
	preSendHistoryTxtFile, err = createOrOpenTxtFile(filePath)
	errorPanic(err)

	// 7. 保存文件

	// 6. 循环读取信号包所有文件，写入行内容
	// err = setPreSendHistoryFileSheetRow(sigDir)
	// errorPanic(err)

	// 关闭文件
	preSendHistoryTxtFile.Close()
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

	// 不启用机型库配置 - 默认逻辑
	if !dronesDbEnable {
		logrus.Info("setPreSendHistoryFileSheetRow(), 参数: sigDir = ", sigDir)
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

	if dronesDbEnable {
		logrus.Info("setPreSendHistoryFileSheetRow(), 参数: sigDir = ", sigDir)
		// 1. 判断信号包路径能否读取到
		// path, err := os.Lstat(sigDir)
		fileInfo, err := os.Lstat(filepath.Join(sigDir))
		if err != nil {
			return err
		}

		// 2. 从信号包根目录，开始遍历 loopSigPkg()
		// 3. 如果是目录
		// 4. 如果是文件 判断: 机型.txt、id.txt、*.bvsp、*.信号后缀
		if fileInfo.IsDir() || (fileInfo.Mode()&os.ModeSymlink != 0) {
			loopDir(sigDir)
		} else {
			loopFile(sigDir)
		}
		logrus.Debug("读取的信号列表sigpkgList= ", sigpkgList)
		return nil
	}
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
	// 不启用机型库的写法
	if !dronesDbEnable {
		// 1. 读取目录所有内容
		files, _ := os.ReadDir(dirPath)
		logrus.Info("---------------------------------------------- 读取目录所有, files = ", files)
		// 2. 排序所有内容
		// logrus.Error("------------------------------------- 排序还没写!!!")

		// 3. 读取每个文件，并写入信息到 待发送列表
		for _, file := range files {
			fileAbsPath := filepath.Join(dirPath, file.Name()) // 文件/目录 绝对路径
			// 如果是目录继续处理
			if file.IsDir() {
				logrus.Debug("当前信号包路径: filepath.Join= ", fileAbsPath)
				// 如果是最后目录, 就不遍历了当前目录
				if dirIsEndDir(fileAbsPath) && !fileExist(fileAbsPath, "机型.txt") {
					continue
				}
				currentDirSigNum = dirSigNum(fileAbsPath)
				logrus.Debug("当前目录信号数量, currentDirSigNum= ", currentDirSigNum)

				// 1. 判断是否有文件: 机型.txt id.txt文件, 有了单独处理该文件
				/*
					// 机型.txt只有一行的写法
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
						currentQueryTargetDroneList = append(currentQueryTargetDroneList, currentQueryTargetDrone)
					}
				*/
				// 机型.txt多行的写法
				if fileExist(fileAbsPath, "机型.txt") {
					logrus.Debug("匹配到 机型.txt, 设置全局变量 currentQueryTargetDrone, path= ", fileAbsPath)

					file, err := os.Open(filepath.Join(fileAbsPath, "机型.txt"))
					if err != nil {
						log.Fatalf("无法打开文件 机型.txt: %v", err)
					}
					defer file.Close()

					scanner := bufio.NewScanner(file)

					for scanner.Scan() {
						line := scanner.Text()
						line = strings.ReplaceAll(line, "\r", "")
						line = strings.ReplaceAll(line, "\n", "")

						parts := strings.Split(line, ":")

						if len(parts) != 2 {
							log.Printf("文件行格式错误: %s", line)
							continue
						}

						// drone := Drone{
						// 	Name: strings.TrimSpace(parts[0]),
						// }
						drone := Drone{}
						drone.Name = strings.TrimSpace(parts[0])

						freq, err := strconv.Atoi(parts[1])
						if err != nil {
							log.Printf("无法解析频率: %v", err)
							continue
						}
						drone.FreqList = freq

						logrus.Info("读取文件 机型.txt 内容 = ", line)
						logrus.Info("匹配到 机型.txt, currentQueryTargetDrone = ", drone)
						currentQueryTargetDrone = append(currentQueryTargetDrone, drone)
						logrus.Info("匹配到 机型.txt, currentQueryTargetDroneList = ", currentQueryTargetDrone)
					}

					if err := scanner.Err(); err != nil {
						log.Fatalf("读取文件 机型.txt 内容出错: %v", err)
					}
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
					// 读取id.txt 多个id
					// 使用 Split 函数按照 "/" 分割字符串
					contentTrimSpace := strings.TrimSpace(string(content))             // 去除前后空格
					currentQueryTargetDroneIds := strings.Split(contentTrimSpace, "/") // 通过/ 分割

					// 如果切片的长度为1，说明原来的字符串中没有 "/"
					if len(currentQueryTargetDroneIds) == 1 {
						// 将 content 添加到切片的第一个元素位置
						currentQueryTargetDroneIds = []string{contentTrimSpace}
					}

					for i := range currentQueryTargetDrone {
						// drone.Id = currentQueryTargetDroneIds  // 这样写，赋值不过去
						currentQueryTargetDrone[i].Id = currentQueryTargetDroneIds // 这样写，能赋值过去 why?
					}
					logrus.Info("匹配到id.txt, currentSigPkgDroneId = ", currentQueryTargetDroneIds)
					logrus.Info("currentQueryTargetDroneList,添加ids后 = ", currentQueryTargetDrone)
				}
				loopDir(fileAbsPath)
			} else { // 如果是文件，解析它、
				loopFile(fileAbsPath)
			}
		}
	}

	// 启用机型库的写法
	if dronesDbEnable {
		// 1. 读取目录所有内容
		files, _ := os.ReadDir(dirPath)
		logrus.Infof("---------------------------------------------- loopDir(), 读取目录%v所有, files = %v", dirPath, files)
		// 2. 排序所有内容
		// logrus.Error("------------------------------------- 排序还没写!!!")

		// 3. 读取每个文件，并写入信息到 待发送列表
		for _, file := range files {
			fileAbsPath := filepath.Join(dirPath, file.Name()) // 文件/目录 绝对路径
			fileInfo, err := os.Lstat(filepath.Join(dirPath, file.Name()))
			if err != nil {
				fmt.Println("获取文件信息出错:", err)
				continue
			}

			// 如果是目录继续处理
			if fileInfo.IsDir() || (fileInfo.Mode()&os.ModeSymlink != 0) { // 链接文件夹，或者普通文件夹
				if fileInfo.Mode()&os.ModeSymlink != 0 { // 读取链接，真实路径
					fileAbsPath, err = os.Readlink(fileAbsPath)
					if err != nil {
						fmt.Println("获取真实路径出错:", err)
						return
					}
				}
				logrus.Info("---------- loopDir(), 进入逻辑:  fileInfo.IsDir() || (fileInfo.Mode()&os.ModeSymlink != 0) ")
				logrus.Info("当前信号包路径: filepath.Join= ", fileAbsPath)
				// 如果是最后目录, 就不遍历了当前目录
				logrus.Info("---------- dirIsEndDir(fileAbsPath) = ", dirIsEndDir(fileAbsPath))
				logrus.Info("---------- !fileExist(fileAbsPath, 机型.txt) = ", !fileExist(fileAbsPath, "机型.txt"))
				if dirIsEndDir(fileAbsPath) && !fileExist(fileAbsPath, "机型.txt") {
					logrus.Info("---------- loopDir(), 进入逻辑:   dirIsEndDir(fileAbsPath) && !fileExist(fileAbsPath, 跳过当前循环")
					continue
				}
				currentDirSigNum = dirSigNum(fileAbsPath)
				logrus.Debug("当前目录信号数量, currentDirSigNum= ", currentDirSigNum)

				// 1. 判断是否有文件: 机型.txt id.txt文件, 有了单独处理该文件
				/*
					// 机型.txt只有一行的写法
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
						currentQueryTargetDroneList = append(currentQueryTargetDroneList, currentQueryTargetDrone)
					}
				*/
				// 机型.txt多行的写法
				if fileExist(fileAbsPath, "机型.txt") {
					logrus.Debug("匹配到 机型.txt, 设置全局变量 currentQueryTargetDrone, path= ", fileAbsPath)

					file, err := os.Open(filepath.Join(fileAbsPath, "机型.txt"))
					if err != nil {
						log.Fatalf("无法打开文件 机型.txt: %v", err)
					}
					defer file.Close()

					scanner := bufio.NewScanner(file)

					for scanner.Scan() {
						line := scanner.Text()
						line = strings.ReplaceAll(line, "\r", "")
						line = strings.ReplaceAll(line, "\n", "")

						parts := strings.Split(line, ":")

						if len(parts) != 2 {
							log.Printf("文件行格式错误: %s", line)
							continue
						}

						// drone := Drone{
						// 	Name: strings.TrimSpace(parts[0]),
						// }
						drone := Drone{}
						drone.Name = strings.TrimSpace(parts[0])

						freq, err := strconv.Atoi(parts[1])
						if err != nil {
							log.Printf("无法解析频率: %v", err)
							continue
						}
						drone.FreqList = freq

						logrus.Info("读取文件 机型.txt 内容 = ", line)
						logrus.Info("匹配到 机型.txt, currentQueryTargetDrone = ", drone)
						currentQueryTargetDrone = append(currentQueryTargetDrone, drone)
						logrus.Info("匹配到 机型.txt, currentQueryTargetDroneList = ", currentQueryTargetDrone)
					}

					if err := scanner.Err(); err != nil {
						log.Fatalf("读取文件 机型.txt 内容出错: %v", err)
					}
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
					// 读取id.txt 多个id
					// 使用 Split 函数按照 "/" 分割字符串
					contentTrimSpace := strings.TrimSpace(string(content))             // 去除前后空格
					currentQueryTargetDroneIds := strings.Split(contentTrimSpace, "/") // 通过/ 分割

					// 如果切片的长度为1，说明原来的字符串中没有 "/"
					if len(currentQueryTargetDroneIds) == 1 {
						// 将 content 添加到切片的第一个元素位置
						currentQueryTargetDroneIds = []string{contentTrimSpace}
					}

					for i := range currentQueryTargetDrone {
						// drone.Id = currentQueryTargetDroneIds  // 这样写，赋值不过去
						currentQueryTargetDrone[i].Id = currentQueryTargetDroneIds // 这样写，能赋值过去 why?
					}
					logrus.Info("匹配到id.txt, currentSigPkgDroneId = ", currentQueryTargetDroneIds)
					logrus.Info("currentQueryTargetDroneList,添加ids后 = ", currentQueryTargetDrone)
				}
				loopDir(fileAbsPath)
			} else { // 如果是文件，解析它、
				loopFile(fileAbsPath)
			}
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
	logrus.Info("fileExistExt(path, .bvsp) = ", fileExistExt(path, ".bvsp"))
	if fileExistExt(path, ".dat") || fileExistExt(path, ".bvsp") {
		nowSigDirPath = filepath.Dir(path)
		logrus.Info("preSigDirPath= ", preSigDirPath)
		logrus.Info("nowSigDirPath= ", nowSigDirPath)
		if preSigDirPath == "" {
			preSigDirPath = nowSigDirPath
			sigpkgList = append(sigpkgList, path)
			logrus.Info("preSigDirPath == 空,判断了1次？？？")
			currentSigCount++ // 计数+1
		} else if preSigDirPath == nowSigDirPath {
			sigpkgList = append(sigpkgList, path)
			currentSigCount++ // 计数+1
			// 设置 当前信号包目录
			if currentSigDirPath != nowSigDirPath {
				currentSigDirPath = nowSigDirPath
			}
			logrus.Debug("preSigDirPath == nowSigDirPat,判断了几次？？？=信号数量")
			logrus.Info("currentSigCount == ", currentSigCount)
			logrus.Info("currentDirSigNum == ", currentDirSigNum)

			// 判断是否是最后一个文件
			if currentSigCount == currentDirSigNum {
				// 尝试在这里写入excel表，并重置 sigpkgList为空
				// 给每个信号文件夹,所有信号 = sigpkgList 排序，排序后，再加上  [换文件夹]
				sortStringArr(sigpkgList)
				logrus.Info("排序后的 sigpkgList= ", sigpkgList)

				logrus.Info("loopFile 到最后一个信号文件")
				// 根据信号回放次数，把sigpkgList 循环添加几次
				sigFolderReplayNum, err := strconv.Atoi(sigFolderReplayNumMap[currentSigDirPath])
				errorEcho(err)
				logrus.Info("回放次数，sigFolderReplayNum= ", sigFolderReplayNum)

				sigpkgListOneFolder := sigpkgList
				for range sigFolderReplayNum {
					sigpkgList = append(sigpkgList, sigpkgListOneFolder...)
				}

				sigpkgList = append(sigpkgList, "[换文件夹]")
				logrus.Infof("写入待发送信号列表, 厂家=%v, 信号包路径=%v, 要查询的无人机=%v, 待发送信号列表=%v", sigDir, currentSigDirPath, currentQueryTargetDrone, sigpkgList)
				// writeExcelTableRowByArgs(sigDir, currentSigDirPath, currentQueryTargetDrone, sigpkgList)  // 原来的写法- sigpkgList
				writeExcelTableRowByArgs(sigDir, currentSigDirPath, currentQueryTargetDrone, sigpkgList) // v0.0.0.1的写法- sigpkgList
				logrus.Info("要写入txt siglist= ", sigpkgList)
				// writeTxtRowByArgs(sigDir, currentSigDirPath, currentQueryTargetDrone, sigpkgList) // v0.0.0.1的写法- sigpkgList
				sigpkgList = make([]string, 0)      // 重置信号列表为空
				currentQueryTargetDrone = []Drone{} // 重置查询飞机列表为空
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
3. currentQueryTargetDrone  要查询的无人机 类型: []Drone
4. sigpkgList  待发送信号列表

返回值：
无
*/
func writeExcelTableRowByArgs(sigDir string, currentSigDirPath string, currentQueryTargetDrone []Drone, sigpkgList []string) {
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

/*
功能：根据参数, 写入tablerow 到txt表
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
func writeTxtRowByArgs(sigDir string, currentSigDirPath string, currentQueryTargetDrone Drone, sigpkgList []string) {
	// writeExcelTableRowByArgs(preSendHistoryFile, "待发送列表", &[]Any{sigDir, currentSigDirPath, currentQueryTargetDrone, sigpkgList})

	// 1. 创建或者打开文件
	preSendHistoryTxtFile, err = createOrOpenTxtFile(preSendHistoryFileTxtPath)
	errorPanic(err)

	logrus.Debug("-----------------------------------------------写入待发送列表 txt前，当前飞机=", currentQueryTargetDrone)
	for index, sig := range sigpkgList {
		logrus.Debugf("写入待发送列表.txt, index=%v, sig= %v, drone=%v", index, sig, currentQueryTargetDrone)
		// 5. 写入
		rowStr := fmt.Sprintf("%v, %v, %v, %v \n", sigDir, currentSigDirPath, currentQueryTargetDrone, sig)
		_, err = preSendHistoryTxtFile.WriteString(rowStr + "\n")
		if err != nil {
			logrus.Error("写入查询txt文件row失败, err=", err)
		}
	}
	// 7. 保存文件
	preSendHistoryTxtFile.Close()
}

/*
功能：升序排列 string数组 - 不好用
参数：
--- 1. []string 信号列表

返回值：
无
*/
func sortStringArr(sigpkgList []string) {
	sort.Slice(sigpkgList, func(i, j int) bool {
		// num1, _ := strconv.Atoi(sigpkgList[i][len(sigpkgList[i])-7 : len(sigpkgList[i])-5])
		// num2, _ := strconv.Atoi(sigpkgList[j][len(sigpkgList[j])-7 : len(sigpkgList[j])-5])
		num1 := regOneSigNum(sigpkgList[i])
		num2 := regOneSigNum(sigpkgList[j])
		return num1 < num2
	})
	for _, val := range sigpkgList {
		fmt.Println(val)
	}
}

/*
功能：提取一个信号的 数字.bvsp
例子："E:\\xinhao\\AEE_Sparrow2_5745\\2.bvsp"
参数：
--- 1. oneSigPath string 信号 "E:\\xinhao\\AEE_Sparrow2_5745\\2.bvsp"

返回值：
num
*/
func regOneSigNum(oneSigPath string) int {
	// fmt.Println("---------------------- oneSigPath= ", oneSigPath)
	re := regexp.MustCompile(`(\d+)\.(bvsp|dat)$`)
	match := re.FindStringSubmatch(oneSigPath)

	num := 0
	if len(match) > 1 {
		num, err = strconv.Atoi(match[1])
		if err != nil {
			fmt.Println("正则匹配oneSig No match found.")
			num = 0
		}
	}
	// fmt.Println(num)
	return num
}

/*
功能：检测机型库 dronesDb 对象，完善该对象数据。并把判断结果（信号路径重复项数量、信号文件夹路径是否存在）写入 机型库.xlsx文件中
参数：
1. dronesDB 对象 不用参数,直接修改全局变量 dronesDB

返回值：
nil
*/
func checkDronesDbAndWrite2Excel() {
	// 步骤1. 准备文件( 不存在即创建)
	file, err := createOrOpenExcelFile(dronesDbPath)
	errorPanic(err)

	// 创建一个哈希表用于存储数组A中每个元素的出现次数,初始化
	pathNumMap := make(map[string]int)
	for _, path := range dronesDb.SigFolderPath {
		pathNumMap[path] = 0 // 初始化 = 0
	}

	for index, path := range dronesDb.SigFolderPath {
		// 步骤2: 判断 信号文件夹路径是否存在，更新 dronesDb 对象
		if checkPathExist(path) {
			dronesDb.SigFolderPathExist = append(dronesDb.SigFolderPathExist, true)
		} else {
			dronesDb.SigFolderPathExist = append(dronesDb.SigFolderPathExist, true)
		}

		// 步骤3: 把 信号文件夹路径是否存在，写入excel
		// 写入内容
		row := dronesDb.SigFolderPathExist[index]
		file.SetCellValue("机型库", "K"+strconv.Itoa(index+2), row) // 从第2行开始写入。 &row 会写进入 内存地址，,不对
		errorPanic(err)

		// 步骤4: 判断 信号路径重复项数量，更新 dronesDb 对象
		// 如果path 在map的key里
		if checkStringInMap(pathNumMap, path) {
			pathNumMap[path]++
		}
		dronesDb.SigFolderPathRepeatNum = append(dronesDb.SigFolderPathRepeatNum, pathNumMap[path])

		// 步骤5: 把 信号路径重复项数量, 写入excel
		file.SetCellValue("机型库", "N"+strconv.Itoa(index+2), dronesDb.SigFolderPathRepeatNum[index]) // 从第2行开始写入。 &row 会写进入 内存地址，,不对
		errorPanic(err)
	}

	// 步骤4. 保存生效
	err = file.SaveAs(dronesDbPath)
	errorPanic(err)

}

/*
功能：检测机型库 allDronesDb 对象，完善该对象数据。并把判断结果（信号路径重复项数量、信号文件夹路径是否存在）写入 机型库.xlsx文件中
参数：
1. allDronesDb 对象 不用参数,直接修改全局变量 allDronesDb

返回值：
nil
*/
func checkAllDronesDbAndWrite2Excel() {
	// 步骤1. 准备文件( 不存在即创建)
	file, err := createOrOpenExcelFile(allDronesDbPath)
	errorPanic(err)

	// 创建一个哈希表用于存储数组A中每个元素的出现次数,初始化
	pathNumMap := make(map[string]int)
	for _, path := range allDronesDb.SigFolderPath {
		pathNumMap[path] = 0 // 初始化 = 0
	}

	for index, path := range allDronesDb.SigFolderPath {
		// 步骤2: 判断 信号文件夹路径是否存在，更新 allDronesDb 对象
		if checkPathExist(path) {
			allDronesDb.SigFolderPathExist = append(allDronesDb.SigFolderPathExist, true)
		} else {
			allDronesDb.SigFolderPathExist = append(allDronesDb.SigFolderPathExist, true)
		}

		// 步骤3: 把 信号文件夹路径是否存在，写入excel
		// 写入内容
		row := allDronesDb.SigFolderPathExist[index]
		file.SetCellValue("机型库", "K"+strconv.Itoa(index+2), row) // 从第2行开始写入。 &row 会写进入 内存地址，,不对
		errorPanic(err)

		// 步骤4: 判断 信号路径重复项数量，更新 allDronesDb 对象
		// 如果path 在map的key里
		if checkStringInMap(pathNumMap, path) {
			pathNumMap[path]++
		}
		allDronesDb.SigFolderPathRepeatNum = append(allDronesDb.SigFolderPathRepeatNum, pathNumMap[path])

		// 步骤5: 把 信号路径重复项数量, 写入excel
		file.SetCellValue("机型库", "N"+strconv.Itoa(index+2), allDronesDb.SigFolderPathRepeatNum[index]) // 从第2行开始写入。 &row 会写进入 内存地址，,不对
		errorPanic(err)
	}

	// 步骤4. 保存生效
	err = file.SaveAs(allDronesDbPath)
	errorPanic(err)

}
