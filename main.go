/**
功能：程序入口
思路：
1. 准备
2. 回放信号
3. 检测
4. 生成报告

1. 准备阶段细分
- 配置日志等级
- 读取配置
- 设置变量（全局变量+局部变量）
- 生成预发送信号列表文件
*/

package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

// 初始化

// 定义一个空接口类型 Any
type Any interface{}

// 飞机结构体
type Drone struct {
	Name     string `json:"name"`
	FreqList int    `json:"freqList"`
	Id       string `json:"id"`
}

// 全局变量
var (
	err          error
	startTime    time.Time // 程序开始时间
	startTimeStr string    // 程序开始时间str

	// 配置相关
	devIp              string //  设备ip
	sigDir             string // 信号包文件夹路径
	sigPkgSendInterval int    // 发送间隔时间:毫秒/MB,按信号包大小
	cdFolderInterval   int    // 换文件夹等待时间:秒
	queryDroneInterval int    // 查询无人机间隔时间:秒

	// 文件相关
	preSendHistoryFilePath      string         // 预发送记录文件 路径
	preSendHistoryFileSheetName string         // 预发送记录文件 sheetName = "待发送列表"
	queryHistroyFilePath        string         // 查询记录文件 路径
	reportFilePath              string         // 分析报告文件 路径
	preSendHistoryFile          *excelize.File // 预发送记录文件
	queryHistroyFile            *excelize.File // 查询记录文件
	reportFile                  *excelize.File // 分析报告文件

	// 发送相关，没研究
	any          Any
	sendIsStart  = make(chan Any, 1)
	sendIsEnd    = make(chan Any, 1) // 发送是否结束
	userEndSend  = make(chan Any, 1)
	userEndQuery = make(chan Any, 1)
)

// 初始化，默认调用
func init() {
	// 配置日志等级
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Debug("------------ ready 阶段 start")

	// 读取配置
	logrus.Debug("读取配置")
	readConfig("config", "ini", ".")

	// 设置变量（全局变量+局部变量）
	logrus.Debug("设置变量（全局变量+局部变量）")
	setVar()

	// 生成预发送信号列表文件
	logrus.Debug("生成预发送信号列表文件")
	createPreSendHistoryFile(preSendHistoryFilePath)
	logrus.Debug("------------ ready 阶段 end")
}

func todoList() {
	logrus.Info("----------- 待办事项 start")
	logrus.Info("----------- 发送时, 最后一条数据是 [换文件夹], 浪费时间")
	logrus.Info("----------- SendRequestByGraphql 方法config.Token有问题")
	logrus.Info("----------- 别人写的代码，为什么不用token? ")
	logrus.Info("----------- 待发送列表，查询的第一个飞机 id没查出来 ")
	logrus.Info("----------- var droneObj Drone 为什么要写在for里 ")
	logrus.Info("----------- 后面把待发送列表，写成json格式的，好处理 ")
	logrus.Info("----------- 查询的id，和id.txt的可能不一样，不能通过id绝对判断 ")
	logrus.Info("----------- currentSigFolderDirList 假设，后面改 ")
	logrus.Info("----------- 写入报告里的 currentSigFolderDir 是临时写的一个变量。后续需要在查询列表加一列，信号报路径，读这个内容")

	logrus.Info("----------- 需要提供断点续查功能，中途掉线")
	logrus.Info("----------- 等待15秒发送，好像没生效")
	logrus.Info("----------- 除了要写入excel,还要写入t'x'ttxt")
	logrus.Info("----------- 配置文件，可以配置bug等级，不区分大小写")
	logrus.Info("----------- 打包程序")
	logrus.Info("----------- 打包程序，并生成命令，feed、ready, report 查看源代码，查看源代码打包代码6")
	logrus.Info("----------- 信号包xinhao里就一个可用的信号时，就会报错？？？什么原因？怎么解决？")
	logrus.Info("----------- 如何让程序写入excel,打开excel表不影响程序？txt可以吗？")
	logrus.Info("----------- 异常原因没写，repoort里的，没有做判断？")
	logrus.Info("----------- 没有id.txt的时候，report肯定时错误的，这种情况如何判断？？写一个逻辑")
	logrus.Info("----------- 待办事项 end")
}

// 程序入口
func main() {
	todoList() // 待办事项，后面删
	logrus.Debug("------------feed 阶段 start")

	// 步骤3：发送信号      - 原来的 feed 环节
	go sendTask()
	queryTask()

	logrus.Debug("------------feed 阶段 end")

	logrus.Debug("--------------- report 环境 start ---------------")
	// 步骤4：判断设备检测的是否对   - 原来的 report 环节
	// 比较
	// 生成报告
	createReport()
	logrus.Debug("--------------- report 环境 end ---------------")

}
