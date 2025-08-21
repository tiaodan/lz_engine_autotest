#!/bin/bash
# 功能: 同时回放多个信号
# 使用方式：./run_feed.sh user task ip num -> 使用人名称 所属任务名称 设备ip 最大并发数
# eg. ./run_feed.sh user1 testEngine 192.168.85.239 5  -> 解释: user1这个用户，为了测试引擎，给85.239这个ip，同时回放5个信号
# 待办
# ---------- 修改feed.py 生成的命令里带user + task,方便pkill

# 配置目录,不带/
xinhao=/home/xinhao
# 测试并发时，的信号目录。回放信号一般是：feed.py -d /home/xinhao-all/900/xx.bvsp ，
# 因此 concurrenceXinhaoPath, 填 xinhao-all前面就行，不用带/ , 即 /home
concurrenceXinhaoPath=/home

# 变量
cmd=$1
user=$2
task=$3
ip=$4
num=$5
echo "输入num = $num"

# 封装检测缪路函数，目录下不存在 run_feed.sh、feed_with_user_task.py、feed.py 和信号文件就报错
checkFolder() {
	# 1 判断常用的5个信号, 脚本文件是否有
	if [[ ! -f "$xinhao/run_feed.sh" || ! -f "$xinhao/feed_with_user_task.py" || ! -f "$xinhao/feed.py" ]];then
		echo "配置目录=========== $xinhao ,无脚本文件, 看看目录对不对 ！！！"
		echo "配置目录=========== $xinhao ,无脚本文件, 看看目录对不对 ！！！"
		echo "配置目录=========== $xinhao ,无脚本文件, 看看目录对不对 ！！！"
	fi
	
	# 2 判断并发目录是否存在 
	if [[ ! -d "$concurrenceXinhaoPath" ]];then
		echo "配置目录=========== $concurrenceXinhaoPath ,目录不存在, 看看目录对不对 ！！！"
	fi
}

# 封装start函数
start_feed() {
	# 变量
	local user=$1
	local task=$2
	local ip=$3
	local num=$4
	
	# 定义5个命令. 并发的所有信号，已经排除了前5个信号
	cmds=(
		#echo "date"
		#echo "ll" 
		#echo "df -h"  -t 250 

		"nohup ./feed_with_user_task.py -i $ip -t 28  -d $xinhao/433M/433/459/ -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28  -d $xinhao/900M/AEE_F100_900/drone/ -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28  -d $xinhao/1400M/入云龙1550/1440_28ms_2/ -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28  -d $xinhao/2400M/DJI\(大疆\)\(报文\)-Air_2S/2412.5-20M/ -p 8001 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28  -d $xinhao/5800M/X7Pro/5800/5180/ -ut '$user $task' >/dev/null 2>&1 &"
		
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/Feima\(飞马\)/F1000/900/840 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/Feima\(飞马\)/D500/900/840 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/Hikvisio\(海康\)/4080a/900/915 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/Hikvisio\(海康\)/4080a/1400/1429 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/Hikvisio\(海康\)/6150a/900/915 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/Hikvisio\(海康\)/6150b/900/840 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/CW\(纵横\)/Cw15/2400/2444 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/CW\(纵横\)/Cw15/900/915 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/KWT\(科卫泰\)/general/2400/2426 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/DAGONG\(大工\)/1400/1495 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/Xiaomi\(小米\)/4k/5800/5765 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 28 -d $concurrenceXinhaoPath/xinhao-all/Xiaomi\(小米\)/mitu/5800/5745 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/PARROT\(派诺特\)/anafi/2400/2422 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/PARROT\(派诺特\)/bebop/5800/5765 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/heiying1/2400/2384 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/heiying2/2400/2384 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/zino2/5800/5825 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/zino_pro/5800/5745 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/zino/5800/5745 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/h107s/2400/2450 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/h507a/2400/2462 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/h216a/2400/2457 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Hubsan\(哈博森\)/ace/2400/2400 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/AEE\(一电\)/Sparrow2/5800/5745 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/AEE\(一电\)/Ap11/2400/2437 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/HOLYSTONE/hs100g/5800/5200 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/WALKERA\(华科尔\)/vitus/5800/5825 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNEEC\(昊翔\)/mantis_q/5800/5745 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNEEC\(昊翔\)/q500/24G_RC/2400/2435 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNEEC\(昊翔\)/q500/58G_WIFI/5800/5745 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNK\(云科\)/f11/5800/5200 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNK\(云科\)/s6/2400/2417 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNK\(云科\)/s15/2400/2417 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNK\(云科\)/bugs5w/5800/5180 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNK\(云科\)/s169/2400/2412 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNK\(云科\)/s20w/2400/2417 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/YUNK\(云科\)/s13/2400/2417 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/NUOBAMAN\(诺巴曼\)/x7/2400/2417 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/CFLY\(畅天游\)/Faith2Pro/2400/2384 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/CFLY\(畅天游\)/Faith2/2400/2462 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/PowerVision\(臻迪\)/poweregg/5800/5745 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Feipai\(飞拍\)/6KPro/5800/5745 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/GDU\(普宙\)/S220/2400/2385 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/GDU\(普宙\)/O2/5800/5825 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/GDU\(普宙\)/S400/2400/2390 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/GDU\(普宙\)/Byrd/2400/2427 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/MJX\(美嘉欣\)/bug3/2400/2449 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Upair\(追云\)/one/2400/2463 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/WLTOY/x1s/5800/5180 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/BaYang\(巴阳\)/X16/2400/2432 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/GoPro/Karma/2400/2437 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/CHENGXING\(橙星\)/Cx10w/2400/2447 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/BEAST/3E/2400/2422 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/JYMIND/5800/5200 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/MaiKenQi\(麦肯奇\)/p8/2400/2427 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/XINGSHETU\(行摄图\)/X10pro/5800/5805 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/4DRC/F13/2400/2467 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/M8/GENERAL/5800/5180 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/LAUMOX/x35pro/5800/5180 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/FUTABA/14sg/2400/2430 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/JJRC/i12/2400/2475 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/JJRC/h31/2400/2418 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/WFLY/wft09sii/2400/2449 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/WFLY/et12/2400/2450 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/RadioLink/at9s/2400/2459 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/RadioLink/t8fb/2400/2440 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/MICROZONE/mc8b/2400/2472 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/TFMODEL/Tf8g/2400/2469 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/XK/Detect-x380/2400/2445 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/FOXEER/Dragonlink-Video/900/918 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/EZUHF/Tx/433/453 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Skydroid\(云卓\)/h12/2400/2504 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/CUAV/Sx10K/900/915 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/TW/433/433 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/TBS\(黑羊\)/Crossfire/900/868-GSFK -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/TBS\(黑羊\)/Crossfire/2400/2423 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/WALKSNAIL/5800/5735 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Skydio/Skydio2/5800/5805 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/EHANG/bluebox/2400/2468 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Skydio/Skydio2plus/5800/5805 -ut '$user $task' >/dev/null 2>&1 &"
		"nohup ./feed_with_user_task.py -i $ip -t 3000 -d $concurrenceXinhaoPath/xinhao-all/Xiaomi\(小米\)/FeiMiX8Pro2025/2400/2326 -ut '$user $task' >/dev/null 2>&1 &"

	)
	
	# 判断变量不为空，否则退出程序
	if [[ -z "$user" || -z "$task" || -z "$ip" || -z "$num" ]]; then
		echo "错误：参数不能为空！"
		echo "用法：$0 start <user> <task> <ip> <num>"
		echo "用法：$0 stop <user task>"
		exit 1
	fi
	# 检查 num 是否为数字（可选）
	#if ! [[ "$num" =~ ^[1-5]+$ ]]; then
	cmdsLen=${#cmds[@]} 
	
	# 如果输入同时回放个数 >5 <=总个数，打印这个
	if [[ $num -gt 5 && $num -le $cmdsLen  ]];then
		echo "总可用信号个数= $cmdsLen"
	fi
	
	# 之前写法 if ! [[ "$num" =~ ^([1-5]|433|900|1400|2400|5800)$ ]]; then 
	# if ! [[ "$num" =~ ^([1-$cmdsLen]|433|900|1400|2400|5800)$ ]]; then， 这样写法不对 不对
	#if [[ $num -lt 1 || $num -gt $cmdsLen || ! "$num" =~ ^(433|900|1400|2400|5800)$ ]];then 不对
	if [[ $num -lt 1 || $num -gt $cmdsLen ]];then
		if ! [[ "$num" =~ ^(433|900|1400|2400|5800)$ ]];then
			echo "错误：num 必须是一个数字[1-$cmdsLen]！或者频段 433 2400 这种"
			exit 1
		fi
	fi
	
	# 原先写法 if [[ $num =~ ^[1-5]$ ]];then
	# if [[ $num =~ ^[1-$cmdsLen]$ ]];then 不对，没想好怎么使用变量
	if [[  $num -gt 1 && $num -le $cmdsLen ]];then
		for ((i=0;i<$num;i++));do
			echo "-------- 传参 ip = $ip"
			echo "执行命令 $((i+1)): ${cmds[i]}"
			eval "${cmds[i]}"
		done
	elif [[ $num == 433 ]];then
		echo "执行命令 $((i+1)): ${cmds[0]}"
		eval "${cmds[0]}"
	elif [[ $num == 900 ]];then
		echo "执行命令 $((i+1)): ${cmds[1]}"
		eval "${cmds[1]}"
	elif [[ $num == 1400 ]];then
		echo "执行命令 $((i+1)): ${cmds[2]}"
		eval "${cmds[2]}"
	elif [[ $num == 2400 ]];then
		echo "执行命令 $((i+1)): ${cmds[3]}"
		eval "${cmds[3]}"
	elif [[ $num == 5800 ]];then
		echo "执行命令 $((i+1)): ${cmds[4]}"
		eval "${cmds[4]}"
	else 
		echo "输入错误,退出"
		exit
	fi
}

# 封装stop函数
stop_feed() {
	local user=$1
	local task=$2
	
	echo "stop 传参 $user $task"
	# 判断变量不为空，否则退出程序
	if [[ -z "$user" || -z "$task" ]]; then
		echo "错误：参数不能为空！"
		echo "用法：$0 stop <user task>"
		exit 1
	fi
	echo "pkill -f '$user $task'"
	pkill -f "$user $task"
}


# 主逻辑
main() {
	# 脚本加权限
	chmod +x feed.py
	chmod +x run_feed.sh
	chmod +x feed_with_user_task.py
	
	case $1 in
		start)
			if [[ $# -ne 5 ]]; then
				echo "错误：参数数量不正确！"
				echo "用法：$0 start <user> <task> <ip> <num>"
				exit 1
			fi
			
			#user=$2
			#task=$3
			#ip=$4
			#num=$5
			
			start_feed "$user" "$task" "$ip" "$num"
			;;
		stop)
			# stop_feed "$@" 这样写错误，传参不对应
			echo "命令行 stop 传参 $user $task"
			stop_feed "$user" "$task"
			;;
		*)
			echo "用法：$0 {start|stop}"
			echo "start: 启动信号回放任务"
			echo "  示例: $0 start user1 testEngine 192.168.85.239 5"
			echo "  示例-只回放一个频段: $0 start user1 testEngine 192.168.85.239  <433 900 1400 2400 5800> <>里随便输入一个频段"
			echo "  示例-同时回放一个频段: $0 start user1 testEngine 192.168.85.239  433 900 1400 2400 5800"
			echo "stop:  停止指定信号回放任务"
			echo "  示例: $0 stop username taskname"
			exit 1
			;;
	esac
}


# 调用,加上￥@,才能确保，传参不丢失
main "$@"