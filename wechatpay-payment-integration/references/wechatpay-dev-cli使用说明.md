# wechatpay-dev-cli 使用说明

> 本 Skill **所有能力、所有对话轮次**都依赖 `wechatpay-dev-cli`（产品选型、示例代码、文档问答、接入质检均需安装）：作答前先用 `knowledge search` 取业务线索；「能力 4 → APIv3 接口动态排障」另需 `api` 系列命令。

## 知识检索命令

```bash
wechatpay-dev-cli knowledge search "<用户问题原文>"
```

- **执行时机**：每轮提问（含追问）在检索、读文档、作答之前执行，**每次对话只执行一次**。
- **检索词**：使用用户本轮输入的原文，不改写、不精简、不翻译；原文含双引号时做转义。
- **报 `error: unknown command 'knowledge'`**：说明本地 CLI 版本过旧，先执行 `npm install -g @tenpay/wechatpay-dev-cli@latest` 升级，再重新执行检索。

---

## 版本与安装

**依赖**：Node.js ≥ 20，包名 `@tenpay/wechatpay-dev-cli`。

首轮对话执行知识检索前，先确认 CLI 可用：

```bash
wechatpay-dev-cli --version
```

能跑通 `--version` 才说明 Node 与 CLI 环境都已就绪——仅确认 `wechatpay-dev-cli` 命令存在不够。

**最新版本为 `1.1.0`**。未安装、或输出不是 `1.1.0` 时，执行一次覆盖安装：

```bash
npm install -g @tenpay/wechatpay-dev-cli@latest
```

---

## 常见问题

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `wechatpay-dev-cli: command not found` | 未安装，或 npm 全局 bin 不在 PATH | 覆盖安装；确认 `npm config get prefix` 下的 bin 已加入 PATH |
| `error: unknown command 'knowledge'` | 本地 CLI 版本过旧，尚无 `knowledge` 命令 | 执行 `npm install -g @tenpay/wechatpay-dev-cli@latest` 升级后重试 |
| `npm: command not found` | 未装 Node | 安装 Node.js 20+ |
| 安装成功但 `--version` 仍报错 | Node 版本过低 | `node --version` 需 ≥ 20 |
| Windows 下 `api build` 参数异常 | PowerShell 剥引号 | 按排障文档用 `@$env:TEMP\xxx.json` 传 `--params`，勿 inline 复杂 JSON |
| 401 SIGN_ERROR | 非安装问题 | 回到排障文档 Step 2/3，检查 `signMessage` 是否原样签名 |
