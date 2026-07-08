package main

import (
	"bufio"
	"context"
	"errors"
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
const pathLink = "/usr/local/bin/ogit"

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

func writeEnvFile(path, appID, installationID, privateKeyPath string) error {
	content := fmt.Sprintf(
		"GITHUB_INSTALLATION_ID=%s\nGITHUB_APP_ID=%s\nGITHUB_PRIVATE_KEY_PATH=%s\n",
		installationID,
		appID,
		privateKeyPath,
	)
	return os.WriteFile(path, []byte(content), 0600)
}

func copySelfToInstallDir() error {
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

	return os.WriteFile(installBin, data, 0755)
}

func updatePathLink() error {
	if err := os.Remove(pathLink); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Symlink(installBin, pathLink)
}

func configInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("GITHUB_APP_ID: ")
	appID, _ := reader.ReadString('\n')
	fmt.Print("GITHUB_INSTALLATION_ID: ")
	installationID, _ := reader.ReadString('\n')
	fmt.Print("GITHUB_PRIVATE_KEY_PATH: ")
	privateKeyPath, _ := reader.ReadString('\n')

	appID = strings.TrimSpace(appID)
	installationID = strings.TrimSpace(installationID)
	privateKeyPath = strings.TrimSpace(privateKeyPath)

	if appID == "" || installationID == "" || privateKeyPath == "" {
		return fmt.Errorf("all fields are required")
	}

	if err := ensureInstallDir(); err != nil {
		return err
	}

	if err := writeEnvFile(envFilePath(), appID, installationID, privateKeyPath); err != nil {
		return err
	}

	if err := copySelfToInstallDir(); err != nil {
		return err
	}

	if err := updatePathLink(); err != nil {
		return err
	}

	fmt.Println("configured ogit in /opt/ogit and linked /usr/local/bin/ogit")
	return nil
}

func runConfig(args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: ogit config init|show")
		return nil
	}

	switch args[0] {
	case "init":
		return configInit()
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
