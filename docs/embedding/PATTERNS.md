# Common Embedding Patterns

Design patterns and use cases for embedding Duso in applications.

## 1. Configuration DSL

Use Duso to replace configuration files (YAML, TOML, JSON).

**Advantages:**
- Turing-complete configuration language
- Can compute values dynamically
- Supports validation logic
- Easy to understand for developers

**Example:**

```go
// config.du
database = {
    host = "localhost",
    port = 5432,
    username = "admin",
    password = "secret"
}

server = {
    port = 8080,
    timeout = 30,
    maxConnections = 100
}

environment = "development"
debug = environment == "development"
```

```go
// main.go
interp := script.NewInterpreter(false)

configFile, _ := os.ReadFile("config.du")
interp.Execute(string(configFile))

database := interp.GetVariable("database")
server := interp.GetVariable("server")
debug := interp.GetVariable("debug")

// Use configuration...
```

## 2. Plugin System

Let users extend your application with scripts.

```go
type PluginManager struct {
    interp *script.Interpreter
}

func (pm *PluginManager) LoadPlugin(path string) error {
    content, _ := os.ReadFile(path)
    return pm.interp.Execute(string(content))
}

func (pm *PluginManager) CallHook(hookName string, data any) (any, error) {
    return pm.interp.Call(hookName, data)
}

// User plugin:
// function onUserLogin(user)
//     print("User logged in: " + user.name)
// end
```

## 3. Data Transformation Pipelines

Build ETL/transformation systems.

```go
interp.RegisterFunction("fetch", fetchFromAPI)
interp.RegisterFunction("transform", transformData)
interp.RegisterFunction("validate", validateData)
interp.RegisterFunction("store", storeInDatabase)

// User writes pipeline
pipeline := `
    raw = fetch(url = "https://api.example.com/data")
    cleaned = transform(data = raw)
    valid = validate(data = cleaned)

    if valid then
        result = store(data = valid)
        print("Stored: " + len(result))
    end
`

interp.Execute(pipeline)
```

## 4. Workflow Orchestration

Coordinate complex multi-step processes.

```go
// Register services as functions
interp.RegisterFunction("creditCard", chargeCreditCard)
interp.RegisterFunction("inventory", updateInventory)
interp.RegisterFunction("email", sendEmail)
interp.RegisterFunction("notification", sendNotification)

// User defines workflow
checkout := `
    function processOrder(order)
        // Step 1: Charge card
        payment = creditCard(amount = order.total)
        if not payment.success then
            return {error = "Payment failed"}
        end

        // Step 2: Update inventory
        for item in order.items do
            inventory(itemID = item.id, quantity = item.qty)
        end

        // Step 3: Send confirmation
        email(to = order.email, subject = "Order confirmed")
        notification(userID = order.userID, message = "Your order is confirmed")

        return {success = true, orderID = payment.id}
    end

    result = processOrder(order = inputOrder)
`
```

## 5. Rule Engine

Implement flexible business logic.

```go
// Register data access functions
interp.RegisterFunction("getUser", getUser)
interp.RegisterFunction("getAccount", getAccount)
interp.RegisterFunction("getRules", getRules)

// User defines rules
ruleScript := `
    function checkApproval(transaction)
        rules = getRules()

        if transaction.amount > rules.maxDaily then
            return false
        end

        user = getUser(id = transaction.userID)
        if user.riskScore > 0.8 then
            return false
        end

        return true
    end
`

interp.Execute(ruleScript)
result, _ := interp.Call("checkApproval", transaction)
```

## 6. Template System

Use Duso as a powerful template engine.

```go
// Register data providers
interp.RegisterFunction("getUser", getUser)
interp.RegisterFunction("getOrders", getOrders)

// Template script
templateScript := `
    user = getUser(id = userID)
    orders = getOrders(userID = userID)

    output = """
    Name: {{user.name}}
    Email: {{user.email}}

    Recent Orders:
    """

    for order in orders do
        output = output + """
        - Order #{{order.id}}: {{format_time(order.date)}}
        """
    end

    return output
`

interp.Execute(templateScript)
result := interp.GetVariable("output")
```

## 7. Job Scheduler

Schedule and configure jobs dynamically.

```go
type Job struct {
    Name    string
    Script  string
    Schedule string
}

func ScheduleJob(job Job) {
    interp := script.NewInterpreter(false)

    // Set job context
    interp.SetVariable("jobName", job.Name)

    // Register job-specific functions
    interp.RegisterFunction("log", logJobMessage)
    interp.RegisterFunction("sendAlert", sendAlert)

    // Execute job script
    interp.Execute(job.Script)
}

// Job definition
jobScript := `
    print("Running " + jobName)

    try
        result = processData()
        log(message = "Success: " + result)
    catch (err)
        log(message = "Error: " + err)
        sendAlert(severity = "high", message = err)
    end
`
```

## 8. API Response Handler

Process API responses dynamically.

```go
interp.RegisterFunction("callAPI", callExternalAPI)
interp.RegisterFunction("storeMetrics", storeMetrics)
interp.RegisterFunction("alertIfNeeded", alertIfNeeded)

// Handler script
handler := `
    function handleMetrics(source)
        response = callAPI(endpoint = source)
        data = parse_json(response)

        // Store metrics
        storeMetrics(metrics = data.metrics)

        // Check for anomalies
        for metric in data.metrics do
            if metric.value > metric.threshold then
                alertIfNeeded(metric = metric.name, value = metric.value)
            end
        end

        return {processed = len(data.metrics)}
    end
`

interp.Execute(handler)
```

## 9. Form Validation

Implement complex validation logic.

```go
interp.RegisterFunction("validateEmail", validateEmail)
interp.RegisterFunction("checkUsername", checkUsername)

validationScript := `
    function validateForm(formData)
        errors = {}

        // Validate each field
        if len(formData.email) == 0 then
            push(errors, "Email is required")
        elseif not validateEmail(email = formData.email) then
            push(errors, "Invalid email format")
        end

        if len(formData.username) < 3 then
            push(errors, "Username must be at least 3 characters")
        elseif checkUsername(username = formData.username).exists then
            push(errors, "Username already taken")
        end

        if len(formData.password) < 8 then
            push(errors, "Password must be at least 8 characters")
        end

        return {
            valid = len(errors) == 0,
            errors = errors
        }
    end
`
```

## 10. Multi-Tenant Customization

Different behavior per tenant.

```go
func ExecuteTenantScript(tenantID string, script string) {
    interp := script.NewInterpreter(false)

    // Load tenant-specific configuration
    tenantConfig := loadTenantConfig(tenantID)
    interp.SetVariable("tenantConfig", tenantConfig)

    // Register tenant-specific functions
    interp.RegisterFunction("getTenantData",
        func(args map[string]any) (any, error) {
            return getTenantDataForID(tenantID, args)
        })

    // Execute tenant script
    interp.Execute(script)
}

// Different tenants can have different logic
tenantAScript := `
    if tenantConfig.discountRate > 0 then
        price = price * (1 - tenantConfig.discountRate)
    end
`

tenantBScript := `
    if tenantConfig.taxEnabled then
        price = price * (1 + tenantConfig.taxRate)
    end
`
```

## Choosing a Pattern

- **Configuration?** → Configuration DSL (Pattern 1)
- **User Extensions?** → Plugin System (Pattern 2)
- **Data Processing?** → Transformation Pipeline (Pattern 3)
- **Multi-Step Process?** → Workflow Orchestration (Pattern 4)
- **Business Rules?** → Rule Engine (Pattern 5)
- **Dynamic Text?** → Template System (Pattern 6)
- **Scheduled Tasks?** → Job Scheduler (Pattern 7)
- **API Processing?** → Response Handler (Pattern 8)
- **Input Validation?** → Form Validation (Pattern 9)
- **Multiple Customers?** → Multi-Tenant (Pattern 10)

## Best Practices

1. **Clear API** - Register functions with clear, consistent names
2. **Error Boundaries** - Wrap execution in try/catch in Duso
3. **Timeout Protection** - Prevent scripts from running forever (handle in Go)
4. **Input Validation** - Validate arguments in your Go functions
5. **Logging** - Log execution for debugging
6. **Security** - Don't expose dangerous operations; users can only do what you register
7. **Performance** - Measure script execution time; optimize hot paths
8. **Documentation** - Document what functions you register and what they do

## See Also

- [Custom Functions Guide](/docs/embedding/custom-functions.md) - More function examples
- [API Reference](/docs/embedding/api-reference.md) - Full API documentation
- [Examples](/docs/embedding/examples.md) - Complete application examples
