# checkout-sdk-go

[CHECKOUT.UZ](https://checkout.uz) merchant API'si uchun rasmiy Go kutubxonasi — bitta paket orqali **v1** va **v2** versiyalarining ikkalasiga ham kirish imkoni.

```go
client, _ := checkoutuz.NewCheckoutClient("kassangizning_api_kaliti")
payment, err := client.V2.CreatePayment(ctx, 50000, checkoutuz.CreatePaymentParamsV2{
    Description: "Buyurtma #1",
})
fmt.Println(payment["payment"].(map[string]interface{})["_url"]) // shu havolani mijozga yuboring
```

## Talablar

- **Go 1.20+** (`net/http`'dan foydalanadi, tashqi HTTP kutubxonasi shart emas)

## O'rnatish

```bash
go get github.com/checkout-uz/checkout-sdk-go
```

## API kalitini qayerdan olish mumkin

1. [checkout.uz](https://checkout.uz)ga kiring → **Kassalar** → kerakli kassani oching → **Sozlamalar**.
2. U yerda ko'rsatilgan **API KEY**ni nusxalang.
3. Bu kalitni hech kimga bermang — faqat server tomonida (backend) ishlating.

## Tezkor boshlash

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    "github.com/checkout-uz/checkout-sdk-go"
)

func main() {
    client, err := checkoutuz.NewCheckoutClient("kassangizning_api_kaliti")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    payment, err := client.V2.CreatePayment(ctx, 50000, checkoutuz.CreatePaymentParamsV2{
        Description: "Buyurtma #1",
        ExpireTime:  3600, // ixtiyoriy, soniyada, standart 1 soat
    })
    if err != nil {
        var apiErr *checkoutuz.APIError
        if errors.As(err, &apiErr) {
            log.Fatalf("Xatolik (%d): %s", apiErr.StatusCode, apiErr.Message)
        }
        log.Fatal(err)
    }

    p := payment["payment"].(map[string]interface{})
    fmt.Println(p["_url"])  // mijozni shu to'lov sahifasiga yo'naltiring
    fmt.Println(p["_uuid"]) // buni saqlab qo'ying - to'lovning doimiy ID'si
}
```

Mijoz to'lov qilgandan keyin, checkout.uz sizning `WebhookURL`'ingizga xabar yuboradi — yoki `StatusPaymentByUUID` orqali o'zingiz tekshirishingiz mumkin:

```go
status, _ := client.V2.StatusPaymentByUUID(ctx, p["_uuid"].(string))
fmt.Println(status["data"].(map[string]interface{})["status"]) // "pending" / "paid" / "canceled" / "reversed"
```

## v1 va v2 farqi

```go
client.V1.CreatePayment(ctx, 50000, checkoutuz.CreatePaymentParams{})
client.V2.CreatePayment(ctx, 50000, checkoutuz.CreatePaymentParamsV2{ExpireTime: 7200, IdempotencyKey: "order-42"})
```

v2 v1'ning hamma imkoniyatlari ustiga quyidagilarni qo'shadi:

- `IdempotencyKey` (`CreatePayment`da) — xavfsiz qayta urinish
- `ExpireTime` — invoys qancha vaqt ochiq turishini belgilash
- `GetHistory(ctx, GetHistoryParams{...})` — sahifalash va filtrlash
- `GetStats()` — endi `by_provider` taqsimot bilan
- `CancelPayment(ctx, orderID)` — hali "kutilmoqda" holatidagi invoysni bekor qilish
- `RefundPayment(ctx, orderID)` — muvaffaqiyatli to'lovni qaytarish
- `GetFiscal(ctx, orderID)` — fiskal chekni olish

## Ixtiyoriy parametrlar haqida

Go'da named/optional parametrlar yo'qligi sababli, ixtiyoriy argumentlar har bir metod uchun maxsus `Params` struct orqali uzatiladi (masalan `checkoutuz.CreatePaymentParamsV2{Description: "...", ExpireTime: 7200}`). Bo'sh qoldirilgan maydonlar (`""`, `0`) avtomatik tushirib qoldiriladi.

## Hisob darajasidagi metodlar (`ProfileClient`)

```go
profile, _ := checkoutuz.NewProfileClient("profil_api_kalitingiz")
shops, _ := profile.GetShops(ctx)
stats, _ := profile.GetProfileStats(ctx, "2026-01-01", "")
```

## Xatoliklarni qayta ishlash

Har bir metod Go'ning standart `error` naqshiga mos — muvaffaqiyatsizlikda `*checkoutuz.APIError` qaytaradi:

```go
_, err := client.V2.RefundPayment(ctx, 123)
var apiErr *checkoutuz.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Message)
}
```

## To'liq API hujjatlari

- v1: https://checkout.uz/api-docs/v1
- v2: https://checkout.uz/api-docs/v2

## Ishlab chiqish (development)

```bash
go test ./...
```

Barcha testlar HTTP transportini soxtalashtiradi (`fakeTransport`) — hech biri haqiqiy checkout.uz API'siga murojaat qilmaydi.

## Litsenziya

MIT
