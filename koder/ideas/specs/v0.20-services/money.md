# Money Service

## Summary

Decimal arithmetic for currency operations. Avoids floating-point errors that
plague JavaScript math. Store money as integer cents, format for display.

## The Problem

```javascript
// JavaScript floating point is broken for money
0.1 + 0.2
// 0.30000000000000004

19.99 * 3
// 59.97000000000001

// These tiny errors compound into real bugs:
// - Cart totals off by pennies
// - Invoice mismatches
// - Accounting discrepancies
```

## The Solution

Store money as integer cents, use proper decimal math:

```javascript
// Integer cents - no precision loss
1999 + 599
// 2598 (exactly $25.98)

// Format for display
fazt.services.money.format(2598, 'USD')
// "$25.98"
```

## Usage

### Basic Arithmetic

```javascript
// All operations work with integer cents
fazt.services.money.add(1999, 599)        // 2598
fazt.services.money.subtract(2598, 599)   // 1999
fazt.services.money.multiply(1999, 3)     // 5997
fazt.services.money.divide(6000, 3)       // 2000

// Chained operations
fazt.services.money.add(1999, 599, 299)   // 2897 (variadic)
```

### Percentages

```javascript
// Calculate 20% of $19.99
fazt.services.money.percent(1999, 20)     // 400 (rounds to nearest cent)

// Add 8.25% tax
fazt.services.money.addPercent(1999, 8.25) // 2164

// Subtract 15% discount
fazt.services.money.subtractPercent(1999, 15) // 1699
```

### Rounding

```javascript
// Explicit rounding control
fazt.services.money.divide(1000, 3)                    // 333 (default: round half up)
fazt.services.money.divide(1000, 3, { round: 'up' })   // 334
fazt.services.money.divide(1000, 3, { round: 'down' }) // 333
fazt.services.money.divide(1000, 3, { round: 'even' }) // 333 (banker's rounding)
```

### Formatting

```javascript
// Basic formatting
fazt.services.money.format(2598, 'USD')           // "$25.98"
fazt.services.money.format(2598, 'EUR')           // "€25.98"
fazt.services.money.format(2598, 'GBP')           // "£25.98"
fazt.services.money.format(259800, 'JPY')         // "¥2,598" (no decimals)

// Locale-aware formatting
fazt.services.money.format(2598, 'EUR', { locale: 'de-DE' }) // "25,98 €"
fazt.services.money.format(2598, 'EUR', { locale: 'fr-FR' }) // "25,98 €"
fazt.services.money.format(2598, 'USD', { locale: 'en-US' }) // "$25.98"

// Without symbol
fazt.services.money.format(2598, 'USD', { symbol: false })   // "25.98"

// With explicit sign
fazt.services.money.format(2598, 'USD', { sign: true })      // "+$25.98"
fazt.services.money.format(-2598, 'USD', { sign: true })     // "-$25.98"
```

### Parsing

```javascript
// Parse formatted strings back to cents
fazt.services.money.parse("$25.98", 'USD')        // 2598
fazt.services.money.parse("25,98 €", 'EUR')       // 2598
fazt.services.money.parse("25.98", 'USD')         // 2598 (no symbol ok)
fazt.services.money.parse("invalid")              // null
```

### Comparison

```javascript
fazt.services.money.compare(1999, 2598)   // -1 (less)
fazt.services.money.compare(2598, 1999)   // 1 (greater)
fazt.services.money.compare(1999, 1999)   // 0 (equal)

fazt.services.money.min(1999, 2598, 999)  // 999
fazt.services.money.max(1999, 2598, 999)  // 2598
```

### Allocation

Split money fairly (handles rounding):

```javascript
// Split $100 three ways
fazt.services.money.split(10000, 3)
// [3334, 3333, 3333] - first gets extra cent

// Proportional allocation
fazt.services.money.allocate(10000, [50, 30, 20])
// [5000, 3000, 2000]

// When it doesn't divide evenly
fazt.services.money.allocate(10001, [50, 50])
// [5001, 5000] - remainder goes to first
```

## Currency Configuration

```javascript
// Get currency info
fazt.services.money.currency('USD')
// { code: 'USD', symbol: '$', decimals: 2, name: 'US Dollar' }

fazt.services.money.currency('JPY')
// { code: 'JPY', symbol: '¥', decimals: 0, name: 'Japanese Yen' }

// List supported currencies
fazt.services.money.currencies()
// ['USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', ...]
```

## JS API

```javascript
// Arithmetic
fazt.services.money.add(...amounts)
fazt.services.money.subtract(a, b)
fazt.services.money.multiply(amount, factor)
fazt.services.money.divide(amount, divisor, options?)
// options: { round: 'up' | 'down' | 'even' | 'halfUp' }

// Percentages
fazt.services.money.percent(amount, percent)
fazt.services.money.addPercent(amount, percent)
fazt.services.money.subtractPercent(amount, percent)

// Formatting
fazt.services.money.format(cents, currency, options?)
// options: { locale, symbol, sign }
fazt.services.money.parse(string, currency)

// Comparison
fazt.services.money.compare(a, b)
fazt.services.money.min(...amounts)
fazt.services.money.max(...amounts)

// Allocation
fazt.services.money.split(amount, parts)
fazt.services.money.allocate(amount, ratios)

// Currency info
fazt.services.money.currency(code)
fazt.services.money.currencies()
```

## HTTP Endpoint

Not exposed via HTTP. Money operations are JS-side calculations.

## Storage Best Practices

```javascript
// Always store as integer cents
await fazt.storage.ds.insert('products', {
  name: 'T-Shirt',
  priceCents: 1999,      // $19.99 as integer
  currency: 'USD'
});

// Never store formatted strings or floats
// BAD: price: "$19.99"
// BAD: price: 19.99
```

## Go Library

Uses `shopspring/decimal` for arbitrary-precision arithmetic:

```go
import "github.com/shopspring/decimal"

func Add(amounts ...int64) int64 {
    result := decimal.Zero
    for _, a := range amounts {
        result = result.Add(decimal.NewFromInt(a))
    }
    return result.IntPart()
}
```

## Common Patterns

### Cart Total

```javascript
async function calculateCartTotal(cartId) {
  const items = await fazt.storage.ds.find('cart_items', { cartId });

  let subtotal = 0;
  for (const item of items) {
    const product = await fazt.storage.ds.findOne('products', { id: item.productId });
    subtotal = fazt.services.money.add(
      subtotal,
      fazt.services.money.multiply(product.priceCents, item.quantity)
    );
  }

  const tax = fazt.services.money.percent(subtotal, 8.25);
  const total = fazt.services.money.add(subtotal, tax);

  return {
    subtotal,
    tax,
    total,
    formatted: {
      subtotal: fazt.services.money.format(subtotal, 'USD'),
      tax: fazt.services.money.format(tax, 'USD'),
      total: fazt.services.money.format(total, 'USD')
    }
  };
}
```

### Invoice Line Items

```javascript
const lineItems = [
  { description: 'Widget', quantity: 3, unitPriceCents: 1999 },
  { description: 'Gadget', quantity: 1, unitPriceCents: 4999 }
];

const lines = lineItems.map(item => ({
  ...item,
  totalCents: fazt.services.money.multiply(item.unitPriceCents, item.quantity),
  formatted: fazt.services.money.format(
    fazt.services.money.multiply(item.unitPriceCents, item.quantity),
    'USD'
  )
}));
// [{ totalCents: 5997, formatted: "$59.97" }, { totalCents: 4999, formatted: "$49.99" }]
```

### Subscription Proration

```javascript
function prorateSubscription(monthlyPriceCents, daysRemaining, totalDays) {
  // Use allocation to avoid rounding errors
  const dailyRates = fazt.services.money.allocate(
    monthlyPriceCents,
    Array(totalDays).fill(1)
  );

  // Sum the days remaining
  return dailyRates.slice(0, daysRemaining).reduce(
    (sum, day) => fazt.services.money.add(sum, day),
    0
  );
}
```

## Supported Currencies

ISO 4217 currencies with correct decimal places:

| Currency | Code | Symbol | Decimals |
|----------|------|--------|----------|
| US Dollar | USD | $ | 2 |
| Euro | EUR | € | 2 |
| British Pound | GBP | £ | 2 |
| Japanese Yen | JPY | ¥ | 0 |
| Swiss Franc | CHF | CHF | 2 |
| Canadian Dollar | CAD | $ | 2 |
| Australian Dollar | AUD | $ | 2 |
| ... | ... | ... | ... |

Full list: ~150 currencies with correct decimal handling.

## Limits

| Limit | Default |
|-------|---------|
| `maxValue` | 9,223,372,036,854,775,807 (int64 max) |
| `minValue` | -9,223,372,036,854,775,808 (int64 min) |

## Implementation Notes

- ~30KB binary addition
- Pure Go (shopspring/decimal has no CGO)
- Integer overflow checked
- Currency data embedded (~10KB)
