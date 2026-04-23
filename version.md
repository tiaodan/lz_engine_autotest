# v0.0.0.1
- 优化后，更好用了

# v0.0.0.2
- 稳定些，excel,通过 “是否优秀数据”进行区分

# v0.0.0.3
AI 改前提交

# v0.0.0.4
优化代码：删除dronesdbenable = false 的大量逻辑判断，此配置弃用

# v0.0.0.5
新增定时任务功能：
- 在 config.ini 中添加 [scheduler] 配置节
- 配置项：enable(开关)、hour(小时)、minute(分钟)、command(执行命令)
- 示例：每天16点执行命令6（一键执行步骤5、4）

# v0.0.0.6
优化: 配置文件更清晰

# v0.0.0.7
优化：能往8001端口发送数据(根据excel)
修复端口配置相关bug：
1. 修复checkAndSwitchPort()缩进问题（之前在if块外，现在正确放在if块内）
2. 在sendTask开始时，先检查第一个信号文件夹的端口配置，使用正确的端口初始化连接
3. 添加调试日志，打印发送字节数和计算出的间隔时间（ms）

# v0.0.0.8
修bug：main.go util.go 2个语法错误

# v0.0.0.9
优化文件结构

# v0.0.0.10
修复盘符替换bug：
- 正则从 `[E]:` 改为 `[A-Za-z]:`
- 之前只替换E盘，Excel里写D盘的路径不会被替换成配置的driveLetter
- 现在任何盘符都会被替换成配置的driveLetter(E)
修复config.ini格式：保留注释、大小写格式、行内注释


# v0.0.0.11 
修复两个bug：
- 盘符替换：正则从 `[E]:` 改为 `[A-Za-z]:`，现在任何盘符都能替换成配置的driveLetter
- report日志重复：删除report()函数中重复调用的createReport()