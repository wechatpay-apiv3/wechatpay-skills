# wechatpay-dev-cli 使用说明

> 本 Skill在 **「能力 4 → APIv3 接口动态排障」** 分支依赖 `wechatpay-dev-cli`。  
> 产品选型、示例代码、文档问答、接入质检 **不需要** 安装 CLI。

## 检测 CLI 是否可用

进入 [APIv3接口动态排障](./APIv3接口动态排障.md) 之前，在终端执行：

```bash
wechatpay-dev-cli --version
```

正常时应输出版本号，例如：`1.0.0`
能跑通 `--version` 才说明 `Node` 环境、`wechatpay-dev-cli` 环境已经准备好；仅知道 `wechatpay-dev-cli` 这个命令名存在不够。

---

## 安装

**依赖**：Node.js ≥ 20（包名 `@tenpay/wechatpay-dev-cli`）。

```bash
npm install -g @tenpay/wechatpay-dev-cli
wechatpay-dev-cli --version
```

---

## 使用时的常见问题

| 现象 | 可能原因 | 处理 |
|------|----------|------|
| `wechatpay-dev-cli: command not found` | 未安装或 npm 全局 bin 不在 PATH | `npm install -g @tenpay/wechatpay-dev-cli`，确认 `npm config get prefix`/bin 已加入 PATH |
| `npm: command not found` | 未装 Node | 安装 Node.js 20+ |
| 安装成功但 `--version` 仍报错 | Node 版本过低 | `node --version` 需 ≥ 20 |
| Windows 下 `api build` 参数异常 | PowerShell 剥引号 | 排障文档要求用 `@$env:TEMP\xxx.json` 传 `--params`，勿 inline 复杂 JSON |
| 401 SIGN_ERROR | 非安装问题 | 回到排障文档 Step 2/3，检查 `signMessage` 是否原样签名 |
