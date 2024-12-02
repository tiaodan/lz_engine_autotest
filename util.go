/*
*
功能：工具类 - 不知道写哪些内容
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	file, err := os.Create("output.txt")
	if err != nil {
		logrus.Error("文件创建或打开失败,err= ", err)
		return nil, err
	}
	defer file.Close()

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
	files, err := os.ReadDir(dirPath)
	if err != nil {
		logrus.Error("目录不存在, Error reading directory: ", err)
		return false
	}

	for _, file := range files {
		if file.IsDir() {
			return false
		}
	}
	return true
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
