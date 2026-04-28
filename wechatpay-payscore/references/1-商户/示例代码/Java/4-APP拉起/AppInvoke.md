# APP 调起支付分（商户）

> 源文档：
> - Android 确认订单页：[4012587909](https://pay.weixin.qq.com/doc/v3/merchant/4012587909.md)
> - iOS 确认订单页：[4012596359](https://pay.weixin.qq.com/doc/v3/merchant/4012596359.md)
> - 鸿蒙 确认订单页：[4015271805](https://pay.weixin.qq.com/doc/v3/merchant/4015271805.md)
> - Android 订单详情页：[4012587980](https://pay.weixin.qq.com/doc/v3/merchant/4012587980.md)
> - iOS 订单详情页：[4012596423](https://pay.weixin.qq.com/doc/v3/merchant/4012596423.md)
> - 鸿蒙 订单详情页：[4015271812](https://pay.weixin.qq.com/doc/v3/merchant/4015271812.md)

## 接入说明

APP 端调起的参数与公众号 / 小程序 一致，由 **商户后端** 根据「创建支付分订单」应答的 `package` 字段拼接、并按 **APIv2 商户密钥 + HMAC-SHA256** 计算 `signature`，下发给 APP。

APP 端通过微信开放平台 SDK 的 `WXOpenBusinessView`（Android）/ `WXOpenBusinessViewReq`（iOS） 调起。

## Android 关键代码

```java
WXOpenBusinessView req = new WXOpenBusinessView();
req.businessType = "wxpayScoreUse"; // 详情页用 wxpayScoreDetail
req.query = "mch_id=xxx&service_id=xxx&out_order_no=xxx&timestamp=xxx&nonce_str=xxx&sign_type=HMAC-SHA256&signature=xxx";
req.extInfo = "{\"miniProgramType\":0}";
api.sendReq(req);
```

## iOS 关键代码

```objective-c
WXOpenBusinessViewReq *req = [[WXOpenBusinessViewReq alloc] init];
req.businessType = @"wxpayScoreUse"; // 详情页用 wxpayScoreDetail
req.query       = @"mch_id=xxx&service_id=xxx&out_order_no=xxx&timestamp=xxx&nonce_str=xxx&sign_type=HMAC-SHA256&signature=xxx";
req.extInfo     = @"{\"miniProgramType\":0}";
[WXApi sendReq:req completion:nil];
```

## 注意事项

| 项 | 要求 |
|---|---|
| 微信版本 | Android ≥ 7.0.5、iOS ≥ 7.0.5、鸿蒙 ≥ 微信支付 SDK 最新版 |
| `appid` | 必须与商户在 `CreateServiceOrder` 时使用的 `appid` 完全一致 |
| 业务结果 | APP 端回调 `errCode == 0` 仅表示用户操作完成，业务成功仍以「确认订单回调」或「查询支付分订单」为准 |
| 鉴权失败 | 出现 `SIGN_ERROR` 多由"误用 APIv3 私钥代替 APIv2 密钥"导致，参考 [签名与验签规则.md](../../../接入指南/签名与验签规则.md) |
