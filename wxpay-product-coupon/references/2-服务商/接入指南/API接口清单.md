# 服务商 — API 接口

请求域名：
- 主域名: `https://api.mch.weixin.qq.com`
- 备域名: `https://api2.mch.weixin.qq.com`

## 一、API 路径

服务商商品券 API 路径以 `/v3/marketing/partner/product-coupon/` 开头，门店/名片 API 以 `/v3/brand/partner/` 开头。

| 操作 | 方法与路径 |
|------|-----------|
| 品牌入驻 | `POST /v3/brand/applyments` |
| 创建品牌门店 | `POST /v3/brand/partner/store/brandstores` |
| 创建商家名片 | `POST /v3/brand/partner/cardtemplate/brandcards` |
| 创建交易连接名片 | `POST /v3/brand/partner/cardtemplate/traderelation` |
| 设置回调地址 | `POST /v3/marketing/product-coupon/notify-config` |
| 创建商品券（单券） | `POST /v3/marketing/partner/product-coupon/product-coupons` |
| 创建商品券（多次优惠） | `POST /v3/marketing/partner/product-coupon/product-coupons` |
| 修改批次（单券） | `PATCH /v3/marketing/partner/product-coupon/product-coupons/{product_coupon_id}/stocks/{stock_id}` |
| 修改批次组（多次优惠） | `PATCH /v3/marketing/partner/product-coupon/product-coupons/{product_coupon_id}/stock-bundles/{bundle_id}` |
| 创建投放计划（摇优惠） | `POST /v3/marketing/partner/delivery-plan/delivery-plans` |
| 预发放商品券 | `POST /v3/marketing/partner/product-coupon/users/{openid}/pre-send-coupon` |
| 确认发放 | `POST /v3/marketing/partner/product-coupon/users/{openid}/coupons/{coupon_code}/confirm` |
| 核销商品券 | `POST /v3/marketing/partner/product-coupon/users/{openid}/coupons/{coupon_code}/use` |

> 路径中的 `{product_coupon_id}` 为创建商品券时返回的商品券 ID。

### 服务商独有 API

以下操作在品牌直连中是 UI 操作，服务商模式**全部有 API**：

| 操作 | API | 说明 |
|------|-----|------|
| 品牌入驻 | `POST /v3/brand/applyments` | 帮品牌方提交入驻申请 |
| 创建商家名片 | `POST /v3/brand/partner/cardtemplate/brandcards` | 服务商独有 |
| 创建交易连接名片 | `POST /v3/brand/partner/cardtemplate/traderelation` | 关联AppID的关键步骤 |
| 创建投放计划 | `POST /v3/marketing/partner/delivery-plan/delivery-plans` | 摇优惠投放 |

## 二、支持的券类型

| # | 券类型 |
|---|--------|
| 1 | 全场折扣-单券 |
| 2 | 全场满减-单券 |
| 3 | 单品折扣-单券 |
| 4 | 单品满减-单券 |
| 5 | 单品兑换-单券 |
| 6 | 全场折扣-多次优惠 |
| 7 | 全场满减-多次优惠 |
| 8 | 单品折扣-多次优惠 |
| 9 | 单品满减-多次优惠 |
| 10 | 单品兑换-多次优惠 |

## 三、创建商品券关键参数约束

> 完整参数说明见官方文档：[服务商-创建商品券接口文档](https://pay.weixin.qq.com/doc/v3/partner/4015781289)

### 顶层参数

| 参数 | 必填 | 类型 | 约束 |
|------|------|------|------|
| out_request_no | 是 | string(40) | 6-40个字符，品牌侧唯一，支持数字、大小写字母、`_`、`-` |
| scope | 是 | string | `ALL`（全场，仅 NORMAL/DISCOUNT）或 `SINGLE`（单品，支持 NORMAL/DISCOUNT/EXCHANGE） |
| type | 是 | string | `NORMAL`（满减）、`DISCOUNT`（折扣）、`EXCHANGE`（兑换，仅 scope=SINGLE） |
| usage_mode | 是 | string | `SINGLE`（单券） |
| out_product_no | 否 | string(40) | 品牌侧商品标识，不校验唯一性 |
| brand_id | 是 | string | **服务商必填**，微信支付为品牌方分配的唯一标识，需与服务商存在授权关系 |

### display_info（展示信息）

| 参数 | 必填 | 类型 | 约束 |
|------|------|------|------|
| name | 是 | string(15) | **3-15 个 UTF-8 字符** |
| image_url | 是 | string | 仅支持图片上传API获取的URL，宽高比1:1，建议1080×1080，≤2M |
| background_url | 是 | string | 仅支持图片上传API获取的URL，宽高比65:77，建议1170×1326，≤2M |
| detail_image_url_list | 否 | array[string] | 最多 **8** 张 |
| original_price | scope=SINGLE 时必填 | integer | 单位：分 |
| combo_package_list | scope=SINGLE 时必填 | array[object] | 最多 **50** 个组合 |

### combo_package_list 内部

| 参数 | 必填 | 类型 | 约束 |
|------|------|------|------|
| name | 是 | string(15) | ≤15 个 UTF-8 字符 |
| pick_count | 是 | integer | 用户可选单品数量 |
| choice_list[].name | 是 | string(15) | ≤15 个 UTF-8 字符 |
| choice_list[].price | 是 | integer | 单位：分 |
| choice_list[].count | 是 | integer | 最多 99 |
| choice_list[].mini_program_appid | 否 | string | 单品跳转小程序AppID，需与品牌方或服务商存在绑定关系 |
| choice_list[].mini_program_path | 否 | string | 单品跳转小程序路径，需和 mini_program_appid 一同填写 |
| choice_list | 是 | array[object] | 最多 **99** 个单品 |

### single_usage_info（全场券优惠规则，scope=ALL 时必填）

| 条件 | 参数 | 必填字段 |
|------|------|---------|
| type=NORMAL | normal_coupon | threshold（分，0=无门槛）、discount_amount（分） |
| type=DISCOUNT | discount_coupon | threshold（分，0=无门槛）、percent_off（如 30=减30%即打7折） |

### stock（批次信息）

| 参数 | 必填 | 约束 |
|------|------|------|
| remark | 否 | ≤20 个 UTF-8 字符，仅品牌可见 |
| coupon_code_mode | 是 | `WECHATPAY` / `UPLOAD` / `API_ASSIGN` |
| stock_send_rule.max_count | 是 | ≤100,000,000 |
| stock_send_rule.max_count_per_day | 否 | ≤100,000,000，默认不限 |
| stock_send_rule.max_count_per_user | 是 | ≤100 |
| store_scope | 是 | `NONE` / `ALL` / `SPECIFIC` |

### stock.single_usage_rule（单券使用规则）

| 参数 | 必填 | 约束 |
|------|------|------|
| coupon_available_period.available_begin_time | 是 | RFC3339 格式，批次有效期最长 365 天，摇优惠场景建议 ≥5 天 |
| coupon_available_period.available_end_time | 是 | RFC3339 格式 |
| coupon_available_period.available_days | 否 | 券生效后 N 天内可用，最多 365 天，不可与 irregular_available_period_list 同时配置 |
| coupon_available_period.wait_days_after_receive | 否 | 领券后等待 N 天生效，最多 30 天，需配合 available_days，不可单独填写 |
| coupon_available_period.weekly_available_period | 否 | 每周固定可用时间，不可与 irregular_available_period_list 同时配置 |
| coupon_available_period.irregular_available_period_list | 否 | 指定可用时间段，最多 2 个区间，不可与 available_days/wait_days_after_receive/weekly_available_period 同时配置 |
| normal_coupon | scope=SINGLE 且 type=NORMAL 时必填 | threshold + discount_amount |
| discount_coupon | scope=SINGLE 且 type=DISCOUNT 时必填 | threshold + percent_off |
| exchange_coupon | scope=SINGLE 且 type=EXCHANGE 时必填 | threshold + exchange_price |

> **注意**：全场券（scope=ALL）的优惠规则放在顶层 `single_usage_info` 中，单品券（scope=SINGLE）的优惠规则放在 `stock.single_usage_rule` 中。

### stock.usage_rule_display_info

| 参数 | 必填 | 约束 |
|------|------|------|
| coupon_usage_method_list | 是 | `OFFLINE` / `MINI_PROGRAM` / `APP` / `PAYMENT_CODE`，可多选 |
| mini_program_appid | 含 MINI_PROGRAM 时必填 | 品牌方或服务商小程序AppID，需和 mini_program_path 一同填写 |
| app_path | 含 APP 时必填 | 品牌方或服务商APP跳转路径 |
| usage_description | 是 | ≤1000 个 UTF-8 字符 |
| coupon_available_store_info.description | 是 | ≤1000 个 UTF-8 字符 |

### stock.coupon_display_info

| 参数 | 必填 | 约束 |
|------|------|------|
| code_display_mode | 是 | `INVISIBLE` / `BARCODE` / `QRCODE` |
| background_color | 否 | 默认 Color020，可选 10 种颜色 |
| entrance_mini_program.entrance_wording | 是 | **≤5 个 UTF-8 字符** |
| entrance_mini_program.guidance_wording | 是 | **≤6 个 UTF-8 字符** |

### stock.notify_config

| 参数 | 必填 | 约束 |
|------|------|------|
| notify_appid | 是 | 服务商或品牌的AppID，支持小程序/服务号/公众号/APP 类型，需与服务商或品牌存在绑定关系 |

### 常见错误码

| 状态码 | 错误码 | 描述 | 解决方案 |
|--------|--------|------|---------|
| 400 | PARAM_ERROR | 参数错误 | 根据错误提示修正参数 |
| 400 | INVALID_REQUEST | 请求不符合业务规则 | 参考文档中字段要求及组合要求 |
| 400 | ALREADY_EXISTS | 商品券已存在 | 确认 out_request_no 是否重复 |
| 401 | SIGN_ERROR | 验证不通过 | 检查签名方案是否使用 WECHATPAY2-SHA256-RSA2048 |
| 429 | RATELIMIT_EXCEEDED | 超过频率限制（20/s） | 稍后使用原参数重试 |
| 500 | SYSTEM_ERROR | 系统异常 | 使用相同 out_request_no 重试 |

## 四、请求体差异（vs 品牌直连）

服务商模式**必须在请求体中传 `brand_id`** 字段，指定品牌方。
