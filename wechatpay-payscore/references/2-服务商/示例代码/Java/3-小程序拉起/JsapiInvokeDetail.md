# JSAPI / 小程序调起支付分订单详情页（服务商）

> 源文档：[JSAPI调起支付分订单详情页](https://pay.weixin.qq.com/doc/v3/partner/4012607518.md)
> 小程序入口：[wx.openBusinessView](https://pay.weixin.qq.com/doc/v3/partner/4012607516.md)
> APP 端入口：[Android](https://pay.weixin.qq.com/doc/v3/partner/4012607513.md) / [iOS](https://pay.weixin.qq.com/doc/v3/partner/4012607514.md) / [鸿蒙](https://pay.weixin.qq.com/doc/v3/partner/4015271776.md)

## 接入说明

服务商在用户使用过程或订单完结后，让用户在微信内查看支付分订单详情时，可调起「订单详情页」。`query_string` 由 **服务商后端** 拼接并使用 **服务商 APIv2 密钥 + HMAC-SHA256** 签名，详见 [签名与验签规则.md](../../../接入指南/签名与验签规则.md) 中「客户端拉起签名（V2）」小节。

公众号 H5 / 小程序 / APP 端代码与「确认订单页」相同，仅 `businessType` 改为 `wxpayScoreDetail`：

```javascript
WeixinJSBridge.invoke(
  'openBusinessView',
  { businessType: 'wxpayScoreDetail', queryString: '<query_string>' },
  function (res) { /* ... */ }
);
```

## 注意事项

- `appid` 必须与「创建支付分订单（服务商）」时的 `sub_appid` 一致。
- 详情页只读不可改：用户的修改 / 退款行为请走服务商后台对应接口。
- 仅当订单已存在时可调起。
