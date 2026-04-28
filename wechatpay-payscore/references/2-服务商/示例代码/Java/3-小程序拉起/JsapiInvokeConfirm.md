# JSAPI / 小程序调起支付分确认订单页（服务商）

> 源文档：[JSAPI调起支付分确认订单页](https://pay.weixin.qq.com/doc/v3/partner/4012607505.md)
> 小程序入口：[wx.openBusinessView](https://pay.weixin.qq.com/doc/v3/partner/4012607510.md)
> APP 端入口：[Android](https://pay.weixin.qq.com/doc/v3/partner/4012607507.md) / [iOS](https://pay.weixin.qq.com/doc/v3/partner/4012607508.md) / [鸿蒙](https://pay.weixin.qq.com/doc/v3/partner/4015271745.md)

## 接入说明

1. 服务商后端调用「创建支付分订单（服务商）」（[CreatePayScoreOrder.java](../1-订单管理/CreatePayScoreOrder.java) / [create_payscore_order.go](../../Go/1-订单管理/create_payscore_order.go)），请求体包含 `sub_mchid`，并将 `need_user_confirm = true`。
2. 接口返回 `package` 字段，由服务商后端组装后下发给前端。
3. 前端通过 `WeixinJSBridge.invoke('openBusinessView', ...)` 或 `wx.openBusinessView(...)` 拉起确认订单页。
4. 用户确认后，微信回调到 **服务商** 的 `notify_url`，服务商按 `sub_mchid` 路由到特约商户业务（参考 [5-回调通知/确认订单回调通知说明.md](../5-回调通知/确认订单回调通知说明.md)）。

## 公众号 H5 示例代码

```javascript
function onBridgeReady() {
  WeixinJSBridge.invoke(
    'openBusinessView',
    {
      businessType: 'wxpayScoreUse',
      queryString:  '<package_in_create_response>'
    },
    function (res) {
      if (res.err_msg === 'open_business_view:ok') {
        // 用户已点击同意，业务成功仍以"确认订单回调通知"或"查询支付分订单"为准
      }
    }
  );
}
if (typeof WeixinJSBridge === 'undefined') {
  document.addEventListener('WeixinJSBridgeReady', onBridgeReady, false);
} else {
  onBridgeReady();
}
```

## 小程序示例代码

```javascript
wx.openBusinessView({
  businessType: 'wxpayScoreUse',
  extraData: {
    // 服务商场景下 package 字段含 sp_mchid + sub_mchid + service_id 等，前端需按 package URL Decode 后逐项填入
    sp_mchid:   'sp_mchid_from_package',
    sub_mchid:  'sub_mchid_from_package',
    service_id: 'service_id_from_package',
    out_order_no:'out_order_no_from_package',
    timestamp:  'timestamp_from_package',
    nonce_str:  'nonce_str_from_package',
    sign_type:  'HMAC-SHA256',
    signature:  'signature_from_package'
  },
  success(res) { /* ... */ },
  fail(err)    { /* ... */ }
});
```

## 注意事项

| 项 | 要求 |
|----|------|
| `signature` 算法 | **服务商 APIv2 密钥 + HMAC-SHA256**（不是子商户的、也不是 APIv3 私钥） |
| `appid` 一致性 | 拉起的 `appid` 必须等于 `CreateServiceOrder` 时的 `sub_appid`（无 sub_appid 时用 `sp_appid`） |
| `package` | 必须取自后端创建订单应答；前端禁止自行拼接 |
| 业务判定 | 前端 `success` 仅代表用户同意 UI；业务成功以回调 / 查单为准 |
