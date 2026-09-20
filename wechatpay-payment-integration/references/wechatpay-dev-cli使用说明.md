# wechatpay-dev-cli 使用说明

> 本 Skill 用 CLI search（`wechatpay-dev-cli knowledge search`）拿检索线索；「APIv3 接口动态排障」另需 `api` 系列命令。何时调用见 `SKILL.md` 与 [文档检索与问答](./文档检索与问答.md)；本文件只讲命令怎么跑。CLI 缺参不报错；命令不可用时按下方「失败降级」，不阻塞作答。

## 前置步骤

> 本会话**第一次**跑 CLI 前做一次即可，已确认过就直接用命令。本步只确认 CLI 能跑，**不要**顺带 `Grep` 知识库或列目录。

```bash
wechatpay-dev-cli --version   # 依赖 Node.js ≥ 20，包名 @tenpay/wechatpay-dev-cli
```

- **跑通且 ≥ 基线版本 `1.2.0`（2026-09-14）**：直接用命令，不要反复覆盖安装。
- **跑不通或低于基线**：执行一次 `npm install -g @tenpay/wechatpay-dev-cli@latest`（一律装 `@latest`）后重跑 `--version`。
- **仍失败**：按下方「失败降级」，仅用本地 `Grep` / `Read` 继续，照常作答。

> 基线是本 Skill 发版时 npm 上的 `latest`，本地更高属正常；CLI 发新版后同步更新此处版本与日期。

---

## 规则

1. **【红线】使用 `search` 参数时禁止夹带敏感信息**：APIv2 / APIv3 密钥、API 证书与私钥、`AppSecret` 这类密钥凭据，以及身份证、手机号、银行卡号这类个人敏感信息，**即使用户主动给出也不要带**——换成占位符或直接删掉。
2. **参数尽量传全，缺参不重试**：每次 `search` 都带用户原话、`--query-rewrite`、`--context`、`--model-name`。缺字段 CLI 也会继续执行，**不要为补参重跑同一条命令**。
3. **一轮只调一次**：本轮开头调一次 `search` 拿线索即可，后续检索全在本地 `Grep` / `Read`（见 `SKILL.md`）。

### 失败降级

缺参不是失败。**没返回线索也不是失败**——只是这次没命中，命令已经成功；直接在当前意图的检索路径上本地 `Grep`。

只有命令不可用（`command not found`、未知命令 `knowledge`、超时、进程报错）时，才执行一次 `npm install -g @tenpay/wechatpay-dev-cli@latest` 重跑；仍失败就跳过 CLI，仅用本地 `Grep` / `Read` 继续。

---

## 命令释义 & 示例

### `knowledge search`

拿「该往哪个产品 / 目录查」的线索。**本轮进入本地 Grep / Read 之前必须先跑完本次命令**（或已失败降级）；不要先本地探路再补 CLI search。

```bash
wechatpay-dev-cli knowledge search "<用户问题的原文>" \
  --query-rewrite "<理解后的问题>" \
  --context "<当前项目的背景与任务>" \
  --model-name "<当前模型名称>"
```

| 参数 | 含义 |
| --- | --- |
| 位置参数 `query` | **用户问题的原文**，只按[规则 1](#规则)去掉密钥凭据与个人敏感信息，其余不作改写；含双引号时做转义 |
| `--query-rewrite` | **理解后的问题**（口语→可检索表述、指代消解）。原话中的 URL、接口路径、错误码、字段名原样保留 |
| `--context` | **当前项目的背景与任务**：工程 / 任务上下文，例如「普通商户的 Node + Express 服务，正在接 JSAPI 支付的支付结果通知」。**不是**思考链，也不要传「产品选型 / 答疑与排障」这类分类名 |
| `--model-name` | 执行本 Skill 的**模型名称**，按下方「`--model-name` 怎么填」自报 |

**响应**：返回 `hint` / `keywords` 线索，用来校准 Grep 前缀；没有线索就按无线索本地 `Grep`，**不要**走失败降级。

#### `--model-name` 怎么填

你就是那个模型，直接写出自己的名称，这不属于猜测或编造。格式 `<系列>-<版本>`，例如 `claude-sonnet-4.5`、`gpt-5`、`gemini-2.5-pro`。记得完整版本就写完整；只确定系列就写到系列（如 `claude-sonnet`）。不得编造不存在的模型名，也不得填 `AI`、`assistant`、`模型` 这类无信息值。

---

## 常见问题

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `wechatpay-dev-cli: command not found` | 未安装，或 npm 全局 bin 不在 PATH | 覆盖安装；确认 `npm config get prefix` 下的 bin 已加入 PATH |
| `error: unknown command 'knowledge'` | 本地 CLI 版本过旧，尚无 `knowledge` 命令 | 执行 `npm install -g @tenpay/wechatpay-dev-cli@latest` 升级后重试 |
| 未知选项 `--query-rewrite` / `--context` / `--model-name` | 本地 CLI 版本过旧，不认识这些参数 | 不要当成缺参。升级 CLI 后再试；升级失败则跳过 CLI、仅用本地 `Grep`。|
| 没返回线索（无 `hint` / `keywords`） | 这次没命中目录线索 | **不是失败**。直接在当前意图的检索路径上本地 `Grep` |
| `npm: command not found` | 未装 Node | 安装 Node.js 20+ |
| 安装成功但 `--version` 仍报错 | Node 版本过低 | `node --version` 需 ≥ 20 |
| Windows 下 `api build` 参数异常 | PowerShell 剥引号 | 按排障文档用 `@$env:TEMP\xxx.json` 传 `--params`，勿 inline 复杂 JSON |
| 401 SIGN_ERROR | 非安装问题 | 回到排障文档 Step 2/3，检查 `signMessage` 是否原样签名 |
