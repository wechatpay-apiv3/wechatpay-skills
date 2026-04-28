# JSAPI / 小程序调起支付分订单详情页（商户）

> 源文档：[JSAPI调起支付分订单详情页](https://pay.weixin.qq.com/doc/v3/merchant/4012587983.md)
> 小程序入口：[wx.openBusinessView](https://pay.weixin.qq.com/doc/v3/merchant/4012587984.md)
> APP 端入口：[Android](https://pay.weixin.qq.com/doc/v3/merchant/4012587980.md) / [iOS](https://pay.weixin.qq.com/doc/v3/merchant/4012596423.md) / [鸿蒙](https://pay.weixin.qq.com/doc/v3/merchant/4015271812.md)

## 接入说明

商户在用户使用过程或订单完结后，希望让用户在微信内查看本笔支付分订单的费用 / 状态 / 明细时，可调起「订单详情页」。

订单详情页所需参数 `query_string` 由商户后端按以下格式拼接并签名（与「调起确认订单页」逻辑相同）：

```
mch_id={mch_id}&service_id={service_id}&out_order_no={out_order_no}&timestamp={timestamp}&nonce_str={nonce_str}&sign_type=HMAC-SHA256&signature={signature}
```

签名算法：使用 **APIv2 商户密钥 + HMAC-SHA256**（详见 [签名与验签规则.md](../../../接入指南/签名与验签规则.md) 中"客户端拉起签名（V2）"小节）。

## 公众号 H5 示例代码

```javascript
WeixinJSBridge.invoke(
  'openBusinessView',
  {
    businessType: 'wxpayScoreDetail',
    queryString:  '<query_string_assembled_by_backend>'
  },
  function (res) {
    if (res.err_msg === 'open_business_view:ok') {
      // 用户已访问详情页
    }
  }
);
```

## 小程序示例代码

```javascript
wx.openBusinessView({
  businessType: 'wxpayScoreDetail',
  extraData: {
    mch_id:      'xxx',
    service_id:  'xxx',
    out_order_no:'xxx',
    timestamp:   'xxx',
    nonce_str:   'xxx',
    sign_type:   'HMAC-SHA256',
    signature:   'xxx'
  },
  success(res) { /* ... */ },
  fail(err)    { /* ... */ }
});
```

## 注意事项

- `appid` 必须与「创建支付分订单」时入参一致。
- 详情页只读不可改：用户的修改 / 退款行为请走商户后台对应接口。
- 仅当订单已存在（`CREATED` / `DOING` / `DONE`）时可调起；未创建会报错。
