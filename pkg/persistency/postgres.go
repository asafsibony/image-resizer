package persistency

import (
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Client implements a wrapper around a DB connection
type Client struct {
	Database           *gorm.DB
	OnConnectionOpened func() error

	logger         *logrus.Logger
	connectionInfo *ConnectionInfo
	debug          bool
}

// ConnectionInfo holds persistency connection details
type ConnectionInfo struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Options  string
}

// NewClient creates and returns a persistency.Client
func NewClient(logger *logrus.Logger, connectionInfo *ConnectionInfo, debug bool) (*Client, error) {
	return &Client{
		Database:           nil,
		OnConnectionOpened: nil,
		logger:             logger,
		connectionInfo:     connectionInfo,
		debug:              debug,
	}, nil
}

// Start connects to the database
func (c *Client) Start() error {
	password := ""

	if c.connectionInfo.Password != "" {
		password = fmt.Sprintf("%s", url.QueryEscape(c.connectionInfo.Password))
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s %s",
		c.connectionInfo.Host,
		c.connectionInfo.User,
		password,
		c.connectionInfo.Database,
		c.connectionInfo.Port,
		c.connectionInfo.Options)

	c.logger.Info("Connecting to persistency", "dsn", dsn)

	var err error
	c.Database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return errors.Wrap(err, "Failed to connect to persistency")
	}

	db, err := c.Database.DB()
	if err != nil {
		return err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	db.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	db.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	db.SetConnMaxLifetime(time.Hour)

	if c.OnConnectionOpened != nil {
		c.logger.Info("Connected to persistency, invoking on connection opened handler")

		if err = c.OnConnectionOpened(); err != nil {
			db.Close()
			return errors.Wrap(err, "Error occurred in on connection opened handler")
		}

		c.logger.Info("On connection opened handler invoked successfully")
	} else {
		c.logger.Warn("No on connection opened handler registered")
	}

	c.logger.Info("Ready")

	return nil
}
