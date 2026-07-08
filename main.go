package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v60/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)


func getEnv(key string) string {

	value := os.Getenv(key)

	if value == "" {
		panic("missing env: " + key)
	}

	return value
}



func createJWT(
	appID string,
	privateKeyPath string,
) (string,error) {


	key,err := os.ReadFile(privateKeyPath)

	if err != nil {
		return "",err
	}


	privateKey,err := jwt.ParseRSAPrivateKeyFromPEM(key)

	if err != nil {
		return "",err
	}


	now := time.Now()


	claims := jwt.MapClaims{
		"iat": now.Add(-60*time.Second).Unix(),
		"exp": now.Add(10*time.Minute).Unix(),
		"iss": appID,
	}


	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		claims,
	)


	return token.SignedString(privateKey)
}



func getGitHubToken() (string,error) {


	appID := getEnv("GITHUB_APP_ID")

	installationID,_ := strconv.ParseInt(
		getEnv("GITHUB_INSTALLATION_ID"),
		10,
		64,
	)


	privateKey := getEnv(
		"GITHUB_PRIVATE_KEY_PATH",
	)



	jwtToken,err := createJWT(
		appID,
		privateKey,
	)


	if err != nil {
		return "",err
	}



	ctx:=context.Background()


	client:=github.NewClient(
		oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: jwtToken,
				},
			),
		),
	)



	token,_,err:=client.Apps.CreateInstallationToken(
		ctx,
		installationID,
		nil,
	)



	if err != nil {
		return "",err
	}



	return token.GetToken(),nil
}





func runGit(args []string, token string) error {


	gitArgs:=[]string{
		"-c",
		"http.extraHeader=Authorization: Bearer "+token,
	}


	gitArgs=append(
		gitArgs,
		args...,
	)



	cmd:=exec.Command(
		"git",
		gitArgs...,
	)


	cmd.Stdout=os.Stdout
	cmd.Stderr=os.Stderr
	cmd.Stdin=os.Stdin


	return cmd.Run()
}





func main(){


	err:=godotenv.Load()

	if err != nil {
		panic("cannot load .env")
	}



	if len(os.Args)<2 {

		fmt.Println(
			"usage: ogit clone|pull|fetch",
		)

		return
	}



	token,err:=getGitHubToken()

	if err != nil {
		panic(err)
	}



	command:=os.Args[1]


	switch command {



	case "clone":

		if len(os.Args)<3 {
			fmt.Println(
				"usage: ogit clone <repo-url>",
			)
			return
		}


		url:=os.Args[2]


		err=runGit(
			[]string{
				"clone",
				url,
			},
			token,
		)



	case "pull":

		err=runGit(
			[]string{
				"pull",
			},
			token,
		)



	case "fetch":

		err=runGit(
			[]string{
				"fetch",
			},
			token,
		)



	default:

		// laisse passer les autres commandes Git
		// ex: ogit status, ogit log...
		err=runGit(
			os.Args[1:],
			token,
		)
	}



	if err != nil {

		if strings.Contains(
			err.Error(),
			"exit status",
		){
			return
		}

		panic(err)
	}
}