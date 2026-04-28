# APP 调起支付分（服务商）

> 源文档：
> - Android 确认订单页：[4012607507](https://pay.weixin.qq.com/doc/v3/partner/4012607507.md)
> - iOS 确认订单页：[4012607508](https://pay.weixin.qq.com/doc/v3/partner/4012607508.md)
> - 鸿蒙 确认订单页：[4015271745](https://pay.weixin.qq.com/doc/v3/partner/4015271745.md)
> - Android 订单详情页：[4012607513](https://pay.weixin.qq.com/doc/v3/partner/4012607513.md)
> - iOS 订单详情页：[4012607514](https://pay.weixin.qq.com/doc/v3/partner/4012607514.md)
> - 鸿蒙 订单详情页：[4015271776](https://pay.weixin.qq.com/doc/v3/partner/4015271776.md)

## 接入说明

APP 端调起的参数由 **服务商后端**根据创建订单应答的 `package` 字段拼接、并按 **服务商 APIv2 密钥 + HMAC-SHA256** 计算 `signature`，下发给 APP。

APP 端通过微信开放平台 SDK 的 `WXOpenBusinessView`（Android） / `WXOpenBusinessViewReq`（iOS） 调起。

## Android 关键代码

```java
WXOpenBusinessView req = new WXOpenBusinessView();
req.businessType = "wxpayScoreUse"; // 详情页用 wxpayScoreDetail
req.query        = "sp_mchid=xxx&sub_mchid=xxx&service_id=xxx&out_order_no=xxx&timestamp=xxx&nonce_str=xxx&sign_type=HMAC-SHA256&signature=xxx";
req.extInfo      = "{\"miniProgramType\":0}";
api.sendReq(req);
```

## iOS 关键代码

```objective-c
WXOpenBusinessViewReq *req = [[WXOpenBusinessViewReq alloc] init];
req.businessType = @"wxpayScoreUse"; // 详情页用 wxpayScoreDetail
req.query       = @"sp_mchid=xxx&sub_mchid=xxx&service_id=xxx&out_order_no=xxx&timestamp=xxx&nonce_str=xxx&sign_type=HMAC-SHA256&signature=xxx";
req.extInfo     = @"{\"miniProgramType\":0}";
[WXApi sendReq:req completion:nil];
```

## 注意事项

| 项 | 要求 |
|---|---|
| 微信版本 | Android ≥ 7.0.5、iOS ≥ 7.0.5 |
| `appid` | APP 注册的 `appid` 必须等于服务端创建订单时的 `sub_appid`（无 sub_appid 时使用 `sp_appid`） |
| 业务结果 | APP 回调 `errCode == 0` 仅代表用户操作完成，业务成功以「确认订单回调」或「查询支付分订单」为准 |
| 鉴权失败 | `SIGN_ERROR` 多由"误用 APIv3 私钥代替 APIv2 密钥"导致，详见 [签名与验签规则.md](../../../接入指南/签名与验签规则.md) |
