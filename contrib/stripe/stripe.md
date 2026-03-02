# Stripe API Client for Duso

Complete REST API wrapper for Stripe payments. Supports customers, payment intents, subscriptions, invoices, refunds, products, prices, and more.

## Installation

The Stripe module is built into Duso. Use `require()` to import it:

```duso
stripe = require("stripe")
```

## Quick Start

```duso
stripe = require("stripe")
client = stripe.create_client("sk_test_...")  // or set STRIPE_API_KEY env var

// Create a customer
customer = client.customers.create("customer@example.com")
print(customer.id)

// Create a payment intent
intent = client.payment_intents.create(
  5000,        // amount in cents ($50.00)
  "usd",       // currency
  customer.id  // optional customer
)
print(intent.client_secret)
```

## Authentication

Set your Stripe API key via environment variable:

```bash
export STRIPE_API_KEY="sk_test_..."
duso script.du
```

Or pass it directly to `create_client()`:

```duso
client = stripe.create_client("sk_test_...")
```

> **Note:** Always use secret API keys (sk_*) in server-side code. Never use publishable keys (pk_*) for backend operations.

## API Reference

### Customers

**Create a customer:**
```duso
customer = client.customers.create(
  "customer@example.com",
  "Optional description",
  {custom_key = "value"}  // metadata
)
```

**Get a customer:**
```duso
customer = client.customers.get("cus_XXXXX")
```

**Update a customer:**
```duso
customer = client.customers.update(
  "cus_XXXXX",
  "newemail@example.com"
)
```

**Delete a customer:**
```duso
client.customers.delete("cus_XXXXX")
```

**List customers:**
```duso
customers = client.customers.list(10)  // limit=10
```

### Payment Intents

**Create a payment intent:**
```duso
intent = client.payment_intents.create(
  5000,        // amount in cents
  "usd",       // currency
  "cus_XXXXX", // customer_id (optional)
  "Order #123" // description (optional)
)
```

**Get a payment intent:**
```duso
intent = client.payment_intents.get("pi_XXXXX")
```

**Confirm a payment intent:**
```duso
intent = client.payment_intents.confirm(
  "pi_XXXXX",
  "pm_XXXXX"  // payment_method
)
```

**List payment intents:**
```duso
intents = client.payment_intents.list("cus_XXXXX", 10)
```

### Charges (Legacy)

> **Note:** Payment Intents are the modern approach. Use Charges for legacy integrations.

**Create a charge:**
```duso
charge = client.charges.create(
  5000,    // amount in cents
  "usd",   // currency
  "tok_XXXXX"  // source token
)
```

**Get a charge:**
```duso
charge = client.charges.get("ch_XXXXX")
```

**List charges:**
```duso
charges = client.charges.list("cus_XXXXX", 10)
```

**Refund a charge:**
```duso
refund = client.charges.refund("ch_XXXXX", 2500)  // partial refund
```

### Refunds

**Create a refund:**
```duso
refund = client.refunds.create(
  "ch_XXXXX",
  2500,      // amount (optional, full refund if omitted)
  "requested_by_customer"  // reason
)
```

**Get a refund:**
```duso
refund = client.refunds.get("re_XXXXX")
```

**List refunds:**
```duso
refunds = client.refunds.list(10, "ch_XXXXX")
```

### Subscriptions

**Create a subscription:**
```duso
subscription = client.subscriptions.create(
  "cus_XXXXX",
  [{price = "price_XXXXX", quantity = 1}],  // items array
  14  // trial_period_days (optional)
)
```

**Get a subscription:**
```duso
subscription = client.subscriptions.get("sub_XXXXX")
```

**Update a subscription:**
```duso
subscription = client.subscriptions.update(
  "sub_XXXXX",
  [{price = "price_XXXXX", quantity = 2}]
)
```

**Cancel a subscription:**
```duso
subscription = client.subscriptions.cancel("sub_XXXXX")
```

**List subscriptions:**
```duso
subscriptions = client.subscriptions.list("cus_XXXXX", 10, "active")
```

### Invoices

**Get an invoice:**
```duso
invoice = client.invoices.get("in_XXXXX")
```

**List invoices:**
```duso
invoices = client.invoices.list("cus_XXXXX", 10, "open")
```

**Finalize an invoice:**
```duso
invoice = client.invoices.finalize("in_XXXXX")
```

**Pay an invoice:**
```duso
invoice = client.invoices.pay("in_XXXXX")
```

### Products

**Create a product:**
```duso
product = client.products.create(
  "Premium Plan",
  "Monthly subscription"
)
```

**Get a product:**
```duso
product = client.products.get("prod_XXXXX")
```

**List products:**
```duso
products = client.products.list(10)
```

### Prices

**Create a price:**
```duso
price = client.prices.create(
  "prod_XXXXX",
  2999,  // $29.99 in cents
  "usd",
  {interval = "month", interval_count = 1}  // recurring config
)
```

**Get a price:**
```duso
price = client.prices.get("price_XXXXX")
```

**List prices:**
```duso
prices = client.prices.list("prod_XXXXX", 10)
```

### Payment Methods

**Get a payment method:**
```duso
pm = client.payment_methods.get("pm_XXXXX")
```

**List payment methods:**
```duso
pms = client.payment_methods.list("cus_XXXXX", 10)
```

## Error Handling

API errors throw exceptions with descriptive messages:

```duso
try
  customer = client.customers.get("invalid_id")
catch (error)
  print("Error: " + error)
end
```

## Examples

### Example 1: Create a customer and charge them

```duso
stripe = require("stripe")
client = stripe.create_client()

// Create customer
customer = client.customers.create("alice@example.com", "Alice Smith")
print("Created customer: " + customer.id)

// Create a charge
charge = client.charges.create(
  5000,      // $50.00
  "usd",
  "tok_visa",
  "Order #001",
  customer.id
)
print("Charge created: " + charge.id + " - Status: " + charge.status)
```

### Example 2: Create a subscription

```duso
stripe = require("stripe")
client = stripe.create_client()

// Create product and price
product = client.products.create("Premium Plan")
price = client.prices.create(
  product.id,
  9999,  // $99.99/month
  "usd",
  {interval = "month"}
)

// Create customer
customer = client.customers.create("subscriber@example.com")

// Create subscription
subscription = client.subscriptions.create(
  customer.id,
  [{price = price.id, quantity = 1}],
  30  // 30-day trial
)

print("Subscription created: " + subscription.id)
print("Trial ends: " + subscription.trial_end)
```

### Example 3: List and refund a charge

```duso
stripe = require("stripe")
client = stripe.create_client()

// List recent charges
charges = client.charges.list("cus_XXXXX", 1)

if len(charges) > 0 then
  charge = charges[0]
  print("Latest charge: " + charge.id + " - " + charge.amount + " " + charge.currency)

  // Refund half
  refund = client.charges.refund(charge.id, charge.amount / 2)
  print("Refunded: " + refund.id)
end
```

### Example 4: List all active subscriptions

```duso
stripe = require("stripe")
client = stripe.create_client()

subscriptions = client.subscriptions.list(nil, 100, "active")

for sub in subscriptions do
  print("Subscription: " + sub.id)
  print("  Customer: " + sub.customer)
  print("  Status: " + sub.status)
  print("  Current period ends: " + sub.current_period_end)
end
```

## API Response Format

All API calls return Stripe objects as-is. Common fields:

- `id` - Object ID (cus_*, ch_*, pi_*, etc.)
- `object` - Type of object (customer, charge, payment_intent, etc.)
- `created` - Unix timestamp of creation
- `amount` - Amount in cents (for charges, intents, etc.)
- `currency` - 3-letter currency code (usd, eur, etc.)
- `status` - Current status (succeeded, pending, failed, etc.)
- `metadata` - Custom key-value data

See [Stripe API documentation](https://stripe.com/docs/api) for complete object schemas.

## Limitations

- Base64 encoding is simplified. For production use, consider adding proper base64 support to Duso.
- Form encoding uses basic URL encoding. Complex nested structures may need additional handling.
- Stripe webhooks are not yet supported (would require HTTP server functionality).

## Testing

Test with Stripe's test API key and tokens:

```bash
export STRIPE_API_KEY="sk_test..."
```

Use test card numbers:
- `4242 4242 4242 4242` - Visa
- `5555 5555 5555 4444` - Mastercard
- `3782 822463 10005` - American Express

See [Stripe Testing Documentation](https://stripe.com/docs/testing) for more test cards.
