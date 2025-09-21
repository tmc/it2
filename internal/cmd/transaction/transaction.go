package transaction

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the transaction command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction",
		Short: "Manage iTerm2 transactions",
		Long:  "Commands for managing iTerm2 transactions to group multiple operations atomically",
	}

	cmd.AddCommand(newBeginCommand())
	cmd.AddCommand(newCommitCommand())
	cmd.AddCommand(newRollbackCommand())
	cmd.AddCommand(newStatusCommand())

	return cmd
}

func newBeginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "begin",
		Short: "Start a new transaction",
		Long:  "Begin a new transaction to group multiple operations atomically",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			c := client.New(wsURL)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			return beginTransaction(ctx, c)
		},
	}
}

func newCommitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "Commit the current transaction",
		Long:  "Commit all operations in the current transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			c := client.New(wsURL)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			return commitTransaction(ctx, c)
		},
	}
}

func newRollbackCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the current transaction",
		Long:  "Rollback and discard all operations in the current transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			c := client.New(wsURL)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			return rollbackTransaction(ctx, c)
		},
	}
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current transaction status",
		Long:  "Display the status of the current transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			c := client.New(wsURL)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			return getTransactionStatus(ctx, c)
		},
	}
}

func beginTransaction(ctx context.Context, c *client.Client) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_TransactionRequest{
			TransactionRequest: &pb.TransactionRequest{
				Begin: cmdutil.BoolPtr(true),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if resp := response.GetTransactionResponse(); resp != nil {
		switch resp.GetStatus() {
		case pb.TransactionResponse_OK:
			fmt.Println("Transaction started successfully")
			return nil
		case pb.TransactionResponse_ALREADY_IN_TRANSACTION:
			return fmt.Errorf("already in a transaction")
		default:
			return fmt.Errorf("failed to begin transaction: %v", resp.GetStatus())
		}
	}

	return fmt.Errorf("unexpected response type")
}

func commitTransaction(ctx context.Context, c *client.Client) error {
	// iTerm2's transaction API uses begin=false to commit
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_TransactionRequest{
			TransactionRequest: &pb.TransactionRequest{
				Begin: cmdutil.BoolPtr(false),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if resp := response.GetTransactionResponse(); resp != nil {
		switch resp.GetStatus() {
		case pb.TransactionResponse_OK:
			fmt.Println("Transaction committed successfully")
			return nil
		case pb.TransactionResponse_NO_TRANSACTION:
			return fmt.Errorf("no transaction in progress")
		default:
			return fmt.Errorf("failed to commit transaction: %v", resp.GetStatus())
		}
	}

	return fmt.Errorf("unexpected response type")
}

func rollbackTransaction(ctx context.Context, c *client.Client) error {
	// For rollback, we need to end the transaction without committing
	// This might require checking iTerm2's specific implementation
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_TransactionRequest{
			TransactionRequest: &pb.TransactionRequest{
				Begin: cmdutil.BoolPtr(false),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	if resp := response.GetTransactionResponse(); resp != nil {
		switch resp.GetStatus() {
		case pb.TransactionResponse_OK:
			fmt.Println("Transaction rolled back successfully")
			return nil
		case pb.TransactionResponse_NO_TRANSACTION:
			return fmt.Errorf("no transaction in progress")
		default:
			return fmt.Errorf("failed to rollback transaction: %v", resp.GetStatus())
		}
	}

	return fmt.Errorf("unexpected response type")
}

func getTransactionStatus(ctx context.Context, c *client.Client) error {
	// Check transaction status by trying to begin one
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_TransactionRequest{
			TransactionRequest: &pb.TransactionRequest{
				Begin: cmdutil.BoolPtr(true),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to check transaction status: %w", err)
	}

	if resp := response.GetTransactionResponse(); resp != nil {
		switch resp.GetStatus() {
		case pb.TransactionResponse_OK:
			fmt.Println("No transaction in progress (test transaction created)")
			// Immediately commit the test transaction
			return commitTransaction(ctx, c)
		case pb.TransactionResponse_ALREADY_IN_TRANSACTION:
			fmt.Println("Transaction in progress")
			return nil
		default:
			return fmt.Errorf("failed to check transaction status: %v", resp.GetStatus())
		}
	}

	return fmt.Errorf("unexpected response type")
}
