# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

无人机信号检测引擎自动化测试工具 (`lz_engine_autotest`)。

**核心功能**：
1. 从机型库 Excel 读取信号包信息，创建软链接
2. 通过 TCP 发送 RF 信号包（.dat/.bvsp 文件）到检测设备
3. 通过 GraphQL API 查询检测结果
4. 生成分析报告，对比预期与实际检测结果

## Build Commands

```bash
go build .                    # 编译生成 engine_autotest.exe
go run .                      # 直接运行（编译+运行）
```

**注意**：Windows 下需要管理员权限运行，因为创建软链接需要管理员权限。

## Running the Application

程序是交互式菜单：

```
1 - ready    # 准备阶段：读取机型库、创建软链接、生成待发送列表
2 - feed     # 回放信号：发送信号+查询检测结果
3 - report   # 生成分析报告
4 - 一键执行步骤123
5 - delete history file
6 - 一键执行 步骤5、4
7 - 并发发送信号
0 - 退出
```

## Configuration

配置文件 `config.ini` (INI 格式)：

| Section | Key | 说明 |
|---------|-----|------|
| `[network]` | `devip` | 设备 IP |
| `[signal]` | `sigdir` | 信号包根目录 |
| `[signal]` | `sigpkgsendinterval` | 发送间隔(ms/MB) |
| `[signal]` | `cdfolderinterval` | 切换文件夹等待时间(秒) |
| `[query]` | `noquerytimes2nextsig` | 查不到多少次后跳到下一个信号 |
| `[query]` | `mistakefreq` | 频率误差容忍值(MHz) |
| `[dronesdb]` | `dronesdbpath` | 机型库 Excel 路径 |
| `[dronesdb]` | `alldronesdbpath` | 完整机型库路径 |
| `[log]` | `loglevel` | 日志级别: debug/info/error |

## Architecture

### Workflow

```
ready() → feed() → report()
   │         │        │
   │         │        └── 生成分析报告 Excel
   │         │
   │         └── sendTask() + queryTask() 并发运行
   │
   └── 读取机型库、创建软链接、生成待发送列表
```

### Key Files

| File | Purpose |
|------|---------|
| `main.go` | 程序入口、全局变量、DroneDB 结构体 |
| `send.go` | TCP 信号发送逻辑 |
| `query.go` | GraphQL 查询逻辑、检测算法 |
| `ready.go` | 准备阶段：读取配置、创建软链接 |
| `report.go` | 生成分析报告 |
| `util.go` | 工具函数、Excel 读写 |

### Key Data Structures

```go
type Drone struct {
    Name     string   // 机型名称
    FreqList int      // 频率 (kHz)
    Id       []string // 无人机 ID (支持多个)
}

type DroneDB struct {
    // Excel 列: A-N, P-Q
    Id, Manufacture, Brand, Model, Protocol, Subtype, FreqBand, Freq []string
    SigFolderName, SigFolderPath []string
    DroneTxt, DroneIdTxt []string     // 机型.txt, id.txt 内容
    SigFolderReplayNum []string       // P列: 回放次数
    ReplayPort []string               // Q列: 回放端口
}
```

### 机型库 Excel 列顺序

| 列 | 字段 |
|----|------|
| A | ID |
| B | 厂商 |
| C | 品牌 |
| D | 型号 |
| E | 协议 |
| F | 子类型 |
| G | 频段 |
| H | 频率 |
| I | 信号文件夹名称 |
| J | 信号文件夹路径 |
| L | 机型.txt 内容 |
| M | id.txt 内容 |
| O | seafile 链接 |
| P | 回放次数 |
| Q | 回放端口 |

### Detection Algorithm (query.go)

判断"检测到"的条件：
1. **机型名称相等** - 不区分大小写
2. **ID 匹配** - 查到的 ID 在 id.txt 列表中，或 id.txt 内容为"随机"
3. **频率匹配** - 无误差或误差在 `mistakefreq` MHz 内

### Signal Package Requirements

信号包目录结构：
```
信号文件夹/
├── *.dat 或 *.bvsp    # 信号文件（按文件名数字排序）
├── id.txt             # 可选：无人机 ID，多个用 / 分隔，或写"随机"
└── 机型.txt           # 可选：格式 "机型名称:频率"，如 "DJI Mavic2:2409500"
```

## Output Files

程序生成带时间戳的 Excel 文件：
- `待发送列表-{timestamp}.xlsx` - 待发送信号列表
- `查询列表{timestamp}.xlsx` - 查询结果记录
- `分析报告{timestamp}.xlsx` - 最终分析报告

## Dependencies

- `github.com/sirupsen/logrus` - 日志
- `github.com/spf13/viper` - 配置管理
- `github.com/xuri/excelize/v2` - Excel 操作
- `github.com/thoas/go-funk` - 函数工具

## Important Notes

1. **管理员权限**: Windows 下创建软链接需要管理员权限
2. **切换文件夹时间**: `cdfolderinterval` 应比设备 TTL 大约 4 秒
3. **查询间隔**: `querydroneinterval=2` 配合 `afterqueriedwaittimes=6` 效果较好
4. **端口功能**: Q 列可配置回放端口，切换信号包时自动重连