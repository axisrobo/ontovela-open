# ONTOVELA

[English](README.md) | 简体中文

**Digital Enterprise Twin & Operational World Model Platform**

ONTOVELA 将物理、数字、组织、流程、Agent 与机器人的状态融合为证据化的时态世界模型，为规划、仿真、干预与闭环自治提供统一的企业现实底座。

## ONTOVELA 解决的问题

自主系统与企业系统常常依赖分散、陈旧或不可验证的“最新值”。规划者无法区分观测与推断、预测与仿真，也无法判断一个事实比另一个更新。ONTOVELA 让企业现实**可计算**：可查询、可重建、可用于决策，同时严格区分观测、推断、预测与仅存在于仿真中的状态。

它用一条可被规划者、执行器、机器人与人工治理者共同信任的权威运营现实平面，取代碎片化的 IoT 看板、CMDB、知识图谱与私有 Agent 状态。

## 核心特性

- **证据优先**：每条状态声明都带来源、证据引用、事件时间、系统时间、状态种类与置信度。
- **六类状态严格分离**：`observed` / `reported` / `derived` / `inferred` / `predicted` / `simulated` 永不混淆；`simulated` 永不进入真实运营状态。
- **双时态历史**：`as_of`（事件时间）与 `as_known`（系统获知时间）可重建任意时刻发生的事实与系统当时所知道的内容。
- **可解释解析**：按来源权威度解析，冲突显式返回 `unknown` / `conflicted`，绝不静默覆盖。
- **Purpose-bound Reality View**：规划者按目的申请最小必要状态，带逐属性新鲜度门槛与来源存活（heartbeat）检查。
- **签名快照**：不可变、可验证的现实切片，供审计、计划签发与 PEIRAVELA 仿真分叉，支持 diff 与策略变化检测。
- **影响、因果谱系与分析**：沿依赖关系与仅限 `causes` 的关系做影响与因果推理。
- **sim-to-real 对比**：比较真实解析状态与仿真分支，且不产生污染。
- **持久订阅**：变更流、消费者游标、订阅定义、过滤与审计导出。
- **来源问责**：租户隔离、来源绑定、幂等写入，并支持可选的身份主体。
- **Go 后端 + PostgreSQL 持久化**：append-only 双时态账本与内嵌迁移。

## 本仓库：ONTOVELA Open

[ONTOVELA](https://github.com/axisrobo/ontovela) 的 Apache-2.0 许可公开开发者平台，是公开采用入口，提供稳定 API 契约、SDK、示例、参考适配器与本地开发者二进制。

包含：

- 版本化 API、Schema 与事件契约
- Go、Python、TypeScript SDK
- 本地开发者二进制与 Docker 快速开始
- assertion、时态查询、快照、订阅与 Reality View 示例
- HTTP/webhook、Kafka/NATS、SQL/REST、MQTT、ROS 2、OPC UA 与 Harmovela 参考适配器
- 契约兼容性与 Reality Integrity 测试资产

不包含（位于核心或企业仓库）：

- ONTOVELA 核心时态状态内核与 reconciliation 实现
- 多租户控制面、跨区域联邦、HA 与商业连接器
- 企业身份集成、合规包与高级支持工具

详细边界见 [`docs/repository-boundary.md`](docs/repository-boundary.md)。

## 快速开始与契约

- OpenAPI：[`api/openapi.yaml`](api/openapi.yaml)
- 快速开始：[`docs/quickstart.md`](docs/quickstart.md)
- 兼容性政策：[`docs/compatibility.md`](docs/compatibility.md)
- 参考场景：[`examples/warehouse-robot.md`](examples/warehouse-robot.md)、[`examples/incident-response.md`](examples/incident-response.md)、[`examples/supply-chain-counterfactual.md`](examples/supply-chain-counterfactual.md)
- 发布：GitHub Release 附带 Windows 开发者二进制。

## 三仓库边界

| 仓库 | 许可 | 职责 |
| --- | --- | --- |
| [ONTOVELA](https://github.com/axisrobo/ontovela) | AGPL-3.0-or-later | Operational World Model 核心运行时与 Reality Kernel |
| ONTOVELA-open | Apache-2.0 | 公开 API、SDK、示例、适配器和开发者资产 |
| ONTOVELA-ee | AxisRobo Enterprise License | 企业部署、联邦、规模、合规和高级运营能力 |

企业版不会改变公开 assertion、state kind、snapshot digest 或 wire contract 的语义。
