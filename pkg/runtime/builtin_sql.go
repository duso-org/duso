package runtime

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// sqlConn wraps *sql.DB with namespace-level configuration
type sqlConn struct {
	db            *sql.DB
	returnObjects bool
}

// Global registry of SQL connections keyed by namespace
var (
	sqlRegistry  = make(map[string]*sqlConn)
	sqlRegistryMu sync.RWMutex
)

// builtinSQL creates a namespaced SQL connection pool for MySQL-compatible databases.
//
// sql(namespace, config) or sql(namespace) returns a connection object with methods:
//   - .query(sql [, values] [, return_objects]) - Execute SELECT, return rows
//   - .exec(sql [, values]) - Execute INSERT/UPDATE/DELETE, return rows affected
//   - .ping() - Test connection, return true/false
//   - .close() - Close connection and remove from registry
//
// Configuration options:
//   - driver (string) - "mysql", "mariadb", or "tidb" (all use MySQL protocol)
//   - host (string) - database host, default "localhost"
//   - port (number) - database port, default 4000 (TiDB); use 3306 for MySQL/MariaDB
//   - database (string) - database name
//   - user (string) - database user
//   - password (string) - database password
//   - max_open_conns (number) - max concurrent connections, default 25
//   - max_idle_conns (number) - max idle connections, default 5
//   - conn_max_lifetime (number) - connection max lifetime in seconds, default 300
//   - return_objects (bool) - default row format (true = objects, false = arrays), default true
//   - dsn (string) - raw DSN string, overrides all other connection parameters
//
// Example:
//   db = sql("users", {driver = "mysql", host = "localhost", port = 4000, database = "myapp", user = "root"})
//   rows = db.query("SELECT id, name FROM users WHERE id = ?", [42])
//   n = db.exec("UPDATE users SET seen = ? WHERE id = ?", [timestamp(), 42])
func builtinSQL(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get namespace
	var namespace string
	if ns, ok := args["0"]; ok {
		namespace = fmt.Sprintf("%v", ns)
	} else if ns, ok := args["namespace"]; ok {
		namespace = fmt.Sprintf("%v", ns)
	} else {
		return nil, fmt.Errorf("sql() requires a namespace argument")
	}

	// Get config (optional)
	var config map[string]any
	if cfg, ok := args["1"]; ok {
		if cfgMap, ok := cfg.(map[string]any); ok {
			config = cfgMap
		}
	} else if cfg, ok := args["config"]; ok {
		if cfgMap, ok := cfg.(map[string]any); ok {
			config = cfgMap
		}
	}

	// If no config, retrieve existing connection
	if config == nil {
		sqlRegistryMu.RLock()
		conn, exists := sqlRegistry[namespace]
		sqlRegistryMu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("sql(%q) does not exist and no configuration provided", namespace)
		}

		return buildSQLObject(evaluator, conn)
	}

	// Config provided: create new connection
	conn, err := createSQLConnection(config)
	if err != nil {
		return nil, err
	}

	// Store in registry
	sqlRegistryMu.Lock()
	sqlRegistry[namespace] = conn
	sqlRegistryMu.Unlock()

	return buildSQLObject(evaluator, conn)
}

// createSQLConnection builds and opens a database connection from config
func createSQLConnection(config map[string]any) (*sqlConn, error) {
	// Check for raw DSN override
	if rawDSN, ok := config["dsn"]; ok {
		dsn := fmt.Sprintf("%v", rawDSN)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open SQL connection: %w", err)
		}

		returnObjects := true
		if ro, ok := config["return_objects"]; ok {
			if b, ok := ro.(bool); ok {
				returnObjects = b
			}
		}

		configurePool(db, config)
		return &sqlConn{db: db, returnObjects: returnObjects}, nil
	}

	// Build DSN from fields
	user := getConfigString(config, "user", "root")
	password := getConfigString(config, "password", "")
	host := getConfigString(config, "host", "localhost")
	port := getConfigFloat(config, "port", 4000)
	database := getConfigString(config, "database", "")

	// Build DSN with parseTime=true so time.Time values are scanned correctly
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, int(port), database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQL connection: %w", err)
	}

	returnObjects := true
	if ro, ok := config["return_objects"]; ok {
		if b, ok := ro.(bool); ok {
			returnObjects = b
		}
	}

	configurePool(db, config)
	return &sqlConn{db: db, returnObjects: returnObjects}, nil
}

// configurePool applies pool settings from config to the database connection
func configurePool(db *sql.DB, config map[string]any) {
	maxOpen := int(getConfigFloat(config, "max_open_conns", 25))
	maxIdle := int(getConfigFloat(config, "max_idle_conns", 5))
	maxLifetime := int(getConfigFloat(config, "conn_max_lifetime", 300))

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
}

// getConfigString safely retrieves a string config value
func getConfigString(config map[string]any, key string, defaultVal string) string {
	if val, ok := config[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return defaultVal
}

// getConfigFloat safely retrieves a float64 config value
func getConfigFloat(config map[string]any, key string, defaultVal float64) float64 {
	if val, ok := config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultVal
}

// buildSQLObject returns a map with query, exec, ping, and close methods
func buildSQLObject(evaluator *Evaluator, conn *sqlConn) (any, error) {
	// query(sql, [values], [return_objects]) method
	queryFn := NewGoFunction(func(qEval *Evaluator, qArgs map[string]any) (any, error) {
		return handleQuery(conn, qArgs)
	})

	// exec(sql, [values]) method
	execFn := NewGoFunction(func(eEval *Evaluator, eArgs map[string]any) (any, error) {
		return handleExec(conn, eArgs)
	})

	// ping() method
	pingFn := NewGoFunction(func(pEval *Evaluator, pArgs map[string]any) (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := conn.db.PingContext(ctx)
		return err == nil, nil
	})

	// close() method
	closeFn := NewGoFunction(func(cEval *Evaluator, cArgs map[string]any) (any, error) {
		return nil, conn.db.Close()
	})

	return map[string]any{
		"query": queryFn,
		"exec":  execFn,
		"ping":  pingFn,
		"close": closeFn,
	}, nil
}

// handleQuery executes a SELECT query and returns rows
func handleQuery(conn *sqlConn, args map[string]any) (any, error) {
	// Extract SQL
	var query string
	if q, ok := args["0"].(string); ok {
		query = q
	} else if q, ok := args["query"].(string); ok {
		query = q
	} else {
		return nil, fmt.Errorf("query() requires a SQL string as first argument")
	}

	// Extract params (optional)
	params, err := extractParams(args)
	if err != nil {
		return nil, err
	}

	// Extract return_objects flag (optional, defaults to conn.returnObjects, then true)
	returnObjects := conn.returnObjects
	if ro, ok := args["return_objects"]; ok {
		if b, ok := ro.(bool); ok {
			returnObjects = b
		}
	} else if ro, ok := args["2"]; ok {
		if b, ok := ro.(bool); ok {
			returnObjects = b
		}
	}

	// Execute query
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := conn.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, &DusoError{
			Message: fmt.Sprintf("query error: %v", err),
		}
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}

	// Scan rows
	result := make([]any, 0)

	for rows.Next() {
		// Prepare destination slices for each column
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan into pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert and add to result
		if returnObjects {
			// Return as object {col1: val1, col2: val2, ...}
			rowObj := make(map[string]any)
			for i, col := range columns {
				rowObj[col] = convertSQLValue(values[i])
			}
			result = append(result, rowObj)
		} else {
			// Return as array [val1, val2, ...]
			rowArr := make([]any, len(values))
			for i, val := range values {
				rowArr[i] = convertSQLValue(val)
			}
			result = append(result, rowArr)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// handleExec executes an INSERT/UPDATE/DELETE query and returns rows affected
func handleExec(conn *sqlConn, args map[string]any) (any, error) {
	// Extract SQL
	var query string
	if q, ok := args["0"].(string); ok {
		query = q
	} else if q, ok := args["query"].(string); ok {
		query = q
	} else {
		return nil, fmt.Errorf("exec() requires a SQL string as first argument")
	}

	// Extract params (optional)
	params, err := extractParams(args)
	if err != nil {
		return nil, err
	}

	// Execute query
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := conn.db.ExecContext(ctx, query, params...)
	if err != nil {
		return nil, &DusoError{
			Message: fmt.Sprintf("exec error: %v", err),
		}
	}

	// Get rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return float64(rowsAffected), nil
}

// extractParams extracts and converts parameter values from arguments
func extractParams(args map[string]any) ([]interface{}, error) {
	var params []interface{}

	// Check for named "values" parameter first, then positional "1"
	var paramsArg any
	var hasParams bool

	if v, ok := args["values"]; ok {
		paramsArg = v
		hasParams = true
	} else if v, ok := args["1"]; ok {
		paramsArg = v
		hasParams = true
	}

	if !hasParams {
		return params, nil // No params provided
	}

	// If it's an array, iterate over it
	if arrPtr, ok := paramsArg.(*[]Value); ok {
		for _, v := range *arrPtr {
			params = append(params, convertDusoValue(v))
		}
		return params, nil
	}

	// If it's a single non-array value, wrap it
	if paramsArg != nil {
		// Convert to Value to handle it uniformly
		v := InterfaceToValue(paramsArg)
		params = append(params, convertDusoValue(v))
		return params, nil
	}

	return params, nil
}

// convertDusoValue converts a duso Value to a Go value for SQL parameters
func convertDusoValue(v Value) interface{} {
	switch v.Type {
	case VAL_NUMBER:
		return v.AsNumber()
	case VAL_STRING:
		return v.AsString()
	case VAL_BOOL:
		return v.AsBool()
	case VAL_NIL:
		return nil
	default:
		// For complex types, convert to string representation
		return v.String()
	}
}

// convertSQLValue converts a scanned SQL value to a duso-compatible type
func convertSQLValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case bool:
		return v
	case []byte:
		// Convert byte slice to string
		return string(v)
	case string:
		return v
	case time.Time:
		// Convert to Unix timestamp (float64) in UTC
		return float64(v.UTC().Unix())
	default:
		// Fallback: convert to string
		return fmt.Sprintf("%v", v)
	}
}
