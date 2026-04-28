# 医保混合收款成功通知说明（商户 - Go）

> 内容与 [`Java/6-回调通知/医保混合收款成功通知说明.md`](../../Java/6-回调通知/医保混合收款成功通知说明.md) 完全一致；本副本仅为 Go 项目按目录约定查找方便而存在。

> 来源：[医保混合收款成功通知](https://pay.weixin.qq.com/doc/v3/merchant/4016781554.md)
> 通用解密 / 验签 / 回包流程参考 [📄 ../../../接入指南/回调处理.md](../../../接入指南/回调处理.md)

## 一、回调时机

当订单 `mix_pay_status = MIX_PAY_SUCCESS`（自费 + 医保两端均结算成功）时，微信支付通过 POST 向 [医保下单接口](https://pay.weixin.qq.com/doc/v3/merchant/4016781466.md) 中传入的 `callback_url` 发送通知。

| 事件类型 | 含义 |
| --- | --- |
| `MEDICAL_INSURANCE.SUCCESS` | 医保混合收款成功 |

> ‼️ 微信仅在订单达到 `MIX_PAY_SUCCESS` 时回调一次。若 5 秒内未收到 200/204 应答，将按指数退避重试，30 秒后**不再重试**。
> ‼️ 商户**必须**对 `MIX_PAY_CREATED` 状态的订单做主动查询兜底（`GET /v3/med-ins/orders/mix-trade-no/{mix_trade_no}`），不能仅依赖回调。

## 二、HTTP 头

| 参数 | 描述 |
| --- | --- |
| `Wechatpay-Serial` | 验签所用微信支付公钥 ID（`PUB_KEY_ID_*`）或微信支付平台证书序列号 |
| `Wechatpay-Signature` | 签名值 |
| `Wechatpay-Timestamp` | 时间戳（秒） |
| `Wechatpay-Nonce` | 随机串 |

> ‼️ `Wechatpay-Serial` 以 `PUB_KEY_ID_` 开头则用**微信支付公钥**验签，否则用**微信支付平台证书**验签。
> ‼️ 微信会随机下发 `Wechatpay-Signature` 以 `WECHATPAY/SIGNTEST/` 开头的[签名探测流量](https://pay.weixin.qq.com/doc/v3/merchant/4013053249.md)，验签必失败，商户必须按规范应答 4XX/5XX。

## 三、报文结构

```json
{
  "id": "EV-2018022511223320873",
  "create_time": "2020-03-26T10:43:39+08:00",
  "event_type": "MEDICAL_INSURANCE.SUCCESS",
  "resource_type": "encrypt-resource",
  "resource": {
    "algorithm": "AEAD_AES_256_GCM",
    "ciphertext": "...",
    "nonce": "...",
    "associated_data": "..."
  }
}
```

## 四、resource.ciphertext 解密后的业务字段

> 算法：`AEAD_AES_256_GCM`，密钥：APIv3 密钥（32 字节）
> `nonce` / `associated_data` 用密文中对应字段，**不是**自己生成

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `mix_trade_no` | string(32) | 医保自费混合订单号 |
| `mix_pay_status` | string | 整体状态（回调中固定 `MIX_PAY_SUCCESS`） |
| `self_pay_status` | string | `SELF_PAY_SUCCESS` / `NO_SELF_PAY` |
| `med_ins_pay_status` | string | `MED_INS_PAY_SUCCESS` / `NO_MED_INS_PAY` |
| `paid_time` | string(64) | 支付完成时间，RFC3339 |
| `passthrough_response_content` | string(2048) | 医保局透传给医疗机构的内容 |
| `mix_pay_type` | string | `CASH_ONLY` / `INSURANCE_ONLY` / `CASH_AND_INSURANCE` |
| `order_type` | string | `REG_PAY` / `DIAG_PAY` / ... 详见 SKILL 总览 |
| `appid` | string(32) | 商户公众号 / 小程序 AppID |
| `openid` | string(128) | 用户在该 AppID 下的 openid |
| `pay_for_relatives` | bool | 是否代亲属支付 |
| `out_trade_no` | string(64) | 商户订单号 |
| `serial_no` | string(20) | 医疗机构订单号 |
| `pay_order_id` | string(64) | 医保局支付单 ID |
| `pay_auth_no` | string(40) | 医保局支付授权码 |
| `geo_location` | string(40) | 用户经纬度 `经度,纬度` |
| `city_id` | string(8) | 城市 ID |
| `med_inst_name` | string(128) | 医疗机构名称 |
| `med_inst_no` | string(32) | 医疗机构编码 |
| `med_ins_order_create_time` | string(64) | 医保下单时间 |
| `total_fee` | uint64 | 订单总金额（分） |
| `med_ins_gov_fee` | uint64 | 医保统筹支付金额（分） |
| `med_ins_self_fee` | uint64 | 医保个账支付金额（分） |
| `med_ins_other_fee` | uint64 | 医保其他津贴金额（分） |
| `med_ins_cash_fee` | uint64 | 医保结算后自费金额（分） |
| `wechat_pay_cash_fee` | uint64 | 微信支付实收金额（分） |
| `cash_add_detail[].cash_add_fee` | uint64 | 现金补充金额 |
| `cash_add_detail[].cash_add_type` | string | `DEFAULT_ADD_TYPE` / `FREIGHT` / `OTHER_MEDICAL_EXPENSES` |
| `cash_reduce_detail[].cash_reduce_fee` | uint64 | 现金减免金额 |
| `cash_reduce_detail[].cash_reduce_type` | string | `DEFAULT_REDUCE_TYPE` / `HOSPITAL_REDUCE` / `PHARMACY_DISCOUNT` / `DISCOUNT` / `PRE_PAYMENT` / `DEPOSIT_DEDUCTION` |
| `callback_url` | string(256) | 回调通知 URL |
| `prepay_id` | string(64) | 自费预下单 ID |
| `attach` | string(128) | 商户自定义透传 |
| `channel_no` | string(32) | 渠道号 |
| `med_ins_test_env` | bool | 是否医保局测试环境 |

## 五、应答规范

| 场景 | HTTP 状态码 | 应答体 |
| --- | --- | --- |
| 验签通过 + 业务处理成功 | 200 / 204 | 无包体 |
| 验签失败 / 业务处理失败 | 4XX / 5XX | `{"code": "FAIL", "message": "失败"}` |

> ‼️ 必须先验签再处理业务，验签失败时**禁止**返回 200，否则微信将认为商户已收，不再重试，造成丢单。
> ‼️ 业务处理建议异步化，回调线程内只做：验签 → 解密 → 写入「待处理订单表」→ 立即应答 200。后续状态更新走异步消费，避免业务慢导致回调超时被重试。

## 六、幂等性

同一订单可能收到多次回调（重试或网络抖动），商户必须按 `mix_trade_no` 做幂等：

```
SELECT mix_pay_status FROM med_ins_orders WHERE mix_trade_no = ?;
IF mix_pay_status = 'MIX_PAY_SUCCESS' THEN
  -- 已处理，直接返回 200
ELSE
  -- 用 SELECT FOR UPDATE / 行锁更新状态，再返回 200
END IF;
```

## 七、与查单的关系

| 场景 | 推荐做法 |
| --- | --- |
| 30 秒内收到回调 | 验签 + 解密 + 幂等更新订单 |
| 30 秒后未收到回调 | 主动调用查单接口确认状态 |
| 收到回调但解密 / 验签失败 | 应答 4XX/5XX，触发微信重试；同时 LOG 报警，调用查单接口兜底 |
| 用户客诉「钱已扣未发货」 | 优先以查单结果为准，不依赖回调记录 |
