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
	"os"
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

	preSendHistoryFileTxtPath string   // 预发送记录文件txt 路径
	queryHistroyFileTxtPath   string   // 查询记录文件txt 路径
	reportFileTxtPath         string   // 分析报告文件txt 路径
	preSendHistoryTxtFile     *os.File // 预发送记录文件
	queryHistroyTxtFile       *os.File // 查询记录文件
	reportTxtFile             *os.File // 分析报告文件

	// 发送相关，没研究
	any                      Any
	sendIsStart              = make(chan Any, 1)
	sendIsEnd                = make(chan Any, 1) // 发送是否结束
	userEndSend              = make(chan Any, 1)
	userEndQuery             = make(chan Any, 1)
	userChangeQuerySigFolder = make(chan Any, 1) // 切换查询信号文件夹。 v0.0.0.1 为优化查询效率新增

	logLevel string // 日志级别
)

// 初始化，默认调用
func init() {
	// 读取配置
	logrus.Debug("读取配置")
	readConfig("config", "ini", ".")

	// 配置日志等级
	// logrus.SetLevel(logrus.DebugLevel)
	// logrus.SetLevel(logrus.InfoLevel)
	logrus.Debug("------------ ready 阶段 start")

	// 设置变量（全局变量+局部变量）
	logrus.Debug("设置变量（全局变量+局部变量）")
	setVar()

	// 生成预发送信号列表文件
	logrus.Debug("生成预发送信号列表文件")
	createPreSendHistoryFile(preSendHistoryFilePath)
	createPreSendHistoryFileTxt(preSendHistoryFileTxtPath) // txt文件
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
	logrus.Info("----------- 把信号回放，弄成并发形式的，提高效率，考虑会有重复信号，是否会有重复id的问题？")
	logrus.Info("----------- 把信号回放，提高效率，如果检测到了，就停止发送？整体时间：从1天-提高到3小时?并考虑引擎能不能承受住")
	logrus.Info("----------- 信号列表xlsx，没排序")
	logrus.Info("----------- 换文件夹，好像没有等待-- 确实等待了, 知识等待期间一直在查")
	logrus.Info("----------- 2个文件夹, {3DR Solo 2462000 8adc9635291e} + {AEE Sparrow2 5745000 3801462ef111} , 生成报告判断不对")
	logrus.Info("----------- 分析报告，最后列，加上 信号文件夹路径")
	logrus.Info("----------- 就一个文件夹的时候，会一直阻塞。不打印：停止查询, sendIsEnd。无法进入生成报告步骤. 原因：是因为只有一个信号文件夹时， 代码没收到终止信号, 一直阻塞。但是按理说queryTask()应该一直查呀。是因为收到userEndSend信号, 才发送userEndQuery信号, 查询任务没收到endQyery信号，所以一直阻塞")
	logrus.Info("----------- 待发送信号列表, 排序-----")
	logrus.Info("----------- 查询结束后的，空挡时间，查询列表excel还会写出2-4条数据，这时应该有个暂停查询的信号，过段时间再恢复 ")
	logrus.Info("----------- 什么日志，能打印出类似controller的 system.go  、 xx.go 这样的日志信息？ ")
	logrus.Info("----------- 删除多余日志 ")
	logrus.Info("----------- 研究如何让2个select不受影响,for循环里, 不会一个select阻塞, 另一个就不能接受信号了 ")
	logrus.Info("----------- 测试修改的bug, 测试是否没问题？才上传github")
	logrus.Info("----------- 生成feed命令后，加入windows 环境变量中")
	logrus.Info("----------- feed.py 读取不了这种路径feed.py -i 192.168.85.95 -d E:斜杠xinhao斜杠EVO II反斜杠EVO_2.4")
	logrus.Info("----------- 如果机型.txt 里2行, 会报错。要处理这种")
	logrus.Info("----------- 如果没有机型.txt , 没有id.txt，也要有一种方式，检测信号")
	logrus.Info("----------- 第一次发送失败的，过完一遍后，再发一遍失败的信号")
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

	logrus.Info("--------------- report 环境 start ---------------")
	// 步骤4：判断设备检测的是否对   - 原来的 report 环节
	// 比较
	// 生成报告
	createReport()
	logrus.Info("--------------- report 环境 end ---------------")

}
