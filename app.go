package main

import (
	"context"
	"db-desktop/database"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// App struct
type App struct {
	ctx       context.Context
	dbManager *database.SimpleDatabaseManager
	logger    *logrus.Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &App{
		dbManager: database.NewSimpleDatabaseManager(),
		logger:    logger,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load saved connections
	if err := a.dbManager.LoadConnections(); err != nil {
		a.logger.Errorf("Failed to load connections: %v", err)
	}

	a.logger.Info("Database Desktop application started")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Database Connection Management

// AddConnection adds a new database connection
func (a *App) AddConnection(config map[string]interface{}) error {
	a.logger.Infof("AddConnection called with config: %+v", config)

	// Convert map to ConnectionConfig
	connConfig := &database.ConnectionConfig{
		Name:     getString(config, "name"),
		Type:     database.DatabaseType(getString(config, "type")),
		Host:     getString(config, "host"),
		Port:     getInt(config, "port"),
		Username: getString(config, "username"),
		Password: getString(config, "password"),
		Database: getString(config, "database"),
		SSLMode:  getString(config, "sslMode"),
		Timeout:  getInt(config, "timeout"), // Store as seconds
		MaxConns: getInt(config, "maxConns"),
	}

	a.logger.Infof("Converted config: %+v", connConfig)

	// Use a channel to handle the operation with timeout
	resultChan := make(chan error, 1)
	go func() {
		resultChan <- a.dbManager.AddConnection(connConfig)
	}()

	// Wait for result with timeout
	select {
	case err := <-resultChan:
		if err != nil {
			a.logger.Errorf("AddConnection failed: %v", err)
			return err
		}
		a.logger.Infof("AddConnection successful")
		return nil
	case <-time.After(10 * time.Second):
		a.logger.Errorf("AddConnection timeout after 10 seconds")
		return fmt.Errorf("operation timeout")
	}
}

// UpdateConnection updates an existing database connection
func (a *App) UpdateConnection(config map[string]interface{}) error {
	// Convert map to ConnectionConfig
	connConfig := &database.ConnectionConfig{
		ID:       getString(config, "id"),
		Name:     getString(config, "name"),
		Type:     database.DatabaseType(getString(config, "type")),
		Host:     getString(config, "host"),
		Port:     getInt(config, "port"),
		Username: getString(config, "username"),
		Password: getString(config, "password"),
		Database: getString(config, "database"),
		SSLMode:  getString(config, "sslMode"),
		Timeout:  getInt(config, "timeout"), // Store as seconds
		MaxConns: getInt(config, "maxConns"),
	}

	return a.dbManager.UpdateConnection(connConfig)
}

// DeleteConnection deletes a database connection
func (a *App) DeleteConnection(id string) error {
	return a.dbManager.DeleteConnection(id)
}

// GetConnection returns a connection configuration by ID
func (a *App) GetConnection(id string) (*database.ConnectionConfig, error) {
	return a.dbManager.GetConnection(id)
}

// ListConnections returns all connection configurations
func (a *App) ListConnections() []*database.ConnectionConfig {
	return a.dbManager.ListConnections()
}

// Connect establishes a connection to a database
func (a *App) Connect(id string) error {
	return a.dbManager.Connect(id)
}

// Disconnect closes a database connection
func (a *App) Disconnect(id string) error {
	return a.dbManager.Disconnect(id)
}

// TestConnection tests a database connection
func (a *App) TestConnection(config map[string]interface{}) error {
	a.logger.Infof("TestConnection called with config: %+v", config)

	// Convert map to ConnectionConfig
	connConfig := &database.ConnectionConfig{
		Name:     getString(config, "name"),
		Type:     database.DatabaseType(getString(config, "type")),
		Host:     getString(config, "host"),
		Port:     getInt(config, "port"),
		Username: getString(config, "username"),
		Password: getString(config, "password"),
		Database: getString(config, "database"),
		SSLMode:  getString(config, "sslMode"),
		Timeout:  getInt(config, "timeout"), // Store as seconds
		MaxConns: getInt(config, "maxConns"),
	}

	a.logger.Infof("Converted config for test: %+v", connConfig)

	err := a.dbManager.TestConnection(connConfig)
	if err != nil {
		a.logger.Errorf("TestConnection failed: %v", err)
		return err
	}

	a.logger.Infof("TestConnection successful")
	return nil
}

// GetConnectionStatus returns the status of a database connection
func (a *App) GetConnectionStatus(id string) *database.ConnectionStatus {
	return a.dbManager.GetConnectionStatus(id)
}

// Query Operations

// ExecuteQuery executes a query on a database
func (a *App) ExecuteQuery(connectionID string, query string) (*database.QueryResult, error) {
	a.logger.Infof("ExecuteQuery called - ConnectionID: %s, Query: %s", connectionID, query)

	result, err := a.dbManager.ExecuteQuery(connectionID, query)
	if err != nil {
		a.logger.Errorf("ExecuteQuery failed - ConnectionID: %s, Error: %v", connectionID, err)
		return result, err
	}

	a.logger.Infof("ExecuteQuery success - ConnectionID: %s, Rows: %d, Time: %dms",
		connectionID, result.Count, result.Time)

	return result, nil
}

// ExecuteQueryWithLimit executes a query with limit
func (a *App) ExecuteQueryWithLimit(connectionID string, query string, limit int) (*database.QueryResult, error) {
	a.logger.Infof("ExecuteQueryWithLimit called - ConnectionID: %s, Query: %s, Limit: %d", connectionID, query, limit)

	result, err := a.dbManager.ExecuteQueryWithLimit(connectionID, query, limit)
	if err != nil {
		a.logger.Errorf("ExecuteQueryWithLimit failed - ConnectionID: %s, Error: %v", connectionID, err)
		return result, err
	}

	a.logger.Infof("ExecuteQueryWithLimit success - ConnectionID: %s, Rows: %d, Time: %dms",
		connectionID, result.Count, result.Time)

	return result, nil
}

// GetDatabases returns list of databases
func (a *App) GetDatabases(connectionID string) ([]string, error) {
	return a.dbManager.GetDatabases(connectionID)
}

// GetTables returns list of tables in a database
func (a *App) GetTables(connectionID string, database string) ([]database.TableInfo, error) {
	return a.dbManager.GetTables(connectionID, database)
}

// GetTableInfo returns detailed information about a table
func (a *App) GetTableInfo(connectionID string, database string, table string) (*database.TableInfo, error) {
	return a.dbManager.GetTableInfo(connectionID, database, table)
}

// GetTableData returns data from a table with pagination
func (a *App) GetTableData(connectionID string, database string, table string, limit int, offset int) (*database.QueryResult, error) {
	return a.dbManager.GetTableData(connectionID, database, table, limit, offset)
}

// GetDatabaseInfo returns general database information
func (a *App) GetDatabaseInfo(connectionID string) (*database.DatabaseInfo, error) {
	return a.dbManager.GetDatabaseInfo(connectionID)
}

// Utility Operations

// FormatQuery formats a query
func (a *App) FormatQuery(connectionID string, query string) string {
	return a.dbManager.FormatQuery(connectionID, query)
}

// ValidateQuery validates a query
func (a *App) ValidateQuery(connectionID string, query string) error {
	return a.dbManager.ValidateQuery(connectionID, query)
}

// GetSupportedDatabaseTypes returns list of supported database types
func (a *App) GetSupportedDatabaseTypes() []string {
	return []string{"mysql", "redis", "clickhouse"}
}

// Helper functions for map conversion
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}
