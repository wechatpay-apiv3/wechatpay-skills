# JSAPI 调起医保支付说明（商户）

> 来源：[JSAPI 调起医保自费混合支付](https://pay.weixin.qq.com/doc/v3/merchant/4016781549.md)

## 整体流程

1. 若有自费金额：调用 [JSAPI 自费下单](https://pay.weixin.qq.com/doc/v3/merchant/4012791882.md) 拿到自费 `prepay_id`，并按 [JSAPI 调起规则](https://pay.weixin.qq.com/doc/v3/merchant/4012791886.md) 计算 `timeStamp` / `nonceStr` / `package` / `signType` / `paySign`
2. 调用商户医保下单接口 `POST /v3/med-ins/orders` 拿到 `mix_trade_no`
3. 公众号 H5 页面通过 `WeixinJSBridge.invoke('requestMedicalInsurancePay', ...)` 调起医保支付
4. 用户输入医保电子凭证密码完成支付
5. H5 调用 `GET /v3/med-ins/orders/mix-trade-no/{mix_trade_no}` 查询最终结果，刷新业务页面
6. 服务端同时接收微信回调 `MEDICAL_INSURANCE.SUCCESS` 兜底

## 兼容性

- iOS / Android 微信版本 ≥ 8.0.44
- HarmonyOS 微信版本 ≥ 8.0.13

## 调用前准备

调用 `WeixinJSBridge.invoke('requestMedicalInsurancePay', ...)` 前必须先通过 [`wx.config`](https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html#4) 注入权限，并在 `jsApiList` 中包含 `requestMedicalInsurancePay`。

## 参数说明

| 参数 | 类型 | 必填 | 描述 |
| --- | --- | --- | --- |
| `appid` | string(32) | 是 | 商户公众号 AppID（必须与下单 `appid` 一致） |
| `mixTradeNo` | string(256) | 是 | 医保下单接口返回的 `mix_trade_no` |
| `timeStamp` | string(32) | 有自费时必填 | 时间戳，秒级 |
| `nonceStr` | string(32) | 有自费时必填 | 随机串 ≤32 位 |
| `package` | string(128) | 有自费时必填 | `prepay_id=...` |
| `signType` | string(32) | 有自费时必填 | 固定 `RSA` |
| `paySign` | string(256) | 有自费时必填 | 商户 API 证书私钥 RSA-SHA256 签名 |

## 调用示例

```html
<script src="https://res.wx.qq.com/open/js/jweixin-1.6.0.js"></script>
<script>
wx.config({
  debug: false,
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
      // res 示例：{ result: 'success', err_msg: 'requestMedicalInsurancePay:ok', msg: '已完成医保支付', err_desc: '' }
      if (res.err_msg === 'requestMedicalInsurancePay:ok') {
        // 调起成功（不代表支付一定成功），调用查单接口确认
      } else {
        // 调起失败
      }
    }
  );
});
</script>
```

## 回调结果

| 回调类型 | errMsg | 说明 |
| :-- | :-- | :-- |
| success | `requestMedicalInsurancePay:ok` | 调起流程结束 |
| fail | `requestMedicalInsurancePay:fail` | 调起流程失败 |

## 常见问题

- **`appid` 与下单不一致**：触发 `PARAM_ERROR`
- **未在 `jsApiList` 中声明**：触发 `the permission value is offline verifying`
- **`paySign` 算法错误**：必须用商户 API 证书私钥做 RSA-SHA256（不是 HMAC-SHA256，也不是 APIv3 密钥）
- **签名串构造错误**：必须按 `appId\ntimeStamp\nnonceStr\npackage\n` 顺序拼接（每行末尾换行）
