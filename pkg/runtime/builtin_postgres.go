package runtime

import (
	"context"
	"fmt"
	"time"
)

// builtinPostgres creates or returns a postgres connection pool and object with methods.
//
// postgres() returns a namespaced database connection with methods:
//   - .query(sql, ...params) - Execute a SELECT query, returns array
//   - .query(sql, ...params, {return_objects = true}) - Same, but returns array of objects
//   - .exec(sql, ...params) - Execute INSERT/UPDATE/DELETE, returns affected row count
//   - .begin() - Start a transaction, returns transaction object with query/exec/commit/rollback
//
// Configuration options:
//   - host (string) - Database host, default "localhost"
//   - port (number) - Database port, default 5432
//   - database (string) - Database name
//   - user (string) - Database user
//   - password (string) - Database password
//   - pool_size (number) - Max connections, default 10
//   - idle_timeout (number) - Idle timeout in seconds
//   - connection_timeout (number) - Connection timeout in seconds
//   - return_objects (bool) - Default return format for queries (true = objects, false = arrays)
//   - sslmode (string) - SSL mode: disable, allow, prefer, require, verify-ca, verify-full (default: disable)
//   - sslcert (string) - Path to client certificate file
//   - sslkey (string) - Path to client key file
//   - sslrootcert (string) - Path to root CA certificate file
//
// Subsequent calls to postgres(resourceName, config) reuse the connection but merge new config.
//
// Example:
//
//	db = postgres("mydb", {
//	  host = "localhost",
//	  database = "myapp",
//	  user = "app",
//	  password = "secret",
//	  pool_size = 10
//	})
//
//	rows = db.query("SELECT * FROM users WHERE active = $1", true)
//	for row in rows then
//	  print(row[0])  // First column
//	end
//
//	objects = db.query("SELECT * FROM users WHERE active = $1", true, {return_objects = true})
//	for obj in objects then
//	  print(obj.name)  // By column name
//	end
//
//	count = db.exec("UPDATE users SET active = $1 WHERE id = $2", false, 123)
//
//	tx = db.begin()
//	tx.exec("INSERT INTO logs VALUES ($1, $2)", "event", "data")
//	tx.commit()
func builtinPostgres(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get resource name from first positional or named argument
	var resourceName string

	if rn, ok := args["0"]; ok {
		resourceName = fmt.Sprintf("%v", rn)
	} else if rn, ok := args["resource"]; ok {
		resourceName = fmt.Sprintf("%v", rn)
	} else {
		return nil, fmt.Errorf("postgres() requires a resource name argument")
	}

	if resourceName == "" {
		return nil, fmt.Errorf("postgres() resource name cannot be empty")
	}

	// Get config from second positional or named argument (optional)
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

	// Get or create the postgres connection
	pgConn, err := GetPostgresConnection(resourceName, config)
	if err != nil {
		return nil, err
	}

	// Get default return_objects setting from config
	returnObjects := false
	if pgConn.config != nil {
		if ro, ok := pgConn.config["return_objects"]; ok {
			if rob, ok := ro.(bool); ok {
				returnObjects = rob
			}
		}
	}

	// Create query(sql, ...args) method
	queryFn := NewGoFunction(func(queryEval *Evaluator, queryArgs map[string]any) (any, error) {
		// Extract SQL (first positional arg)
		var sql string
		if s, ok := queryArgs["0"]; ok {
			sql = fmt.Sprintf("%v", s)
		} else {
			return nil, fmt.Errorf("query() requires a SQL string")
		}

		// Extract params (remaining positional args)
		var params []any
		for i := 1; ; i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := queryArgs[key]; ok {
				params = append(params, val)
			} else {
				break
			}
		}

		// Check for return_objects override in final arg (if it's a map, it's config)
		useReturnObjects := returnObjects
		if len(params) > 0 {
			if lastArg, ok := params[len(params)-1].(map[string]any); ok {
				if ro, ok := lastArg["return_objects"]; ok {
					if rob, ok := ro.(bool); ok {
						useReturnObjects = rob
						// Remove config from params
						params = params[:len(params)-1]
					}
				}
			}
		}

		// Execute query with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if useReturnObjects {
			rows, err := pgConn.Query(ctx, sql, params...)
			if err != nil {
				return map[string]any{"error": err.Error()}, nil
			}
			// Convert []map[string]any to []any
			result := make([]any, len(rows))
			for i, row := range rows {
				result[i] = row
			}
			return result, nil
		} else {
			rows, err := pgConn.QueryRaw(ctx, sql, params...)
			if err != nil {
				return map[string]any{"error": err.Error()}, nil
			}
			// Convert [][]any to []any
			result := make([]any, len(rows))
			for i, row := range rows {
				result[i] = row
			}
			return result, nil
		}
	})

	// Create exec(sql, ...args) method
	execFn := NewGoFunction(func(execEval *Evaluator, execArgs map[string]any) (any, error) {
		// Extract SQL (first positional arg)
		var sql string
		if s, ok := execArgs["0"]; ok {
			sql = fmt.Sprintf("%v", s)
		} else {
			return nil, fmt.Errorf("exec() requires a SQL string")
		}

		// Extract params (remaining positional args)
		var params []any
		for i := 1; ; i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := execArgs[key]; ok {
				params = append(params, val)
			} else {
				break
			}
		}

		// Execute with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		count, err := pgConn.Exec(ctx, sql, params...)
		if err != nil {
			return map[string]any{"error": err.Error()}, nil
		}

		return count, nil
	})

	// Create begin() method
	beginFn := NewGoFunction(func(beginEval *Evaluator, beginArgs map[string]any) (any, error) {
		// Use background context for transaction (no timeout, operations handle their own)
		ctx := context.Background()

		tx, err := pgConn.BeginTx(ctx)
		if err != nil {
			return nil, err
		}

		// Create transaction query method
		txQueryFn := NewGoFunction(func(txQueryEval *Evaluator, txQueryArgs map[string]any) (any, error) {
			var sql string
			if s, ok := txQueryArgs["0"]; ok {
				sql = fmt.Sprintf("%v", s)
			} else {
				return nil, fmt.Errorf("query() requires a SQL string")
			}

			var params []any
			for i := 1; ; i++ {
				key := fmt.Sprintf("%d", i)
				if val, ok := txQueryArgs[key]; ok {
					params = append(params, val)
				} else {
					break
				}
			}

			// Check for return_objects override
			useReturnObjects := returnObjects
			if len(params) > 0 {
				if lastArg, ok := params[len(params)-1].(map[string]any); ok {
					if ro, ok := lastArg["return_objects"]; ok {
						if rob, ok := ro.(bool); ok {
							useReturnObjects = rob
							params = params[:len(params)-1]
						}
					}
				}
			}

			if useReturnObjects {
				rows, err := tx.Query(sql, params...)
				if err != nil {
					return map[string]any{"error": err.Error()}, nil
				}

				result := make([]any, len(rows))
				for i, row := range rows {
					result[i] = row
				}
				return result, nil
			} else {
				rows, err := tx.QueryRaw(sql, params...)
				if err != nil {
					return map[string]any{"error": err.Error()}, nil
				}

				result := make([]any, len(rows))
				for i, row := range rows {
					result[i] = row
				}
				return result, nil
			}
		})

		// Create transaction exec method
		txExecFn := NewGoFunction(func(txExecEval *Evaluator, txExecArgs map[string]any) (any, error) {
			var sql string
			if s, ok := txExecArgs["0"]; ok {
				sql = fmt.Sprintf("%v", s)
			} else {
				return nil, fmt.Errorf("exec() requires a SQL string")
			}

			var params []any
			for i := 1; ; i++ {
				key := fmt.Sprintf("%d", i)
				if val, ok := txExecArgs[key]; ok {
					params = append(params, val)
				} else {
					break
				}
			}

			count, err := tx.Exec(sql, params...)
			if err != nil {
				return map[string]any{"error": err.Error()}, nil
			}

			return count, nil
		})

		// Create transaction commit method
		commitFn := NewGoFunction(func(commitEval *Evaluator, commitArgs map[string]any) (any, error) {
			err := tx.Commit()
			if err != nil {
				return nil, err
			}
			return nil, nil
		})

		// Create transaction rollback method
		rollbackFn := NewGoFunction(func(rollbackEval *Evaluator, rollbackArgs map[string]any) (any, error) {
			err := tx.Rollback()
			if err != nil {
				return nil, err
			}
			return nil, nil
		})

		// Return transaction object with methods
		return map[string]any{
			"query":    txQueryFn,
			"exec":     txExecFn,
			"commit":   commitFn,
			"rollback": rollbackFn,
		}, nil
	})

	// Return postgres object with methods
	return map[string]any{
		"query": queryFn,
		"exec":  execFn,
		"begin": beginFn,
	}, nil
}
