# Stripe API Client for Duso

Complete REST API wrapper for Stripe payments. Supports customers, payment intents, subscriptions, checkout sessions, invoices, refunds, products, prices, and webhook verification.

## Installation

The Stripe module is built into Duso. Use `require()` to import it:

```duso
stripe = require("stripe")
```

To vendor a copy — to patch it, or to pin it against a contrib update —
extract it and require it **by path**:

```bash
duso extract contrib/stripe .
```

```duso
stripe = require("./stripe/stripe.du")
```

The leading `./` is not optional. A bare `require("stripe")` resolves to
the embedded module even when a local copy sits right beside the script,
and it does so silently — the vendored file is simply never read. Paths
resolve against the directory of the entry script, not the working
directory or the file doing the requiring.

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

**Update a customer** — takes an options object, and passes through any
field Stripe accepts:

```duso
customer = client.customers.update("cus_XXXXX", {
  email = "new@example.com",
  name = "Acme Inc",
  phone = "+1 555 0100",
  address = {line1 = "1 Main St", city = "Springfield", country = "US"}
})
```

Sending `metadata` again replaces the whole object rather than merging
into it.

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
  {trial_period_days = 14, metadata = {plan = "byok"}}
)
```

Options: `trial_period_days`, `metadata`, `payment_behavior`,
`proration_behavior`, `default_payment_method`, `collection_method`.

Stripe rejects two items sharing one price on the same subscription. Bill
several of the same thing with `quantity`, or give each one its own
subscription.

**Get a subscription:**
```duso
subscription = client.subscriptions.get("sub_XXXXX")
```

**Update a subscription:**
```duso
subscription = client.subscriptions.update("sub_XXXXX", {
  items = [{price = "price_XXXXX", quantity = 2}],
  proration_behavior = "none"
})
```

Options: `items`, `metadata`, `cancel_at_period_end`, `proration_behavior`,
`default_payment_method`, `collection_method`, `trial_end`.

**Stop a subscription renewing:**
```duso
subscription = client.subscriptions.cancel_at_period_end("sub_XXXXX")
```

The customer keeps what they paid for until the period ends, and Stripe
moves the status to `canceled` at the boundary on its own. `resume()`
undoes it while the period is still running:

```duso
subscription = client.subscriptions.resume("sub_XXXXX")
```

**Cancel a subscription immediately:**
```duso
subscription = client.subscriptions.cancel("sub_XXXXX")
subscription = client.subscriptions.cancel("sub_XXXXX", {invoice_now = true})
```

This ends it mid-period. Reach for `cancel_at_period_end` unless the
subscription really should stop now — this one strands a customer who has
already paid through the period.

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

### Checkout Sessions

The hosted payment page. Stripe collects the card on its own domain, so
the card never reaches your server.

**Create a session:**
```duso
session = client.checkout.sessions.create({
  mode = "subscription",
  customer = "cus_XXXXX",
  line_items = [{price = "price_XXXXX", quantity = 1}],
  success_url = "https://example.com/done?session={CHECKOUT_SESSION_ID}",
  cancel_url = "https://example.com/cancelled",
  metadata = {order = "1234"}
})
print(session.url)   // send the customer here
```

`mode` is `"payment"`, `"subscription"` or `"setup"`. `line_items` and
`success_url` are required. `{CHECKOUT_SESSION_ID}` in `success_url` is
substituted by Stripe. Other options: `cancel_url`, `customer`,
`customer_email`, `client_reference_id`, `metadata`, `subscription_data`,
`allow_promotion_codes`.

Metadata on the session comes back on the `checkout.session.completed`
webhook, which is the usual way to know what a payment was *for*.
`subscription_data.metadata` lands on the subscription instead, where it
survives for the life of the subscription.

**Get a session:**
```duso
session = client.checkout.sessions.get("cs_XXXXX")
```

Check `session.payment_status == "paid"` before treating it as complete.

**Expire / list:**
```duso
client.checkout.sessions.expire("cs_XXXXX")
sessions = client.checkout.sessions.list(10)
```

## Webhooks

Verifying an inbound event is a pure function of the raw body, the
signature header and the endpoint secret, so it does not need an API key
and does not live on the client:

```duso
event = stripe.webhooks.construct_event(payload, sig_header, secret)
print(event.type)
```

`secret` defaults to `STRIPE_WEBHOOK_SECRET`. A fourth argument sets the
replay tolerance in seconds (default 300). Anything wrong — bad signature,
missing header, stale timestamp — throws, so a failed verification can
never be mistaken for a valid event.

**The payload must be the exact bytes Stripe sent.** Parsing it to JSON
and re-encoding reorders keys and changes whitespace; the signature is
over bytes, so a round-tripped body never verifies. In an `http_server()`
handler that means `req.body`, untouched:

```duso
ctx = context()
req = ctx.request()
res = ctx.response()

try
  event = stripe.webhooks.construct_event(
    req.body,
    req.headers["Stripe-Signature"],
    env("STRIPE_WEBHOOK_SECRET")
  )
catch (e)
  res.json({error = "invalid signature"}, 400)
end

res.json({received = true}, 200)
```

Get the endpoint secret from the Stripe dashboard when you add the
endpoint, or from `stripe listen` when forwarding to a local server.

## Error Handling

API errors throw exceptions with descriptive messages:

```duso
try
  customer = client.customers.get("invalid_id")
catch (e)
  print("Error: " + e)
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

- No Connect, Terminal, Issuing, tax or quote endpoints.
- Invoices can be read, finalized and paid, but not created or edited.
- List calls return one page. There is no auto-pagination — pass
  `starting_after` where the endpoint accepts it.
- No idempotency keys. A retried write can create a second object.

## Testing

Test against a sandbox, never a live account. A sandbox is disposable and
its keys die with it:

```bash
export STRIPE_API_KEY="sk_test_..."
```

Use test card numbers on the Checkout page:
- `4242 4242 4242 4242` - Visa, succeeds
- `4000 0000 0000 9995` - declines for insufficient funds
- `4000 0025 0000 3155` - requires 3D Secure authentication

Any future expiry, any CVC, any postcode.

To exercise webhooks against a server on localhost, forward them with the
Stripe CLI — it prints the endpoint secret to set as
`STRIPE_WEBHOOK_SECRET`:

```bash
stripe listen --forward-to localhost:8090/stripe/webhook
stripe trigger checkout.session.completed
```

See [Stripe Testing Documentation](https://stripe.com/docs/testing) for more test cards.
