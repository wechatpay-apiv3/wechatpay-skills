# JSAPI / 小程序调起支付分确认订单页（商户）

> 源文档：[JSAPI调起支付分确认订单页](https://pay.weixin.qq.com/doc/v3/merchant/4012587945.md)
> 小程序入口：[wx.openBusinessView](https://pay.weixin.qq.com/doc/v3/merchant/4012587949.md)
> APP 端入口：[Android](https://pay.weixin.qq.com/doc/v3/merchant/4012587909.md) / [iOS](https://pay.weixin.qq.com/doc/v3/merchant/4012596359.md) / [鸿蒙](https://pay.weixin.qq.com/doc/v3/merchant/4015271805.md)

## 接入说明

1. 商户后端调用「创建支付分订单」（[CreatePayScoreOrder.java](../1-订单管理/CreatePayScoreOrder.java) / [create_payscore_order.go](../../Go/1-订单管理/create_payscore_order.go)），并将请求中 `need_user_confirm` 设为 `true`。
2. 接口返回 `package` 字段（形如 `mch_id=...&service_id=...&out_order_no=...&timestamp=...&nonce_str=...&sign_type=HMAC-SHA256&signature=...`），由商户后端组装后下发到前端。
3. 前端通过 `WeixinJSBridge.invoke('openBusinessView', ...)`（公众号 H5）或 `wx.openBusinessView(...)`（小程序）拉起确认订单页。
4. 用户确认后，微信会回调商户的 `notify_url`（参考 [5-回调通知/确认订单回调通知说明.md](../5-回调通知/确认订单回调通知说明.md)）。

## 公众号 H5 示例代码

```javascript
function onBridgeReady() {
  WeixinJSBridge.invoke(
    'openBusinessView',
    {
      businessType: 'wxpayScoreUse',
      queryString:  '<package_in_create_response>'   // 后端 CreateServiceOrder 应答中的 package 字段
    },
    function (res) {
      if (res.err_msg === 'open_business_view:ok') {
        // 用户已点击同意，前端不要直接据此判断业务成功
        // 必须以「确认订单回调通知」或主动「查询支付分订单」为准
      } else {
        // open_business_view:cancel 用户取消；其它为异常
      }
    }
  );
}

if (typeof WeixinJSBridge === 'undefined') {
  if (document.addEventListener) {
    document.addEventListener('WeixinJSBridgeReady', onBridgeReady, false);
  } else if (document.attachEvent) {
    document.attachEvent('WeixinJSBridgeReady', onBridgeReady);
    document.attachEvent('onWeixinJSBridgeReady', onBridgeReady);
  }
} else {
  onBridgeReady();
}
```

## 小程序示例代码

```javascript
wx.openBusinessView({
  businessType: 'wxpayScoreUse',
  extraData: {
    // package 字段需 URL Decode 后逐项填入：mch_id / service_id / out_order_no / timestamp / nonce_str / sign_type / signature
    mch_id:      'mch_id_from_package',
    service_id:  'service_id_from_package',
    out_order_no:'out_order_no_from_package',
    timestamp:   'timestamp_from_package',
    nonce_str:   'nonce_str_from_package',
    sign_type:   'HMAC-SHA256',
    signature:   'signature_from_package'
  },
  success(res) {
    // res.errMsg === 'openBusinessView:ok' 表示用户已点击同意
    // 业务成功仍以回调 / 查单为准
  },
  fail(err) {
    // 失败处理
  }
});
```

## 注意事项

| 项 | 要求 |
|----|------|
| `signature` 算法 | **APIv2 商户密钥 + HMAC-SHA256**（不要使用 APIv3 私钥）。错误使用 APIv3 私钥会返回 `SIGN_ERROR` |
| `package` 来源 | 必须取自后端 `CreateServiceOrder` 应答字段；前端禁止自行拼接 |
| `timestamp` 时效 | 5 分钟内有效，超时需重新下单 |
| 业务判定 | 前端 `success` 仅代表用户同意 UI，必须以「确认订单回调」或「查询支付分订单」为准 |
| `appid` 一致 | 调起的 appid 必须与 `CreateServiceOrder` 时入参 `appid` 相同 |
