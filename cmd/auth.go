package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth [token]",
	Short: "Authenticate user",
	Long:  `Authenticate CLI through OpenAPI token. This is a mandatory command to perform any LLM call`,
	Example: `hange auth < token-file
cat token-file | hange auth
echo "token-value" | hange auth
hange auth "token-value"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var token string

		if len(args) == 1 {
			token = args[0]
		} else {
			var err error

			if token, err = readTokenFromStdin(); err != nil {
				return err
			}
		}

		if token == "" {
			return errors.New("failed to parse token argument")
		}

		if err := appFromCtx(cmd).Auth.SaveToken(token); err != nil {
			return err
		}

		return nil
	},
}

func readTokenFromStdin() (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("no token provided; pass it as argument or via stdin")
	}

	// Read first line (works for file, pipe, echo, etc.)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read token from stdin: %w", err)
	}

	return line, nil
}

func init() {
	rootCmd.AddCommand(authCmd)
}
