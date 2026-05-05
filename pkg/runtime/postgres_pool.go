package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/duso-org/duso/pkg/script"
)

// Global registry of postgres connection pools, keyed by resource name
var (
	postgresRegistry = make(map[string]*PostgresConnection)
	postgresLock     sync.RWMutex
)

// PostgresConnection wraps a pgx connection pool with resource name and config
type PostgresConnection struct {
	resourceName string
	pool         *pgxpool.Pool
	config       map[string]any
	configMutex  sync.RWMutex
}

// GetPostgresConnection returns or creates a postgres connection pool for the given resource name.
// If config is provided, it merges with existing config (for subsequent calls to update settings).
func GetPostgresConnection(resourceName string, config map[string]any) (*PostgresConnection, error) {
	postgresLock.Lock()
	defer postgresLock.Unlock()

	// If connection exists, merge new config if provided
	if conn, exists := postgresRegistry[resourceName]; exists {
		if config != nil && len(config) > 0 {
			conn.configMutex.Lock()
			// Merge new config into existing config
			for k, v := range config {
				conn.config[k] = v
			}
			conn.configMutex.Unlock()
		}
		return conn, nil
	}

	// Create new connection
	if config == nil {
		config = make(map[string]any)
	}

	conn := &PostgresConnection{
		resourceName: resourceName,
		config:       config,
	}

	// Connect to database
	pool, err := createPool(config)
	if err != nil {
		return nil, err
	}

	conn.pool = pool

	// Store in registry
	postgresRegistry[resourceName] = conn

	return conn, nil
}

// createPool creates a pgx connection pool from config
func createPool(config map[string]any) (*pgxpool.Pool, error) {
	// Build connection string from config
	var host, database, user, password string
	var port int64 = 5432

	if h, ok := config["host"]; ok {
		host = fmt.Sprintf("%v", h)
	}
	if d, ok := config["database"]; ok {
		database = fmt.Sprintf("%v", d)
	}
	if u, ok := config["user"]; ok {
		user = fmt.Sprintf("%v", u)
	}
	if p, ok := config["password"]; ok {
		password = fmt.Sprintf("%v", p)
	}
	if pt, ok := config["port"]; ok {
		if portNum, ok := pt.(float64); ok {
			port = int64(portNum)
		}
	}

	// Build DSN with query parameters
	params := []string{}

	// SSL settings (default to disable for local dev, can be overridden)
	sslmode := "disable"
	if sm, ok := config["sslmode"]; ok {
		sslmode = fmt.Sprintf("%v", sm)
	}
	params = append(params, fmt.Sprintf("sslmode=%s", sslmode))

	// Optional SSL certificate paths
	if cert, ok := config["sslcert"]; ok {
		params = append(params, fmt.Sprintf("sslcert=%v", cert))
	}
	if key, ok := config["sslkey"]; ok {
		params = append(params, fmt.Sprintf("sslkey=%v", key))
	}
	if rootcert, ok := config["sslrootcert"]; ok {
		params = append(params, fmt.Sprintf("sslrootcert=%v", rootcert))
	}

	queryString := strings.Join(params, "&")
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?%s",
		user, password, host, port, database, queryString)

	// Create config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("invalid postgres config: %w", err)
	}

	// Apply pool settings from config
	if poolSize, ok := config["pool_size"]; ok {
		if size, ok := poolSize.(float64); ok {
			poolConfig.MaxConns = int32(size)
		}
	}

	if idleTimeout, ok := config["idle_timeout"]; ok {
		if timeout, ok := idleTimeout.(float64); ok {
			poolConfig.MaxConnIdleTime = time.Duration(timeout) * time.Second
		}
	}

	if connTimeout, ok := config["connection_timeout"]; ok {
		if timeout, ok := connTimeout.(float64); ok {
			poolConfig.ConnConfig.ConnectTimeout = time.Duration(timeout) * time.Second
		}
	}

	// Create pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}

	return pool, nil
}

// convertParam converts Duso values back to Go types for postgres parameters
func convertParam(v any) any {
	if v == nil {
		return nil
	}

	// If it's a script.Value, extract the underlying Go value
	if val, ok := v.(script.Value); ok {
		switch val.Type {
		case script.VAL_NIL:
			return nil
		case script.VAL_NUMBER:
			return val.AsNumber()
		case script.VAL_STRING:
			return val.AsString()
		case script.VAL_BOOL:
			return val.AsBool()
		case script.VAL_BINARY:
			bv := val.AsBinary()
			if bv.Data != nil {
				return *bv.Data
			}
			return nil
		case script.VAL_ARRAY:
			arr := val.AsArray()
			result := make([]any, len(arr))
			for i, elem := range arr {
				result[i] = convertParam(elem)
			}
			return result
		case script.VAL_OBJECT:
			obj := val.AsObject()
			result := make(map[string]any)
			for k, elem := range obj {
				result[k] = convertParam(elem)
			}
			return result
		default:
			return nil
		}
	}

	// For raw Go values, return as-is
	return v
}

// convertValue converts postgres values to Duso-compatible types
// Duso only has float64 for numbers, so we convert all numeric types to float64
// Timestamps are converted to unix time (seconds since epoch)
// Binary data is converted to Duso BinaryValue with metadata
func convertValue(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case bool:
		return val
	case string:
		return val
	case []byte:
		// Create a BinaryValue with size metadata
		dataCopy := make([]byte, len(val))
		copy(dataCopy, val)
		return script.Value{
			Type: script.VAL_BINARY,
			Data: &script.BinaryValue{
				Data: &dataCopy,
				Metadata: map[string]script.Value{
					"size": script.NewNumber(float64(len(val))),
				},
			},
		}
	case time.Time:
		return float64(val.Unix())
	default:
		return v
	}
}

// Query executes a query and returns rows
func (pc *PostgresConnection) Query(ctx context.Context, sql string, args ...any) ([]map[string]any, error) {
	// Convert Duso values to Go types for postgres
	convertedArgs := make([]any, len(args))
	for i, arg := range args {
		convertedArgs[i] = convertParam(arg)
	}

	rows, err := pc.pool.Query(ctx, sql, convertedArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		cols := rows.FieldDescriptions()
		for i, col := range cols {
			if i < len(values) {
				row[col.Name] = convertValue(values[i])
			}
		}
		result = append(result, row)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

// QueryRaw executes a query and returns rows as arrays
func (pc *PostgresConnection) QueryRaw(ctx context.Context, sql string, args ...any) ([][]any, error) {
	// Convert Duso values to Go types for postgres
	convertedArgs := make([]any, len(args))
	for i, arg := range args {
		convertedArgs[i] = convertParam(arg)
	}

	rows, err := pc.pool.Query(ctx, sql, convertedArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result [][]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		// Convert all values for Duso compatibility
		converted := make([]any, len(values))
		for i, v := range values {
			converted[i] = convertValue(v)
		}
		result = append(result, converted)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

// Exec executes a command and returns number of rows affected
func (pc *PostgresConnection) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	// Convert Duso values to Go types for postgres
	convertedArgs := make([]any, len(args))
	for i, arg := range args {
		convertedArgs[i] = convertParam(arg)
	}

	result, err := pc.pool.Exec(ctx, sql, convertedArgs...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// BeginTx starts a transaction
func (pc *PostgresConnection) BeginTx(ctx context.Context) (*PostgresTransaction, error) {
	tx, err := pc.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &PostgresTransaction{
		tx:  tx,
		ctx: ctx,
	}, nil
}

// PostgresTransaction wraps a pgx transaction
type PostgresTransaction struct {
	tx  pgx.Tx
	ctx context.Context
}

// Query executes a query within the transaction, returning array of objects
func (pt *PostgresTransaction) Query(sql string, args ...any) ([]map[string]any, error) {
	// Convert Duso values to Go types for postgres
	convertedArgs := make([]any, len(args))
	for i, arg := range args {
		convertedArgs[i] = convertParam(arg)
	}

	rows, err := pt.tx.Query(pt.ctx, sql, convertedArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		cols := rows.FieldDescriptions()
		for i, col := range cols {
			if i < len(values) {
				row[col.Name] = convertValue(values[i])
			}
		}
		result = append(result, row)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

// QueryRaw executes a query within the transaction, returning arrays
func (pt *PostgresTransaction) QueryRaw(sql string, args ...any) ([][]any, error) {
	// Convert Duso values to Go types for postgres
	convertedArgs := make([]any, len(args))
	for i, arg := range args {
		convertedArgs[i] = convertParam(arg)
	}

	rows, err := pt.tx.Query(pt.ctx, sql, convertedArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result [][]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		// Convert all values for Duso compatibility
		converted := make([]any, len(values))
		for i, v := range values {
			converted[i] = convertValue(v)
		}
		result = append(result, converted)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

// Exec executes a command within the transaction
func (pt *PostgresTransaction) Exec(sql string, args ...any) (int64, error) {
	// Convert Duso values to Go types for postgres
	convertedArgs := make([]any, len(args))
	for i, arg := range args {
		convertedArgs[i] = convertParam(arg)
	}

	result, err := pt.tx.Exec(pt.ctx, sql, convertedArgs...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// Commit commits the transaction
func (pt *PostgresTransaction) Commit() error {
	return pt.tx.Commit(pt.ctx)
}

// Rollback rolls back the transaction
func (pt *PostgresTransaction) Rollback() error {
	return pt.tx.Rollback(pt.ctx)
}
