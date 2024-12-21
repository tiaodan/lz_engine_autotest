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
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
)

// 初始化

// 定义一个空接口类型 Any
type Any interface{}

// 飞机结构体
type Drone struct {
	Name     string   `json:"name"`
	FreqList int      `json:"freqList"`
	Id       []string `json:"id"` // 改成[], 为了能识别id.txt 中多个id
}

// 机型库 结构体
type DroneDB struct {
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
	SeaFilePath            []string `json:"seaFilePath"`            // 信号seafile链接
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
	mistakeFreqConfig  int    // 查询无人机频率 最大误差值 单位：Mhz 配置文件中写的

	// 配置-机型库相关
	dronesDb        DroneDB // 机型库结构体-回放信号用
	allDronesDb     DroneDB // 机型库结构体-all
	dronesDbEnable  bool    // 是否根据机型库，进行自动化引擎测试
	dronesDbPath    string  // 机型库路径 - 回放信号用
	allDronesDbPath string  // all机型库路径

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

	logLevel            string // 日志级别
	changeFolderFlag    bool   // 换文件夹标志
	changeFolderFlagNum int    // 换文件夹标志 num

)

// 初始化，默认调用
func init() {
	// 默认调用
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
	logrus.Info("----------- 还是有时查询无法进去report阶段?好像有误差结果=TRUE时, 不会中断。需要再试试。是这样的")
	logrus.Info("----------- 写入txt文件, 写的不全，只写入了最后一个？确实写了，好像旧的把新的覆盖了，没有在原来的基础上写入？")
	logrus.Info("----------- 分阶段操作, ready send report阶段,有命令")
	logrus.Info("----------- 如何在windows 界面编译成linux执行程序")
	logrus.Info("----------- ready feed report 三个命令整合到一个控制台上。1-ready 2-feed 3 report 4一键执行上述内容")
	logrus.Info("----------- vipper 写入配置文件，写的都是小写，跟原始文件不同")
	logrus.Info("----------- 换文件时候，查询的记录算在下一个信号了，如果查询时间是4秒，很容易当前信号发完了，下个信号时候才检测到。如果优化，换文件夹时，按当前文件信号来记录。先2秒用着，有问题了再说")
	logrus.Info("----------- 统计里加一个总用时")
	logrus.Info("----------- 找一个发送信号的最优效率解")
	logrus.Info("----------- 要判断一个信号，有没有误报、多报")
	logrus.Info("----------- MMC 1550 会报很多信号")
	logrus.Info("----------- KEWEITAI_x6lm?2420 这个信号目前有2个信号, 海外=1434927124, 标准版=753341352, 如何处理这种情况？？问无线同事，有的机型， 比如遥控器，可能随机分配id。有的无人机解不出id，给了一个固定id")
	logrus.Info("----------- 是否要考虑频率检测的误差调大些，比如20M?RC-MICROZONE_MC6C_2450 这个显示2477M 差快30? 异常原因要些 误差大")
	logrus.Info("----------- 配置项, 加一个频率误差范围，允许的范围 int型,表示多少M 比如10  20")
	logrus.Info("----------- 查询列表，切换文件夹的时间间隔，查询的当前飞机应该 = 上一次飞机")
	logrus.Info("----------- 待写一个使用说明-给同事看（同事看完不用问，直接就能用，说明写好了）")
	logrus.Info("----------- 把id 和机型.txt 信息同步到 云excel, 方便同事，自己创建id.txt文件")
	logrus.Info("----------- 查询列表如果只有1条信号, 报告不会出现该条. 信号=RTK-RTK_5767_3")
	logrus.Info("----------- 如果信号放一轮，没有的话，是否需要循环放2遍？")
	logrus.Info("----------- 测试稳定了，看分析报告的文件少？咋回事？")
	logrus.Info("----------- 信号个数少于10,需要循环2遍/3遍")
	logrus.Info("----------- 三个信号包，要查询的id+机型一致, 会认为是1个, 有问题")
	logrus.Info("----------- 日志老是显示: tcp发送失败，等待3秒后重试")
	logrus.Info("----------- 能检测到的, 要稳定百分百检测到")
	logrus.Info("----------- 换文件夹后，没发送新的信号")
	logrus.Info("----------- 测试效率最高, 超时时间ttl=8秒, 切换文件夹时间=6秒, 查询间隔2秒")
	logrus.Info("----------- 统计报告report excel 加上总用时")
	logrus.Info("----------- 生成report 某一个sheet, 写上测试的版本信息，如下：")
	/*
		软件版本-名称	2024.12.09
		软件版本-编译时间(引擎)	2024-12-09T00:52:37
		软件版本-githash(引擎)	cff9ee4266e979481a4e744cfe6e579c1e652ae5
		软件版本-md5
	*/
	logrus.Info("----------- 处理警告")
	logrus.Info("----------- 考虑多报、误报怎么判断")
	logrus.Info("----------- report 判断 机型名称for 循环里，判断误差。因为要判断该行的数，而不是匹配到就行")
	logrus.Info("----------- 要求: 上一个信号, 不能影响下一个信号,一般切文件夹时间 - 超时时间 =4秒")
	logrus.Info("----------- 优化: 生成查询列表, 判断跳过部分信号条件: 机型== && 有id ,频率在误差范围内")
	logrus.Info("----------- 误报、多报如何判定？")
	logrus.Info("----------- 优化, 查询excel, 机型名称匹配+id匹配+误差范围内, 才发切换文件夹信号")
	logrus.Info("----------- 切换文件夹期间，查到数据，会影响文件夹切换逻辑, 如何修改？")
	logrus.Info("----------- 最后一个切换文件夹, 会消耗10秒时间，如何判断最后一个[换文件夹], 然后就不time.sleep了")
	logrus.Info("----------- 对比机型名称, 要不区分名字大小写")
	logrus.Info("----------- HUBSAN_ZINOpro_5825 比如这个机型, 有时候检测不到5825批量, 是别的频段,误差>10M, 思考检测不到, 循环发送2次!")
	logrus.Info("----------- 每个版本,机型名称不同。比如老版本{DJI Phantom4/4Pro 2406500 60601f09c0b0} 最新：{DJI Phantom4/4Pro/Inspire2 2406500 60601f09c0b0}, 这种情况如何处理？是否弄个文件夹, 专门放配置文件, 机型txt文件?每个版本,建一个机型库文件，或者一个表格也可以。像一种关联方法")
	logrus.Info("----------- 查询列表, 判断有误差、无误差true 好像有问题，是否加一列id是否相等? ")
	logrus.Info("----------- 回放信号时，有实体飞机,比如mavic 2, 回放mavic2 会识别不到 O2 机型名称的信号")
	logrus.Info("----------- 多报, 可以回放dji o2信号")
	logrus.Info("----------- DJI最后再测 , 可能会影响其它信号")
	logrus.Info("----------- id.txt 多个id, 查询时也能识别。id.txt中用/区分. 比如 753341352/abdce123")
	logrus.Info("----------- 机型.txt 多个机型，查询时也能识别")
	logrus.Info("----------- DJI重命名文件夹, 让该文件夹最后回放。真实情况, 从公司角度触发, 是否DJI的飞机要首先回放?")

	logrus.Info("----------- 用编译好的版本，测试 C:/A_software/xinhao/xinhao-ok/jinjuli-测过-完全通过 所有信号, 完成时, 会报错。如下: 。并且会少待发送列表excel文件.文件夹目录: 单个信号, 多个信号,DJI 。原因是: 查询列表excel 空了")
	/*
		time="2024-12-11T17:19:37+08:00" level=info msg="--------------- report 环境 start ---------------"
		time="2024-12-11T17:19:37+08:00" level=info msg="配置 devIp (设备ip)= 192.168.84.179"
		time="2024-12-11T17:19:37+08:00" level=info msg="配置 sigDir (信号包根目录)= C:\\A_software\\xinhao\\xinhao-ok\\jinjuli-测过-完全通过"
		time="2024-12-11T17:19:37+08:00" level=info msg="配置 sigPkgSendInterval (发送间隔时间:毫秒/MB,按信号包大小)= 100"
		time="2024-12-11T17:19:37+08:00" level=info msg="配置 cdFolderInterval (换文件夹等待时间:秒)= 9"
		time="2024-12-11T17:19:37+08:00" level=info msg="配置 queryDroneInterval (查询无人机间隔时间:秒)= 2"
		time="2024-12-11T17:19:37+08:00" level=info msg="配置 logLevel (日志级别)= info"
		time="2024-12-11T17:19:37+08:00" level=info msg="配置 开始时间str = 20241211-154922"
		panic: runtime error: index out of range [5] with length 0

		goroutine 1 [running]:
		main.createReport()
		        C:/Users/root1/Desktop/lz_engine_autotest/report.go:60 +0x1872
		main.report()
		        C:/Users/root1/Desktop/lz_engine_autotest/main.go:306 +0x231
		main.main()
		        C:/Users/root1/Desktop/lz_engine_autotest/main.go:228 +0x5b9
	*/
	logrus.Info("----------- 所有的打印生成日志, 生成log, 用于后续定位问题")
	logrus.Info("----------- 开2个回放cmd,遇到同一个信号，可能报错")
	logrus.Info("----------- 只有1个信号, feed程序发不出信号。不知道是不是没有id原因? 因为待发送列表没东西")
	logrus.Info("----------- rows.Next()。如何避免，获取到最后一条数据，下一行的空数据")
	logrus.Info("----------- 配置里，如果文件夹下没有文件夹，ready阶段生成不了内容")
	logrus.Info("----------- ubuntu 是否需要管理员权限，才能创建软连接，还没测试")
	logrus.Info("----------- id。txt 要支持正则表达式，因为id有可能是随机的")
	logrus.Info("----------- 要写单元测试")
	logrus.Info("----------- 把所有 if dronesDbEnable 改成if eles形式")
	logrus.Info("----------- allDronesDb 存在数组数量不一致情况，如果excel某一行没写，allDronedB 那一行就少")
	logrus.Info("----------- 对于不容易检测到的信号，待发送列表，创建2遍/3遍发送信号")
	logrus.Info("----------- ready 阶段, 删除excel文件，会把子目录excel也删除")
	logrus.Info("----------- 删除目录 xinhao-test 已创建的文件夹不会删除，仍然会保留")
	logrus.Info("----------- 待办事项 end")
}

// 程序入口
func main() {
	// todoList() // 待办事项，后面删

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("请输入要执行的命令：")
	fmt.Println("1 - ready")
	fmt.Println("2 - feed")
	fmt.Println("3 - report,")
	fmt.Println("4 - 一键执行步骤123")
	fmt.Println("5 - delete history file")
	fmt.Println("6 - 一键执行 步骤5、4")
	fmt.Println("0 - 退出")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	programInit() // 程序初始化
	switch input {
	case "1":
		logrus.Info("执行 ready 命令")
		// 执行对应的命令代码
		ready()
	case "2":
		logrus.Info("执行 feed 命令")
		// 执行对应的命令代码
		feed()
	case "3":
		logrus.Info("执行 report 命令")
		// 执行对应的命令代码
		report()
	case "4":
		logrus.Info("一键执行以上所有命令")
		// 执行对应的命令代码
		ready()
		feed()
		report()
	case "5":
		logrus.Info("删除当前目录 xlsx文件, txt文件")
		// 执行对应的命令代码
		deleteHistroyFile()
	case "6":
		logrus.Info("一键执行 步骤5、4")
		// 执行对应的命令代码
		deleteHistroyFile()
		ready()
		feed()
		report()
	case "0":
		logrus.Info("退出程序")
		return
	default:
		logrus.Info("无效输入，请重新输入")
	}

}

// 功能: 程序初始，把ready() feed() report() 初始化操作，都放在这，在main()方法调用
func programInit() {
	// 读取配置
	logrus.Debug("读取配置")
	// readConfig("config", "ini", ".") // 现在配置文件，全是小写了，等修复了此问题再改过来
	readLowerConfig("config", "ini", ".") // 现在配置文件，全是小写了，等修复了此问题再改过来

	// 配置日志等级
	// logrus.SetLevel(logrus.DebugLevel)  // 设置已经写在setVar()方法里
	// logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("------------ ready 阶段 start")

	// 设置变量（全局变量+局部变量）
	logrus.Debug("设置变量（全局变量+局部变量）")
	setVar()

	// 步骤1：读取 机型库文件
	getRowsFromExcel(dronesDbPath, "机型库")
	logrus.Info("机型库= ", dronesDb)

	// 步骤2：判断机型库内容：，并写到机型库文件里
	checkDronesDbAndWrite2Excel()

	// 步骤5：读取 all机型库文件
	getAllDronesDbFromExcel(allDronesDbPath, "机型库")
	logrus.Info("all机型库= len ", len(allDronesDb.Manufacture))

	// 步骤6：判断机型库内容：，并写到all机型库文件里
	checkAllDronesDbAndWrite2Excel()
}

// 功能 ready 流程
func ready() {
	// ready 阶段，写入配置文件 开始时间
	// 设置 程序开始时间变量 初始值
	startTime = time.Now()
	startTimeStr = time2stringforFilename(startTime)
	// 文件相关变量。设置成 文件-时间.xlsx
	preSendHistoryFilePath = "待发送列表-" + startTimeStr + ".xlsx"
	preSendHistoryFileSheetName = "待发送列表"
	preSendHistoryFileTxtPath = "待发送列表-" + startTimeStr + ".txt" // 预发送记录文件txt 路径

	queryHistroyFilePath = "查询列表" + startTimeStr + ".xlsx"   // 查询文件路径
	queryHistroyFileTxtPath = "查询列表" + startTimeStr + ".txt" // 查询记录文件txt 路径
	reportFilePath = "分析报告" + startTimeStr + ".xlsx"         // 查询文件路径

	// 设置配置文件，当前时间
	// viper.SetConfigName("config") // 设置 配置文件名 eg: viper.SetConfigName("config")
	// viper.SetConfigType("ini")    // 设置 配置文件后缀名 eg: viper.SetConfigType("ini")
	// viper.AddConfigPath(".")      // 设置 配置文件路径 eg: viper.AddConfigPath(".")
	viper.Set("time.startTime", startTimeStr)
	viper.Set("file.presendhistoryfilepath", preSendHistoryFilePath)
	viper.Set("file.queryhistroyfilePath", queryHistroyFilePath)
	viper.Set("file.reportfilepath", reportFilePath)

	err = viper.WriteConfig() // 写到配置文件里
	errorPanic(err)
	// 打印变量
	logrus.Debug("全局变量 startTime (程序开始时间)= ", startTime)
	logrus.Debug("全局变量 startTimeStr (查程序开始时间str)= ", startTimeStr)
	logrus.Debug("全局变量 preSendHistoryFilePath (待发送信号记录文件路径)= ", preSendHistoryFilePath)

	// 判断配置文件-配置 dronesDbEnable
	if !dronesDbEnable {
		// 生成预发送信号列表文件
		logrus.Debug("生成预发送信号列表文件")
		createPreSendHistoryFile(preSendHistoryFilePath)
		// createPreSendHistoryFileHeaderTxt(preSendHistoryFileTxtPath) // txt文件表头
		// createPreSendHistoryFileTxt(preSendHistoryFileTxtPath)       // txt文件,这个方法用不到，因为已经在上面写excel方法里createPreSendHistoryFile()，写了txt
	}

	if dronesDbEnable {
		/*
			思路：
			步骤：
			1. 读取机型.xlsx文件，包含N列 （ID）（厂家）（品牌）（型号）（协议）（协议子类型）（频段）(详细频段) （信号文件夹名称）（信号文件夹路径)（信号文件夹是否存在) （机型.txt内容) （id.txt内容) (信号文件夹路径重复序号) 。
			1）判断文件路径是否存在，写到xlsx文件中
			2）判断文件是否打开，如果被别人打开了，提示关闭xlsx文件
			3）写入 信号路径重复的 索引值，重复一个 +1
			3）读取所有信息
			2. 在本地目录，创建信号库的软链接文件 （直接剔除重复项）,信号

			----- 所有逻辑，判断 dronesdbenable，不要影响原先的逻辑
			3. 创建待发送列表
			3. 回放信号
			4. 判断结果：软链接接相同的情况下，如果有1个已经回放完了，其它都设置成相同的结果
			5. 生成报告 （传统方式，生成一个列表。根据机型库.xlsx，会生成另一个列表）

			报告 xlsx 加上索引，做一些dbutil操作。类似于操作db一样
		*/
		// 生成预发送信号列表文件
		logrus.Debug("生成预发送信号列表文件")

		// 步骤3：在本地目录，创建信号库的软链接文件 （直接剔除重复项）,信号。信号路径sigdir 改为 软链接目录
		xinhaoTestPath := createFolderLink()

		// 步骤4：把信号包总路径 改为 软连接的。不影响后续逻辑
		sigDir = xinhaoTestPath

		createPreSendHistoryFile(preSendHistoryFilePath)

		// createPreSendHistoryFileHeaderTxt(preSendHistoryFileTxtPath) // txt文件表头
		// createPreSendHistoryFileTxt(preSendHistoryFileTxtPath)       // txt文件,这个方法用不到，因为已经在上面写excel方法里createPreSendHistoryFile()，写了txt

		logrus.Info("ready end 阶段, dronesDb.SigFolderPath = ", dronesDb.SigFolderPath)
		logrus.Info("ready end 阶段, dronesDb = ", dronesDb)
		logrus.Info("ready end 阶段, allDronesDb.SigFolderPath = ", allDronesDb.SigFolderPath)
		logrus.Info("ready end 阶段, allDronesDb = ", allDronesDb)
	}
	logrus.Info("------------ ready 阶段 end")
}

// 功能 feed 流程
func feed() {
	// 不启用 机型库 写法
	if !dronesDbEnable {
		logrus.Info("------------feed 阶段 start")
		// 1. 创建或者打开文件
		preSendHistoryFile, err = createOrOpenExcelFile(preSendHistoryFilePath)
		errorPanic(err)

		// 步骤3：发送信号      - 原来的 feed 环节
		go sendTask()
		queryTask()

		logrus.Info("------------feed 阶段 end")
	}

	// 启用 机型库 写法
	if dronesDbEnable {
		logrus.Info("------------feed 阶段 start")

		logrus.Info("feed start 阶段, dronesDb.SigFolderPath = ", dronesDb.SigFolderPath)
		logrus.Info("feed start 阶段, dronesDb = ", dronesDb)

		// 1. 创建或者打开文件
		preSendHistoryFile, err = createOrOpenExcelFile(preSendHistoryFilePath)
		errorPanic(err)

		// 步骤3：发送信号      - 原来的 feed 环节
		go sendTask()
		queryTask()

		logrus.Info("------------feed 阶段 end")
		logrus.Info("feed end 阶段, dronesDb.SigFolderPath = ", dronesDb.SigFolderPath)
		logrus.Info("feed end 阶段, dronesDb = ", dronesDb)
	}

}

// 功能 report 流程
func report() {
	// 不启用机型库文件 配置
	if !dronesDbEnable {
		logrus.Info("--------------- report 阶段 进入分支： !dronesDbEnable ---------------")
		logrus.Info("--------------- report 阶段 start ---------------")
		// 1. 创建或者打开文件
		queryHistroyFile, err = createOrOpenExcelFile(queryHistroyFilePath)
		errorPanic(err)
		// 步骤4：判断设备检测的是否对   - 原来的 report 环节
		// 比较
		// 生成报告
		// queryHistroyFilePath = "查询列表20241204-101626.xlsx"                   // 注释：临时测试report模块时用，属于测试代码
		// queryHistroyFile, err = createOrOpenExcelFile(queryHistroyFilePath) // 注释：临时测试report模块时用，属于测试代码
		createReport()
		logrus.Info("--------------- report 阶段 end ---------------")
	}

	// 启用机型库文件 配置
	if dronesDbEnable {
		logrus.Info("--------------- report 阶段 进入分支： dronesDbEnable ---------------")
		logrus.Info("--------------- report 阶段 start ---------------")

		// 1. 创建或者打开文件
		queryHistroyFile, err = createOrOpenExcelFile(queryHistroyFilePath)
		errorPanic(err)

		logrus.Info("report start 阶段, dronesDb.SigFolderPath = ", dronesDb.SigFolderPath)
		logrus.Info("report start 阶段, dronesDb = ", dronesDb)
		logrus.Info("report start 阶段, allDronesDb.SigFolderPath = ", allDronesDb.SigFolderPath)
		logrus.Info("report start 阶段, allDronesDb = ", allDronesDb)
		logrus.Info("--------- report 阶段 dronesDbEnable = ", dronesDbEnable)

		// 步骤5：判断设备检测的是否对   - 原来的 report 环节
		// 比较
		createReport()
		// 步骤6：分析报告-关联 机型库(已经回放信号的)
		createReportRelateSigReplayDronesDb()
		// 步骤7：分析报告-关联 最全机型库
		createReportRelateAllDronesDb()

		logrus.Info("--------------- report 阶段 end ---------------")
		logrus.Info("report end 阶段, dronesDb.SigFolderPath = ", dronesDb.SigFolderPath)
		logrus.Info("report end 阶段, dronesDb = ", dronesDb)
		logrus.Info("report end 阶段, allDronesDb.SigFolderPath = ", allDronesDb.SigFolderPath)
		logrus.Info("report end 阶段, allDronesDb = ", allDronesDb)
	}

}

// 删除当前目录 xlsx文件, txt文件
func deleteHistroyFile() {
	/*
		// 会递归删除当前目录及子目录  xlsx文件, txt文件
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) == ".xlsx" || filepath.Ext(path) == ".txt" {
				fmt.Println("删除文件= ", path)
				err := os.Remove(path)
				if err != nil {
					fmt.Println("删除文件出错:", err)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Println("遍历目录出错:", err)
		}
	*/

	// 只删除当前目录
	// 获取当前目录下的文件列表
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("读取当前目录出错:", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".xlsx" || filepath.Ext(file.Name()) == ".txt" {
			filePath := filepath.Join(".", file.Name())
			fmt.Println("删除文件= ", filePath)
			err := os.Remove(filePath)
			if err != nil {
				fmt.Println("删除文件出错:", err)
			}
		}
	}
}

/*
功能：创建链接。级联创建目录，最后一个子目录，关联上信号文件夹
参数： 这些参数从 dronesDb 对象里拿
1. 品牌 brand string
2. 型号 model string
3. 频段 freqBand string
4. 详细频率 freq string
5. 最后子目录 folderName string 由 拼接而成：品牌-型号-频段-详细频率

返回值：
当前软连接信号总目录 sigDir string
*/
func createFolderLink() string {
	// 获取当前目录的绝对路径
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("无法获取当前目录: %v", err)
	}

	// 先删除当前目录 xinhao-test目录
	xinhaoTestPath := filepath.Join(currentDir, "xinhao-test")
	logrus.Info("xinhaoTestPath 目录= ", xinhaoTestPath)
	os.RemoveAll(xinhaoTestPath)
	if err != nil {
		log.Fatalf("无法删除文件夹 xinhao-test : %v", err)
	}

	// 指定厂家、品牌、频段、频率 - 参数里都有了
	for index, path := range dronesDb.SigFolderPath {
		// 信号路径存在，且重复数=1 才创建链接
		if dronesDb.SigFolderPathExist[index] && dronesDb.SigFolderPathRepeatNum[index] == 1 {
			brand := strings.TrimSpace(dronesDb.Brand[index])
			model := strings.TrimSpace(dronesDb.Model[index])
			freqBand := strings.TrimSpace(dronesDb.FreqBand[index])
			freq := strings.TrimSpace(dronesDb.Freq[index])
			sigFolderName := strings.TrimSpace(dronesDb.SigFolderName[index])
			// 构建文件夹路径 最后再-上 信号文件夹名称
			folderName := brand + "-" + model + "-" + freqBand + "-" + freq + "-" + sigFolderName
			folderPath := filepath.Join(currentDir, "xinhao-test", brand, model, freqBand, freq, sigFolderName)

			// 创建子文件夹
			err = os.MkdirAll(folderPath, 0755)
			if err != nil {
				log.Fatalf("无法创建子文件夹: %v", err) // 先注释，如果还不行，再修改代码
			}
			// 给文件夹关联上
			// 参数1： 被链接的文件 oldname string ; 参数2：链接 newname string. newName 建议用id,因为id比较唯一
			folderNamePath := filepath.Join(folderPath, folderName)

			logrus.Info("创建软链接, oldName=  ", path)
			logrus.Info("创建软链接, newName=  ", folderNamePath)
			err = os.Symlink(path, folderNamePath)
			if err != nil {
				log.Fatalf("无法创建链接: %v", err)
			}

			logrus.Info("成功在当前目录 xinhao-test 下创建子文件夹:  ", folderPath)
		}
	}

	return xinhaoTestPath
}
