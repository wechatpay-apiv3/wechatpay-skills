# 确认订单回调通知（商户 - Go）

> 内容与 [`Java/5-回调通知/确认订单回调通知说明.md`](../../Java/5-回调通知/确认订单回调通知说明.md) 完全一致；本副本仅为 Go 项目按目录约定查找方便而存在。

> 源文档：[确认订单回调通知](https://pay.weixin.qq.com/doc/v3/merchant/4012587953.md)
> 通用解密 / 验签 / 回包流程：[../接入指南/回调处理.md](../../../接入指南/回调处理.md)
> 通用签名规则：[../接入指南/签名与验签规则.md](../../../接入指南/签名与验签规则.md)

## 触发场景

用户在「确认订单」页面（小程序 / APP / H5）点击同意后，微信支付分会以 **POST** 方式向商户在创建订单时填写的 `notify_url` 推送本通知，标识订单已变为 `DOING` 状态、可正式发起服务。

## 回调报文骨架

```json
{
  "id": "EV-2018022511223320873",
  "create_time": "2015-05-20T13:29:35+08:00",
  "resource_type": "encrypt-resource",
  "event_type": "PAYSCORE.USER_CONFIRM",
  "summary": "支付分订单用户已确认",
  "resource": {
    "original_type": "payscore",
    "algorithm": "AEAD_AES_256_GCM",
    "ciphertext": "<密文，使用商户 APIv3 密钥 + AEAD_AES_256_GCM 解密后得到 ServiceOrderEntity>",
    "associated_data": "transaction",
    "nonce": "<随机串>"
  }
}
```

## 关键字段

| 字段 | 说明 |
|------|------|
| `event_type` | 固定为 `PAYSCORE.USER_CONFIRM`；用于回调路由判断（与 `PAYSCORE.USER_PAID` / `REFUND.SUCCESS` 区分） |
| `resource.algorithm` | 固定 `AEAD_AES_256_GCM` |
| `resource.ciphertext` | 解密后为商户视角的支付分订单实体，包含 `out_order_no`、`service_id`、`appid`、`mchid`、`state`、`openid`、`order_id`、`risk_fund`、`time_range`、`location` 等字段 |
| 解密所得 `state` | `DOING`（用户已确认，可开始提供服务） |

## 商户处理要求

1. **验签**：使用 `Wechatpay-Signature` / `Wechatpay-Timestamp` / `Wechatpay-Nonce` / `Wechatpay-Serial` 头，配合微信支付公钥校验请求体；任一不匹配立即返回 `401`。
2. **解密**：用商户 APIv3 密钥 + AEAD_AES_256_GCM 解密 `resource.ciphertext`。
3. **路由**：以 `event_type` 区分确认 / 支付 / 退款回调；以 `out_order_no` 入库去重（`UNIQUE INDEX(out_order_no, event_type)`）。
4. **入库**：将订单状态更新为 `DOING`，登记 `openid`、`order_id`，触发后续业务（如开门、放行、发货）。
5. **应答**：处理成功返回 `200 + {"code":"SUCCESS","message":"成功"}`；业务异常返回 `500 + {"code":"FAIL","message":"<原因>"}` 触发重试。
6. **幂等**：同一 `out_order_no + event_type` 多次到达必须只生效一次。

## 测试要点

- 主动调用「查询支付分订单」接口与回调入库结果做交叉校验，避免遗漏。
- 模拟回调延迟 / 重试场景（处理超时不影响业务）。
