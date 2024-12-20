/*
*
功能：工具类 - 不知道写哪些内容
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

/*
功能：创建或者打开excel文件
参数：
1. filePath 文件路径

返回值：
1. excelize 对象

思路：
1. 检测文件是否存在，不存在就创建文件
2. 文件存在，就打开
*/
func createOrOpenExcelFile(filePath string) (*excelize.File, error) {
	_, err = os.Stat(filePath)
	// 1. 检测文件是否存在，不存在就创建文件
	if os.IsNotExist(err) {
		logrus.Debugf("文件%v 不存在,创建新文件", filePath)
		return excelize.NewFile(), nil
	} else if err != nil {
		logrus.Errorf("获取文件%v状态时出错", err)
		return nil, err
	}
	// 2. 文件存在，就打开
	logrus.Debug("文件存在,打开文件, path=", filePath)
	return excelize.OpenFile(filePath)
}

/*
功能：创建或者打开txt文件
参数：
1. filePath 文件路径

返回值：
1. excelize 对象

思路：
1. 检测文件是否存在，不存在就创建文件
2. 文件存在，就打开
*/
func createOrOpenTxtFile(filePath string) (*os.File, error) {
	// 1. 检测文件是否存在，不存在就创建文件  2. 文件存在，就打开
	// 打开文件或创建新文件
	// file, err := os.Create(filePath) // 好像只能创建 - 也能用
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666) // 另一种写法,这种更推荐
	if err != nil {
		logrus.Error("文件创建或打开失败,err= ", err)
		return nil, err
	}
	// defer file.Close() // 注释掉。不然报错：msg="写入txt文件失败, err=write 待发送列表-20241203-163152.txt: file already closed"

	return file, err
}

/*
功能：time 转成 string类型, 为了文件命名 20240109-103022 (年月日-时分秒)
参数：
1. time (time.Time) 时间类型

返回值：
1. timeStr (string)
2. error (本来打算返回error,但不知道怎么操作)
*/
func time2stringforFilename(time time.Time) string {
	timeStr := fmt.Sprintf("%d%02d%02d-%02d%02d%02d", time.Year(), time.Month(),
		time.Day(), time.Hour(), time.Minute(), time.Second())
	return timeStr
}

/*
功能: string 转成 time 类型
参数：
1. timeStr (string) eg. 20241209-094427

返回值：
1. time (time.Time) 时间类型
2. error (本来打算返回error,但不知道怎么操作)
*/
func string2time(timeStr string) time.Time {
	t, err := time.Parse("20060102-150405", timeStr)
	if err != nil {
		logrus.Error("转换str时间类型, 报错, err= ", err)
	}
	return t
}

/*
功能：判断文件 后缀是否 以指定后缀结尾。如判断文件 结尾，是否包含 “机型.txt”
参数：
1. path 文件路径
2. specifySuffixStr 指定结尾内容

返回值：
1. bool
*/
func fileExistSuffixStr(path string, specifySuffixStr string) bool {
	return strings.HasSuffix(path, specifySuffixStr)
}

/*
功能：判断文件 后缀是否 包含指定 扩展名。如扩展名，是否以指定 “.bvsp”
参数：
1. path 文件路径, 建议绝对路径
2. specifyExt 指定扩展名

返回值：
1. bool
*/
func fileExistExt(path string, specifyExt string) bool {
	fileExt := filepath.Ext(path)
	return fileExt == specifyExt
}

/*
功能：判断目录下是否包含指定文件。如sig 目录下是否包含 “机型.txt”
参数：
1. dirPath 目录路径
2. specifyFile 指定文件名. 如 “机型.txt”

返回值：
1. bool
*/
func fileExist(dirPath string, specifyFile string) bool {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		logrus.Error("func=fileExist(), 目录不存在, Error reading directory: ", err)
		return false
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == specifyFile {
			return true
		}
	}
	return false
}

/*
功能：判断目录下是否还存在目录
参数：
1. dirPath 目录路径, 建议绝对路径

返回值：
1. bool
*/
func dirExistDir(dirPath string) bool {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		logrus.Error("func=dirExistDir(), 目录不存在, Error reading directory: ", err)
		return false
	}

	for _, file := range files {
		if file.IsDir() {
			return true
		}
	}
	return false
}

/*
功能：判断目录下是否是最终目录
参数：
1. dirPath 目录路径, 建议绝对路径

返回值：
1. bool
*/
func dirIsEndDir(dirPath string) bool {
	// 不启用机型库 配置 写法
	if !dronesDbEnable {
		files, err := os.ReadDir(dirPath)
		if err != nil {
			logrus.Error("func=dirIsEndDir(), 目录不存在, Error reading directory: ", err)
			return false
		}

		for _, file := range files {
			if file.IsDir() {
				return false
			}
		}
		return true
	}

	// 启用机型库 配置 写法
	if dronesDbEnable {
		files, err := os.ReadDir(dirPath)
		if err != nil {
			logrus.Error("func=dirIsEndDir(), 目录不存在, Error reading directory: ", err)
			return false
		}

		for _, file := range files {
			fileInfo, err := os.Lstat(filepath.Join(dirPath, file.Name()))
			errorPanic(err)
			if fileInfo.IsDir() || (fileInfo.Mode()&os.ModeSymlink != 0) { // 是文件夹 或者链接形式文件夹
				return false
			}
		}
		return true
	}
	return true
}

/*
功能：判断路径是否存在
参数：
1. dirPath 目录路径, 建议绝对路径

返回值：
1. bool
*/
func checkPathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	fmt.Println("发生错误:", err)
	return false
}

/*
功能：获取excel已有数据行数
参数：
1. path 文件路径, 建议绝对路径
2. sheetName 表名

返回值：
1. int
*/
func getExcelRowsCount(path string, sheetName string) (int, error) {
	file, err := createOrOpenExcelFile(path)
	logrus.Debugf("func=getExcelRowsCount(), path===== %v, sheetName=%v", path, sheetName)
	if err != nil {
		logrus.Error("func=getExcelRowsCount(), 文件不存在, Error reading directory= ", err)
		return 0, err
	}

	// 获取工作表的最大行数
	rows, err := file.Rows(sheetName)
	count := 0
	for rows.Next() {
		count++
	}
	logrus.Debug("func=getExcelRowsCount(), rowCount=====", count)
	return count, err

}

/*
功能：从excel获取内容
参数：
1. path 文件路径, 建议绝对路径
2. sheetName 表名

返回值：
1. dronesDb DronesDB
*/
func getRowsFromExcel(path string, sheetName string) DroneDB {
	// returnRows = make([]string, 0) // 重置信号列表为空

	file, err := createOrOpenExcelFile(path)
	logrus.Infof("func=getRowsFromExcel(), path= %v, sheetName=%v", path, sheetName)
	errorPanic(err)

	// 获取工作表所有列
	rows, err := file.Rows(sheetName)
	errorPanic(err)
	logrus.Info("rows= ", rows)

	index := 2
	for rows.Next() {
		/*
			Id                     []string `json:"id"`                     // ID
			Manufacture            []string `json:"manufacture"`            // 厂商
			Brand                  []string `json:"brand"`                  // 品牌
			Model                  []string `json:"model"`                  // 型号
			Protocol               []string `json:"protocol"`               // 协议
			Subtype                []string `json:"subtype"`                // 子类型
			FreqBand               []string `json:"freqBand"`               // 频段
			Freq                   []string `json:"freq"`                   // 频率
			SigFolderName          []string `json:"sigFolderName"`          // 信号文件夹名称(品牌-型号-频段-详细频率)
			SigFolderPath          []string `json:"sigFolderPath"`          // 信号文件夹路径
			SigFolderPathExist     []bool   `json:"sigFolderPathExist"`     // 信号文件夹路径是否存在
			DroneTxt               []string `json:"droneTxt"`               // 机型.txt内容
			DroneIdTxt             []string `json:"droneIdTxt"`             // id.txt内容
			SigFolderPathRepeatNum []int    `json:"sigFolderPathRepeatNum"` // 信号文件夹路径重复数量
			SeaFilePath            []string `json:"seaFilePath"`            // seafile链接
		*/
		// logrus.Info("getRowsFromExcel, 从excel读取内容, index= ", index)
		id, err := file.GetCellValue(sheetName, "A"+strconv.Itoa(index)) // ID
		errorPanic(err)
		// if id != "" { // 不判断发送最后一条空数据时，报错。这个好像没用，如果运行正常，就删了
		dronesDb.Id = append(dronesDb.Id, id)
		// }
		manufacture, err := file.GetCellValue(sheetName, "B"+strconv.Itoa(index)) // 厂商
		errorPanic(err)
		if manufacture == "" { // 如果不判断发送最后一条空数据时，会报错。因为有 rows.Next()。会获取到最后一条数据，下一行的空数据
			return dronesDb
		}
		dronesDb.Manufacture = append(dronesDb.Manufacture, manufacture)

		brand, err := file.GetCellValue(sheetName, "C"+strconv.Itoa(index)) // 品牌
		errorPanic(err)
		dronesDb.Brand = append(dronesDb.Brand, brand)

		model, err := file.GetCellValue(sheetName, "D"+strconv.Itoa(index)) // 型号
		errorPanic(err)
		dronesDb.Model = append(dronesDb.Model, model)

		protocol, err := file.GetCellValue(sheetName, "E"+strconv.Itoa(index)) // 协议
		errorPanic(err)
		dronesDb.Protocol = append(dronesDb.Protocol, protocol)

		subtype, err := file.GetCellValue(sheetName, "F"+strconv.Itoa(index)) // 子类型
		errorPanic(err)
		dronesDb.Subtype = append(dronesDb.Subtype, subtype)

		freqBand, err := file.GetCellValue(sheetName, "G"+strconv.Itoa(index)) // 频段
		errorPanic(err)
		dronesDb.FreqBand = append(dronesDb.FreqBand, freqBand)

		freq, err := file.GetCellValue(sheetName, "H"+strconv.Itoa(index)) // 频率
		errorPanic(err)
		dronesDb.Freq = append(dronesDb.Freq, freq)

		sigFolderName, err := file.GetCellValue(sheetName, "I"+strconv.Itoa(index)) // 信号文件夹名称
		errorPanic(err)
		dronesDb.SigFolderName = append(dronesDb.SigFolderName, sigFolderName)

		sigFolderPath, err := file.GetCellValue(sheetName, "J"+strconv.Itoa(index)) // 信号文件夹路径
		errorPanic(err)
		dronesDb.SigFolderPath = append(dronesDb.SigFolderPath, sigFolderPath)

		// sigFolderPathExistStr, err := file.GetCellValue(sheetName, "K"+strconv.Itoa(index)) // 信号文件夹路径是否存在。这个需要后期, 程序判断后赋值的。而不是取值
		// logrus.Info("--------------- sigFolderPathExistStr = ", sigFolderPathExistStr)
		// errorPanic(err)
		// sigFolderPathExist, err := strconv.ParseBool(sigFolderPathExistStr)
		// errorPanic(err)
		// dronesDb.SigFolderPathExist = append(dronesDb.SigFolderPathExist, sigFolderPathExist)

		droneTxt, err := file.GetCellValue(sheetName, "L"+strconv.Itoa(index)) // 机型.txt内容
		errorPanic(err)
		dronesDb.DroneTxt = append(dronesDb.DroneTxt, droneTxt)

		droneIdTxt, err := file.GetCellValue(sheetName, "M"+strconv.Itoa(index)) // id.txt内容
		errorPanic(err)
		dronesDb.DroneIdTxt = append(dronesDb.DroneIdTxt, droneIdTxt)

		// sigFolderPathRepeatNumStr, err := file.GetCellValue(sheetName, "N"+strconv.Itoa(index)) // 信号文件夹路径重复数量。 这个需要后期, 程序判断后赋值的。而不是取值
		// errorPanic(err)
		// sigFolderPathRepeatNum, err := strconv.Atoi(sigFolderPathRepeatNumStr)
		// errorEcho(err)
		// dronesDb.SigFolderPathRepeatNum = append(dronesDb.SigFolderPathRepeatNum, sigFolderPathRepeatNum)

		seaFilePath, err := file.GetCellValue(sheetName, "O"+strconv.Itoa(index)) // seafile链接
		errorPanic(err)
		dronesDb.SeaFilePath = append(dronesDb.SeaFilePath, seaFilePath)

		index++
	}
	return dronesDb
}

/*
功能：从all机型库 excel 获取内容
参数：
1. path 文件路径, 建议绝对路径
2. sheetName 表名

返回值：
1. dronesDb DronesDB
*/
func getAllDronesDbFromExcel(path string, sheetName string) DroneDB {

	file, err := createOrOpenExcelFile(path)
	logrus.Infof("func=getAllDronesDbFromExcel(), path= %v, sheetName=%v", path, sheetName)
	errorPanic(err)

	// 获取工作表所有列
	rows, err := file.Rows(sheetName)
	errorPanic(err)
	logrus.Info("rows= ", rows)

	index := 2
	for rows.Next() {
		/*
			Id                     []string `json:"id"`                     // ID
			Manufacture            []string `json:"manufacture"`            // 厂商
			Brand                  []string `json:"brand"`                  // 品牌
			Model                  []string `json:"model"`                  // 型号
			Protocol               []string `json:"protocol"`               // 协议
			Subtype                []string `json:"subtype"`                // 子类型
			FreqBand               []string `json:"freqBand"`               // 频段
			Freq                   []string `json:"freq"`                   // 频率
			SigFolderName          []string `json:"sigFolderName"`          // 信号文件夹名称(品牌-型号-频段-详细频率)
			SigFolderPath          []string `json:"sigFolderPath"`          // 信号文件夹路径
			SigFolderPathExist     []bool   `json:"sigFolderPathExist"`     // 信号文件夹路径是否存在
			DroneTxt               []string `json:"droneTxt"`               // 机型.txt内容
			DroneIdTxt             []string `json:"droneIdTxt"`             // id.txt内容
			SigFolderPathRepeatNum []int    `json:"sigFolderPathRepeatNum"` // 信号文件夹路径重复数量
			SeaFilePath            []string `json:"seaFilePath"`       	    // seafile链接
		*/
		// logrus.Info("getRowsFromExcel, 从excel读取内容, index= ", index)
		id, err := file.GetCellValue(sheetName, "A"+strconv.Itoa(index)) // ID
		errorPanic(err)
		// if id != "" { // 不判断发送最后一条空数据时，报错。这个好像没用，如果运行正常，就删了
		allDronesDb.Id = append(allDronesDb.Id, id)
		// }
		manufacture, err := file.GetCellValue(sheetName, "B"+strconv.Itoa(index)) // 厂商
		errorPanic(err)
		if manufacture == "" { // 如果不判断发送最后一条空数据时，会报错。因为有 rows.Next()。会获取到最后一条数据，下一行的空数据
			return allDronesDb
		}
		allDronesDb.Manufacture = append(allDronesDb.Manufacture, manufacture)

		brand, err := file.GetCellValue(sheetName, "C"+strconv.Itoa(index)) // 品牌
		errorPanic(err)
		allDronesDb.Brand = append(allDronesDb.Brand, brand)

		model, err := file.GetCellValue(sheetName, "D"+strconv.Itoa(index)) // 型号
		errorPanic(err)
		allDronesDb.Model = append(allDronesDb.Model, model)

		protocol, err := file.GetCellValue(sheetName, "E"+strconv.Itoa(index)) // 协议
		errorPanic(err)
		allDronesDb.Protocol = append(allDronesDb.Protocol, protocol)

		subtype, err := file.GetCellValue(sheetName, "F"+strconv.Itoa(index)) // 子类型
		errorPanic(err)
		allDronesDb.Subtype = append(allDronesDb.Subtype, subtype)

		freqBand, err := file.GetCellValue(sheetName, "G"+strconv.Itoa(index)) // 频段
		errorPanic(err)
		allDronesDb.FreqBand = append(allDronesDb.FreqBand, freqBand)

		freq, err := file.GetCellValue(sheetName, "H"+strconv.Itoa(index)) // 频率
		errorPanic(err)
		allDronesDb.Freq = append(allDronesDb.Freq, freq)

		sigFolderName, err := file.GetCellValue(sheetName, "I"+strconv.Itoa(index)) // 信号文件夹名称
		errorPanic(err)
		allDronesDb.SigFolderName = append(allDronesDb.SigFolderName, sigFolderName)

		sigFolderPath, err := file.GetCellValue(sheetName, "J"+strconv.Itoa(index)) // 信号文件夹路径
		errorPanic(err)
		allDronesDb.SigFolderPath = append(allDronesDb.SigFolderPath, sigFolderPath)

		// sigFolderPathExistStr, err := file.GetCellValue(sheetName, "K"+strconv.Itoa(index)) // 信号文件夹路径是否存在。这个需要后期, 程序判断后赋值的。而不是取值
		// logrus.Info("--------------- sigFolderPathExistStr = ", sigFolderPathExistStr)
		// errorPanic(err)
		// sigFolderPathExist, err := strconv.ParseBool(sigFolderPathExistStr)
		// errorPanic(err)
		// allDronesDb.SigFolderPathExist = append(allDronesDb.SigFolderPathExist, sigFolderPathExist)

		droneTxt, err := file.GetCellValue(sheetName, "L"+strconv.Itoa(index)) // 机型.txt内容
		errorPanic(err)
		allDronesDb.DroneTxt = append(allDronesDb.DroneTxt, droneTxt)

		droneIdTxt, err := file.GetCellValue(sheetName, "M"+strconv.Itoa(index)) // id.txt内容
		errorPanic(err)
		allDronesDb.DroneIdTxt = append(allDronesDb.DroneIdTxt, droneIdTxt)

		// sigFolderPathRepeatNumStr, err := file.GetCellValue(sheetName, "N"+strconv.Itoa(index)) // 信号文件夹路径重复数量。 这个需要后期, 程序判断后赋值的。而不是取值
		// errorPanic(err)
		// sigFolderPathRepeatNum, err := strconv.Atoi(sigFolderPathRepeatNumStr)
		// errorEcho(err)
		// allDronesDb.SigFolderPathRepeatNum = append(allDronesDb.SigFolderPathRepeatNum, sigFolderPathRepeatNum)

		seaFilePath, err := file.GetCellValue(sheetName, "O"+strconv.Itoa(index)) // seafile链接
		errorPanic(err)
		allDronesDb.SeaFilePath = append(dronesDb.SeaFilePath, seaFilePath)

		index++
	}
	return allDronesDb
}

/*
功能：判断string 是否在 map的key里
参数：
1. path 文件路径, 建议绝对路径
2. str 字符串

返回值：
1. dronesDb DronesDB
*/
func checkStringInMap(m map[string]int, str string) bool {
	_, exists := m[str]
	return exists
}
