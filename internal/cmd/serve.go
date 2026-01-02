package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/web"
	"github.com/steveyegge/gastown/internal/workspace"
)

var (
	servePort int
	serveHost string
	serveOpen bool
	serveDev  bool
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"web", "dashboard"},
	GroupID: GroupServices,
	Short:   "Start the web dashboard server",
	Long: `Start the Gas Town web dashboard server.

The dashboard provides a browser-based interface for monitoring:
- Town status and health
- Convoy progress (work tracking)
- Agent status (polecats, witnesses, refinery)
- Activity feed
- Mail/communication

By default, the server listens on localhost:8080.

Examples:
  gt serve                    # Start on localhost:8080
  gt serve --port 3000        # Custom port
  gt serve --host 0.0.0.0     # Bind to all interfaces
  gt serve --open             # Start and open browser
  gt serve --dev              # Development mode (enables CORS)`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "localhost", "Host to bind to")
	serveCmd.Flags().BoolVar(&serveOpen, "open", false, "Open browser automatically")
	serveCmd.Flags().BoolVar(&serveDev, "dev", false, "Development mode (enable CORS)")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Find town root
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	// Create server config
	cfg := web.Config{
		Host:     serveHost,
		Port:     servePort,
		DevMode:  serveDev,
		TownRoot: townRoot,
	}

	// Create and start server
	server := web.NewServer(cfg)

	// Open browser if requested
	if serveOpen {
		url := fmt.Sprintf("http://%s", server.Addr())
		go openBrowser(url)
	}

	// Start with graceful shutdown
	return server.StartWithGracefulShutdown()
}

// openBrowser opens the default browser to the given URL.
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}

	cmd.Start()
}
