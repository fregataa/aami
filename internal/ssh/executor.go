package ssh

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Executor manages SSH connections and command execution
type Executor struct {
	config ExecutorConfig
}

// ExecutorConfig contains settings for the SSH executor
type ExecutorConfig struct {
	MaxParallel    int
	ConnectTimeout time.Duration
	CommandTimeout time.Duration
	MaxRetries     int
	BackoffBase    time.Duration
	BackoffMax     time.Duration
}

// Node represents a remote node for SSH connection
type Node struct {
	Name    string
	Host    string
	Port    int
	User    string
	KeyPath string
}

// Result represents the result of a command execution
type Result struct {
	Node     string
	Output   string
	Error    error
	Duration time.Duration
}

// NewExecutor creates a new SSH executor with the given configuration
func NewExecutor(cfg ExecutorConfig) *Executor {
	if cfg.MaxParallel == 0 {
		cfg.MaxParallel = 50
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 10 * time.Second
	}
	if cfg.CommandTimeout == 0 {
		cfg.CommandTimeout = 300 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.BackoffBase == 0 {
		cfg.BackoffBase = 2 * time.Second
	}
	if cfg.BackoffMax == 0 {
		cfg.BackoffMax = 30 * time.Second
	}
	return &Executor{config: cfg}
}

// NewExecutorFromConfig creates an executor from AAMI config values
func NewExecutorFromConfig(maxParallel, connectTimeout, commandTimeout, maxRetries, backoffBase, backoffMax int) *Executor {
	return NewExecutor(ExecutorConfig{
		MaxParallel:    maxParallel,
		ConnectTimeout: time.Duration(connectTimeout) * time.Second,
		CommandTimeout: time.Duration(commandTimeout) * time.Second,
		MaxRetries:     maxRetries,
		BackoffBase:    time.Duration(backoffBase) * time.Second,
		BackoffMax:     time.Duration(backoffMax) * time.Second,
	})
}

// Run executes a command on a single node
func (e *Executor) Run(ctx context.Context, node Node, command string) Result {
	start := time.Now()

	client, err := e.connect(node)
	if err != nil {
		return Result{Node: node.Name, Error: err, Duration: time.Since(start)}
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return Result{Node: node.Name, Error: fmt.Errorf("create session: %w", err), Duration: time.Since(start)}
	}
	defer session.Close()

	// Create a context with timeout for command execution
	cmdCtx, cancel := context.WithTimeout(ctx, e.config.CommandTimeout)
	defer cancel()

	// Run command with context
	done := make(chan struct{})
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = session.CombinedOutput(command)
		close(done)
	}()

	select {
	case <-cmdCtx.Done():
		return Result{
			Node:     node.Name,
			Error:    fmt.Errorf("command timed out after %v", e.config.CommandTimeout),
			Duration: time.Since(start),
		}
	case <-done:
		return Result{
			Node:     node.Name,
			Output:   string(output),
			Error:    cmdErr,
			Duration: time.Since(start),
		}
	}
}

// connect establishes an SSH connection to the node
func (e *Executor) connect(node Node) (*ssh.Client, error) {
	key, err := os.ReadFile(node.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("read key %s: %w", node.KeyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            node.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: implement proper host key verification
		Timeout:         e.config.ConnectTimeout,
	}

	port := node.Port
	if port == 0 {
		port = 22
	}

	addr := fmt.Sprintf("%s:%d", node.Host, port)
	return ssh.Dial("tcp", addr, config)
}

// TestConnection tests if a node is reachable via SSH
func (e *Executor) TestConnection(ctx context.Context, node Node) error {
	result := e.Run(ctx, node, "echo ok")
	return result.Error
}
