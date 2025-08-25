package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"rulestack/internal/client"
	"rulestack/internal/config"
)

// authCmd represents the auth command group
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long: `Authentication commands for user registration, login, and logout.
	
Manage your user account and authentication tokens for publishing packages.`,
}

// registerCmd handles user registration
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user account",
	Long: `Register a new user account with username, email, and password.
	
After successful registration, you'll be automatically logged in.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRegister()
	},
}

// loginCmd handles user login
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your user account",
	Long: `Login to your user account with username and password.
	
Your JWT token will be saved locally for future API calls.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogin()
	},
}

// logoutCmd handles user logout
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your user account",
	Long: `Logout from your user account and remove saved credentials.
	
This will invalidate your current session and remove the JWT token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogout()
	},
}

// whoamiCmd shows current user info
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user information",
	Long:  `Display information about the currently authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWhoami()
	},
}

func runRegister() error {
	// Get current registry
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Current == "" {
		return fmt.Errorf("no active registry configured. Use 'rfh registry add' to add one")
	}

	registry, exists := cfg.Registries[cfg.Current]
	if !exists {
		return fmt.Errorf("active registry '%s' not found", cfg.Current)
	}

	fmt.Printf("📝 Registering new account at %s\n\n", registry.URL)

	// Get user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)
	fmt.Println() // New line after password input

	fmt.Print("Confirm password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	confirm := string(confirmBytes)
	fmt.Println() // New line after password input

	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	// Validate input
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Create auth client and register
	authClient := client.NewAuthClient(registry.URL)
	authResp, err := authClient.Register(client.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	// Save user credentials to the current registry
	registryConfig := cfg.Registries[cfg.Current]
	registryConfig.Username = authResp.User.Username
	registryConfig.JWTToken = authResp.Token
	cfg.Registries[cfg.Current] = registryConfig

	// Also save to global user config for backward compatibility
	cfg.User = &config.User{
		Username: authResp.User.Username,
		Token:    authResp.Token,
	}

	if err := config.SaveCLI(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Successfully registered and logged in as %s\n", authResp.User.Username)
	fmt.Printf("👤 Role: %s\n", authResp.User.Role)
	fmt.Printf("🔑 Authentication token saved\n")

	return nil
}

func runLogin() error {
	// Get current registry
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Current == "" {
		return fmt.Errorf("no active registry configured. Use 'rfh registry add' to add one")
	}

	registry, exists := cfg.Registries[cfg.Current]
	if !exists {
		return fmt.Errorf("active registry '%s' not found", cfg.Current)
	}

	fmt.Printf("🔑 Logging in to %s\n\n", registry.URL)

	// Get user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)
	fmt.Println() // New line after password input

	// Validate input
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Create auth client and login
	authClient := client.NewAuthClient(registry.URL)
	authResp, err := authClient.Login(client.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save user credentials to the current registry
	registryConfig := cfg.Registries[cfg.Current]
	registryConfig.Username = authResp.User.Username
	registryConfig.JWTToken = authResp.Token
	cfg.Registries[cfg.Current] = registryConfig

	// Also save to global user config for backward compatibility
	cfg.User = &config.User{
		Username: authResp.User.Username,
		Token:    authResp.Token,
	}

	if err := config.SaveCLI(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Successfully logged in as %s\n", authResp.User.Username)
	fmt.Printf("👤 Role: %s\n", authResp.User.Role)
	fmt.Printf("🔑 Authentication token saved\n")

	return nil
}

func runLogout() error {
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var username string
	var tokenToLogout string

	// Check for per-registry authentication first
	if cfg.Current != "" {
		if registry, exists := cfg.Registries[cfg.Current]; exists && registry.Username != "" {
			username = registry.Username
			tokenToLogout = registry.JWTToken
			if tokenToLogout == "" {
				tokenToLogout = registry.Token // fallback to legacy token
			}
		}
	}

	// Fallback to global user authentication
	if username == "" && cfg.User != nil {
		username = cfg.User.Username
		tokenToLogout = cfg.User.Token
	}

	if username == "" {
		fmt.Println("ℹ️  You are not currently logged in")
		return nil
	}

	// Try to logout from server (invalidate session)
	if cfg.Current != "" && tokenToLogout != "" {
		if registry, exists := cfg.Registries[cfg.Current]; exists {
			authClient := client.NewAuthClient(registry.URL)
			if err := authClient.Logout(tokenToLogout); err != nil {
				// Don't fail if server logout fails - we'll clear local credentials anyway
				fmt.Printf("⚠️  Warning: Failed to logout from server: %v\n", err)
			}
		}
	}

	// Clear per-registry credentials
	if cfg.Current != "" {
		if registryConfig, exists := cfg.Registries[cfg.Current]; exists {
			registryConfig.Username = ""
			registryConfig.JWTToken = ""
			cfg.Registries[cfg.Current] = registryConfig
		}
	}

	// Clear global user credentials for backward compatibility
	cfg.User = nil

	if err := config.SaveCLI(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Successfully logged out %s\n", username)
	fmt.Printf("🗑️  Local credentials removed\n")

	return nil
}

func runWhoami() error {
	cfg, err := config.LoadCLI()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var username string
	var token string

	// Check for per-registry authentication first
	if cfg.Current != "" {
		if registry, exists := cfg.Registries[cfg.Current]; exists && registry.Username != "" {
			username = registry.Username
			token = registry.JWTToken
			if token == "" {
				token = registry.Token // fallback to legacy token
			}
		}
	}

	// Fallback to global user authentication
	if username == "" && cfg.User != nil {
		username = cfg.User.Username
		token = cfg.User.Token
	}

	if username == "" {
		fmt.Println("❌ You are not currently logged in")
		fmt.Println("Use 'rfh auth login' to authenticate or 'rfh auth register' to create an account")
		return nil
	}

	fmt.Printf("👤 Logged in as: %s\n", username)

	// Try to get detailed profile from server
	if cfg.Current != "" && token != "" {
		if registry, exists := cfg.Registries[cfg.Current]; exists {
			authClient := client.NewAuthClient(registry.URL)
			if profile, err := authClient.GetProfile(token); err == nil {
				fmt.Printf("📧 Email: %s\n", profile.Email)
				fmt.Printf("🎭 Role: %s\n", profile.Role)
				if profile.LastLogin != nil {
					fmt.Printf("🕐 Last login: %s\n", profile.LastLogin.Format("2006-01-02 15:04:05"))
				}
				fmt.Printf("📅 Account created: %s\n", profile.CreatedAt.Format("2006-01-02"))
			} else {
				fmt.Printf("⚠️  Could not fetch profile details: %v\n", err)
			}
		}
	}

	fmt.Printf("🔑 Token: [saved]\n")

	return nil
}

func init() {
	authCmd.AddCommand(registerCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(whoamiCmd)
}