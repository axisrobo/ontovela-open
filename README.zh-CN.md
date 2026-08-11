# ONTOVELA Open

[English](README.md) | 简体中文

ONTOVELA Open 是 [ONTOVELA](https://github.com/axisrobo/ontovela) 的 Apache-2.0 许可公开开发者平台。ONTOVELA 是面向企业、Agent、机器人、流程和数字孪生的 Operational World Model Platform，将物理、数字、组织与运营状态表示为可查询、可重建、可验证的证据化时态世界模型。

## 产品定位

ONTOVELA 不把“最新值”直接当作事实。每一项状态声明都应有来源、证据、事件时间和系统时间，并严格区分：

- `observed`：来自传感器、执行器或可信系统的直接观测
- `reported`：来自外部系统、人员或 Agent 的报告
- `derived`：由确定性规则计算得到的状态
- `inferred`：由模型或推理产生的推断
- `predicted`：面向未来的概率性预测
- `simulated`：来自可能世界或仿真分支的状态

`simulated` 状态不会进入现实的 resolved state 或 Reality Snapshot。

## 本仓库范围

本仓库是 ONTOVELA 的公开采用与集成入口，包含：

- 版本化 API、Schema 和事件契约
- Go、Python、TypeScript SDK
- 示例、参考适配器和本地开发者二进制
- 互操作性与 Reality Integrity 契约测试资产

本仓库不包含：

- ONTOVELA 核心时态状态内核与 reconciliation 实现
- 多租户控制面、跨区域联邦、HA 和商业连接器
- 企业身份集成、合规包和高级支持工具

详细边界见 [`docs/repository-boundary.md`](docs/repository-boundary.md)。

## v0.1 开发者入口

- OpenAPI：[`api/openapi.yaml`](api/openapi.yaml)
- 仓库机器人示例：[`examples/warehouse-robot.md`](examples/warehouse-robot.md)
- 兼容性政策：[`docs/compatibility.md`](docs/compatibility.md)
- Go SDK：[`sdk/go/`](sdk/go/)

## 三仓库边界

| 仓库 | 许可 | 职责 |
| --- | --- | --- |
| [ONTOVELA](https://github.com/axisrobo/ontovela) | AGPL-3.0-or-later | Operational World Model 核心运行时与 Reality Kernel |
| ONTOVELA-open | Apache-2.0 | 公开 API、SDK、示例、适配器和开发者资产 |
| ONTOVELA-ee | AxisRobo Enterprise License | 企业部署、联邦、规模、合规和高级运营能力 |

企业版不会改变公开 assertion、state kind、snapshot digest 或 wire contract 的语义。
