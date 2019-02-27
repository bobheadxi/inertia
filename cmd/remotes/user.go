package remotescmd

import (
	"context"
	"fmt"
	"strings"
	"syscall"

	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/local"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
	"golang.org/x/crypto/ssh/terminal"
)

// UserCmd is the parent class for the 'user' subcommands
type UserCmd struct {
	*cobra.Command
	host *HostCmd
}

// AttachUserCmd attaches the 'user' subcommands to the given parent
func AttachUserCmd(host *HostCmd) {
	var user = &UserCmd{
		Command: &cobra.Command{
			Use:     "user",
			Short:   "Configure user access to Inertia Web",
			Long:    `Configure user access to the Inertia Web application.`,
			Aliases: []string{"u"},
		},
		host: host,
	}

	// attach children
	user.attachLoginCmd()
	AttachTotpCmd(user)
	user.attachAddCmd()
	user.attachRemoveCmd()
	user.attachListCmd()
	user.attachResetCmd()

	// attach to parent
	host.AddCommand(user.Command)
}

// context returns the root host command's context
func (root *UserCmd) context() context.Context { return root.host.ctx }

// getUserClient returns the root host command's user client
func (root *UserCmd) getUserClient() *client.UserClient { return root.host.client.GetUserClient() }

func (root *UserCmd) attachAddCmd() {
	const flagAdmin = "admin"
	var add = &cobra.Command{
		Use:   "add [user]",
		Short: "Create a user with access to this remote's Inertia daemon",
		Long: `Creates a user with access to this remote's Inertia daemon.

This user will be able to log in and view or configure the deployment
from the Inertia CLI (using 'inertia [remote] user login').

Use the --admin flag to create an admin user.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print("Enter a password for user: ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				output.Fatal("Invalid password")
			}
			var password = strings.TrimSpace(string(bytePassword))
			fmt.Print("\n")

			var admin, _ = cmd.Flags().GetBool(flagAdmin)
			if err := root.getUserClient().AddUser(root.context(), args[0], password, admin); err != nil {
				output.Fatal(err)
			}
		},
	}
	add.Flags().Bool(flagAdmin, false, "create a user with administrator permissions")
	root.AddCommand(add)
}

func (root *UserCmd) attachRemoveCmd() {
	var remove = &cobra.Command{
		Use:   "rm [user]",
		Short: "Remove a user",
		Long: `Removes the given user from Inertia's user database.

This user will no longer be able to log in and view or configure the deployment
remotely.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := root.getUserClient().RemoveUser(root.context(), args[0]); err != nil {
				output.Fatal(err)
			}
			println("user has been removed")
		},
	}
	root.AddCommand(remove)
}

func (root *UserCmd) attachLoginCmd() {
	var login = &cobra.Command{
		Use:   "login [user]",
		Short: "Authenticate with the remote",
		Long:  "Retreives an access token from the remote using your credentials.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var username = args[0]
			fmt.Print("Password: ")
			pwBytes, err := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				output.Fatal(err)
			}

			var totp, _ = cmd.Flags().GetString("totp")
			var req = client.AuthenticateRequest{
				User:     username,
				Password: string(pwBytes),
				TOTP:     totp,
			}
			token, err := root.getUserClient().Authenticate(root.context(), req)
			if err != nil && err != client.ErrNeedTotp {
				output.Fatal(err)
			}

			if err == client.ErrNeedTotp {
				// a TOTP is required
				fmt.Print("Authentication code (or backup code): ")
				totpBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if err != nil {
					output.Fatal(err)
				}
				req.TOTP = string(totpBytes)
				token, err = root.getUserClient().Authenticate(root.context(), req)
				if err != nil {
					output.Fatal(err)
				}
			}

			root.host.getRemote().Daemon.Token = token
			if err = local.SaveRemote(root.host.getRemote()); err != nil {
				output.Fatal(err)
			}

			fmt.Println("you have been logged in successfully, and a token has been saved")
		},
	}
	login.Flags().String("totp", "", "auth code or backup code for 2FA")
	root.AddCommand(login)
}

func (root *UserCmd) attachResetCmd() {
	var reset = &cobra.Command{
		Use:   "reset",
		Short: "Reset user database on your remote",
		Long: `Removes all users credentials on your remote. All configured user
will no longer be able to log in and view or configure the deployment
remotely.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := root.getUserClient().ResetUsers(root.context()); err != nil {
				output.Fatal(err)
			}
			println("all users removed")
		},
	}
	root.AddCommand(reset)
}

func (root *UserCmd) attachListCmd() {
	var list = &cobra.Command{
		Use:   "ls",
		Short: "List all users registered on your remote.",
		Long:  `Lists all users registered in Inertia's user database.`,
		Run: func(cmd *cobra.Command, args []string) {
			users, err := root.getUserClient().ListUsers(root.context())
			if err != nil {
				output.Fatal(err)
			}
			println(strings.Join(users, "\n"))
		},
	}
	root.AddCommand(list)
}