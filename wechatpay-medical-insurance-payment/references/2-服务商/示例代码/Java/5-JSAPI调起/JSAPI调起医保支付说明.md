# JSAPI 调起医保支付说明（服务商）

> 来源：[JSAPI 调起医保自费混合支付](https://pay.weixin.qq.com/doc/v3/merchant/4016781549.md)（接口同商户）

## 整体流程

1. 若有自费金额：调用 [服务商 JSAPI 自费下单](https://pay.weixin.qq.com/doc/v3/partner/4012692411.md) 拿到自费 `prepay_id`，按 [JSAPI 调起规则](https://pay.weixin.qq.com/doc/v3/partner/4012692421.md) 计算调起参数
2. 调用服务商医保下单接口 `POST /v3/med-ins/orders`（带 `sub_mchid` / `sub_appid`）拿到 `mix_trade_no`
3. 公众号 H5 页面调用 `WeixinJSBridge.invoke('requestMedicalInsurancePay', ...)`
4. 用户输入医保电子凭证密码完成支付
5. H5 调用查单接口（带 `sub_mchid`）确认结果

## 兼容性

- iOS / Android 微信版本 ≥ 8.0.44
- HarmonyOS 微信版本 ≥ 8.0.13

## 调用前准备

通过 [`wx.config`](https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html#4) 注入权限，`jsApiList` 必须包含 `requestMedicalInsurancePay`。注入用的 `appId` 与 `signature` **必须**与即将调起所用 `appid` 一致（服务商 H5 用服务商 AppID，子商户 H5 用子商户 AppID）。

## 服务商场景的 AppID 选择

| 下单时传 | 调起 `appid` 取值 |
| --- | --- |
| `openid` | 服务商 AppID |
| `sub_openid` | 子商户 AppID（`sub_appid`） |

## 调用参数

| 参数 | 类型 | 必填 | 描述 |
| --- | --- | --- | --- |
| `appid` | string(32) | 是 | 与下单一致：传 `openid` 时填服务商 AppID，传 `sub_openid` 时填 `sub_appid` |
| `mixTradeNo` | string(256) | 是 | 服务商医保下单接口返回的 `mix_trade_no` |
| `timeStamp` | string(32) | 有自费时必填 | 时间戳，秒级 |
| `nonceStr` | string(32) | 有自费时必填 | 随机串 ≤32 位 |
| `package` | string(128) | 有自费时必填 | `prepay_id=...` |
| `signType` | string(32) | 有自费时必填 | 固定 `RSA` |
| `paySign` | string(256) | 有自费时必填 | 服务商 API 证书私钥 RSA-SHA256 签名 |

## 调用示例

```html
<script src="https://res.wx.qq.com/open/js/jweixin-1.6.0.js"></script>
<script>
wx.config({
  appId: 'wx8888888888888888',
  timestamp: 1414561699,
  nonceStr: 'XXXXXXXX',
  signature: 'XXXXXXXX',
  jsApiList: ['requestMedicalInsurancePay']
});

wx.ready(function () {
  WeixinJSBridge.invoke(
    'requestMedicalInsurancePay',
    {
      appid: 'wx8888888888888888',
      mixTradeNo: '1217752501201407033233368318',
      timeStamp: '1414561699',
      nonceStr: '5K8264ILTKCH16CQ2502SI8ZNMTM67VS',
      package: 'prepay_id=wx201410272009395522657a690389285100',
      signType: 'RSA',
      paySign: 'oR9d8Puhn...'
    },
    function (res) {
      if (res.err_msg === 'requestMedicalInsurancePay:ok') {
        // 调起结束，调用查单接口（带 sub_mchid）确认
      } else {
        // 调起失败
      }
    }
  );
});
</script>
```

## 常见问题

- **`appid` 与 `signature` 不一致**：`wx.config` 注入的 `appId` 必须与调起 `appid` 一致
- **签名错误**：用**服务商**API 证书私钥；子商户无独立证书
- **回调路由错乱**：服务端回调中以 `sub_mchid` 路由到对应业务系统
