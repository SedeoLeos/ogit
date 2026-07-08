package main

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v60/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

const defaultEnvPath = "/opt/ogit/.env"
const installDir = "/opt/ogit"
const installBin = "/opt/ogit/ogit"
const pathBin = "/usr/local/bin/ogit"
const privateKeyFile = "/opt/ogit/private-key.pem"

type configInitOptions struct {
	privateFile string
}

func loadEnv() error {
	envPath := os.Getenv("OGIT_ENV_FILE")
	if envPath == "" {
		envPath = defaultEnvPath
	}

	if err := godotenv.Load(envPath); err == nil {
		return nil
	}

	if envPath != ".env" {
		if err := godotenv.Load(".env"); err == nil {
			return nil
		}
	}

	return fmt.Errorf("cannot load env file %s (or .env)", envPath)
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing env: " + key)
	}
	return value
}

func createJWT(appID string, privateKeyPath string) (string, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		return "", err
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"iat": now.Add(-60 * time.Second).Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": appID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func getGitHubToken() (string, error) {
	appID := getEnv("GITHUB_APP_ID")

	installationID, _ := strconv.ParseInt(
		getEnv("GITHUB_INSTALLATION_ID"),
		10,
		64,
	)

	privateKey := getEnv("GITHUB_PRIVATE_KEY_PATH")

	jwtToken, err := createJWT(appID, privateKey)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	client := github.NewClient(
		oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: jwtToken},
			),
		),
	)

	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return "", err
	}

	return token.GetToken(), nil
}

func authCloneURL(rawURL, token string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if parsed.Scheme != "https" {
		return rawURL, nil
	}

	parsed.User = url.UserPassword("x-access-token", token)
	return parsed.String(), nil
}

func runGit(args []string, token string) error {
	gitArgs := []string{}
	if token != "" {
		gitArgs = append(gitArgs, "-c", "http.extraHeader=Authorization: Bearer "+token)
	}

	gitArgs = append(gitArgs, args...)

	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func envFilePath() string {
	envPath := os.Getenv("OGIT_ENV_FILE")
	if envPath == "" {
		return defaultEnvPath
	}
	return envPath
}

func ensureInstallDir() error {
	return os.MkdirAll(installDir, 0755)
}

func writeEnvFile(path, appID, installationID string) error {
	content := fmt.Sprintf(
		"GITHUB_INSTALLATION_ID=%s\nGITHUB_APP_ID=%s\nGITHUB_PRIVATE_KEY_PATH=%s\n",
		installationID,
		appID,
		privateKeyFile,
	)
	return os.WriteFile(path, []byte(content), 0600)
}

func writePrivateKey(content string) error {
	return os.WriteFile(privateKeyFile, []byte(content), 0600)
}

func copySelfTo(path string) error {
	src, err := os.Executable()
	if err != nil {
		return err
	}

	src, err = filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.WriteFile(path, data, 0755)
}

func installBinary() error {
	if err := copySelfTo(installBin); err != nil {
		return err
	}
	return copySelfTo(pathBin)
}

func promptLine(reader *bufio.Reader, label string) (string, error) {
	fmt.Print(label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptPrivateKey(reader *bufio.Reader) (string, error) {
	fmt.Println("Paste the private key content, then type END on its own line:")
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		clean := strings.TrimRight(line, "\r\n")
		if clean == "END" {
			break
		}
		lines = append(lines, clean)
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func readPrivateKeyFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseConfigInitOptions(args []string) (configInitOptions, error) {
	opts := configInitOptions{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--private-file":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--private-file requires a path")
			}
			opts.privateFile = args[i+1]
			i++
		case strings.HasPrefix(arg, "--private-file="):
			opts.privateFile = strings.TrimPrefix(arg, "--private-file=")
		default:
			return opts, fmt.Errorf("unknown option: %s", arg)
		}
	}
	return opts, nil
}

func configInit(args []string) error {
	opts, err := parseConfigInitOptions(args)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	appID, err := promptLine(reader, "GITHUB_APP_ID: ")
	if err != nil {
		return err
	}

	installationID, err := promptLine(reader, "GITHUB_INSTALLATION_ID: ")
	if err != nil {
		return err
	}

	privateKeyContent := ""
	if opts.privateFile != "" {
		privateKeyContent, err = readPrivateKeyFromFile(opts.privateFile)
		if err != nil {
			return err
		}
	} else {
		privateKeyContent, err = promptPrivateKey(reader)
		if err != nil {
			return err
		}
	}

	if appID == "" || installationID == "" || strings.TrimSpace(privateKeyContent) == "" {
		return fmt.Errorf("all fields are required")
	}

	if err := ensureInstallDir(); err != nil {
		return err
	}

	if err := writePrivateKey(privateKeyContent); err != nil {
		return err
	}

	if err := writeEnvFile(envFilePath(), appID, installationID); err != nil {
		return err
	}

	if err := installBinary(); err != nil {
		return err
	}

	fmt.Println("configured ogit in /opt/ogit and installed /usr/local/bin/ogit")
	return nil
}

func runConfig(args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: ogit config init|show")
		return nil
	}

	switch args[0] {
	case "init":
		return configInit(args[1:])
	case "show":
		fmt.Println(envFilePath())
		return nil
	default:
		return fmt.Errorf("unknown config command: %s", args[0])
	}
}

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "config" {
		if err := runConfig(os.Args[2:]); err != nil {
			panic(err)
		}
		return
	}

	err := loadEnv()
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: ogit clone|pull|fetch|config")
		return
	}

	token, err := getGitHubToken()
	if err != nil {
		panic(err)
	}

	command := os.Args[1]

	switch command {
	case "clone":
		if len(os.Args) < 3 {
			fmt.Println("usage: ogit clone <repo-url> [directory]")
			return
		}

		authURL, err := authCloneURL(os.Args[2], token)
		if err != nil {
			panic(err)
		}

		cloneArgs := []string{"clone", authURL}
		if len(os.Args) > 3 {
			cloneArgs = append(cloneArgs, os.Args[3:]...)
		}

		err = runGit(cloneArgs, "")
	case "pull":
		pullArgs := append([]string{"pull"}, os.Args[2:]...)
		err = runGit(pullArgs, token)
	case "fetch":
		fetchArgs := append([]string{"fetch"}, os.Args[2:]...)
		err = runGit(fetchArgs, token)
	default:
		// laisse passer les autres commandes Git
		// ex: ogit status, ogit log...
		err = runGit(os.Args[1:], token)
	}

	if err != nil {
		if strings.Contains(err.Error(), "exit status") {
			return
		}
		panic(err)
	}
}
