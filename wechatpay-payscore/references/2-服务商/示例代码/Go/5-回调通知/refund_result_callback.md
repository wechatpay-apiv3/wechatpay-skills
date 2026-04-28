# 退款结果回调通知（服务商 - Go）

> 内容与 [`Java/5-回调通知/退款结果回调通知说明.md`](../../Java/5-回调通知/退款结果回调通知说明.md) 完全一致；本副本仅为 Go 项目按目录约定查找方便而存在。

> 源文档：[退款结果通知](https://pay.weixin.qq.com/doc/v3/partner/4012586138.md)
> 通用解密 / 验签 / 回包流程：[../../../接入指南/回调处理.md](../../../接入指南/回调处理.md)

## 触发场景

服务商代特约商户调用「申请退款」接口受理后，微信支付分异步处理实际退款，退款进入终态时回推本通知。

## 回调报文骨架

```json
{
  "id": "EV-2018022511223320873",
  "create_time": "2015-05-20T13:29:35+08:00",
  "resource_type": "encrypt-resource",
  "event_type": "REFUND.SUCCESS",
  "summary": "退款成功",
  "resource": {
    "original_type": "refund",
    "algorithm": "AEAD_AES_256_GCM",
    "ciphertext": "<密文，使用服务商 APIv3 密钥解密>",
    "associated_data": "refund",
    "nonce": "<随机串>"
  }
}
```

## event_type 一览

| event_type | 含义 | 处理建议 |
|------------|------|---------|
| `REFUND.SUCCESS` | 退款成功 | 更新业务退款单 |
| `REFUND.CLOSED` | 退款被关闭 | 检查 `error_msg` 与 `refund_status` |
| `REFUND.ABNORMAL` | 退款异常 | 人工介入或调用「异常退款」接口补救 |

## 解密后关键字段

| 字段 | 说明 |
|------|------|
| `sp_mchid` / `sub_mchid` | 服务商号、特约商户号 |
| `out_refund_no` | 商户退款单号（幂等键） |
| `refund_id` | 微信侧退款单号 |
| `out_order_no` | 关联支付分订单 |
| `transaction_id` | 关联的支付单号 |
| `amount.refund` | 实际退款金额（分） |
| `refund_status` | `SUCCESS` / `CLOSED` / `PROCESSING` / `ABNORMAL` |

## 服务商处理要求

- 解密 / 验签均使用 **服务商**密钥与公钥。
- 路由按 `sub_mchid` → 业务幂等按 `out_refund_no`。
- 多次部分退款累加 `amount.refund`，确保不超过 `total_amount`。
- 应答：成功返回 `200 + {"code":"SUCCESS","message":"成功"}`。
