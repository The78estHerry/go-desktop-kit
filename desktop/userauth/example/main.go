// Command example checks whether the OS can verify the user and, if so,
// shows the prompt once. It doubles as a manual test.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/The78estHerry/go-desktop-kit/desktop/userauth"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := userauth.Available(ctx); err != nil {
		fmt.Println("Not available:", err)
		os.Exit(1)
	}
	fmt.Println("Available. Showing the prompt...")
	err := userauth.Verify(ctx, userauth.Request{Message: "userauth example: confirm it is you"})
	switch {
	case err == nil:
		fmt.Println("Verified.")
	case errors.Is(err, userauth.ErrCanceled):
		fmt.Println("Canceled.")
	default:
		fmt.Println("Failed:", err)
		os.Exit(1)
	}
}
