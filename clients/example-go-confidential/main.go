package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var indexTemplate = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
<title>example-go-confidential</title>
<style>
body {
	font-family: monospace;
	color: #555;
	background: #e6edf4;
	padding: 1.25rem;
	margin: 0;
}
table {
	background: #fff;
	border: .0625rem solid #c4cdda;
	border-radius: 0 0 .25rem .25rem;
	border-spacing: 0;
	margin-bottom: 1.25rem;
	padding: .75rem 1.25rem;
	text-align: left;
	white-space: pre;
}
table > caption {
	background: #f1f6fb;
	text-align: left;
	font-weight: bold;
	padding: .75rem 1.25rem;
	border: .0625rem solid #c4cdda;
	border-radius: .25rem .25rem 0 0;
	border-bottom: 0;
}
table td, table th {
	padding: .25rem;
}
table > tbody > tr:hover {
	background: #f1f6fb;
}
</style>
</head>
<body>
	<table>
		<caption>Actions</caption>
		<tbody>
			<tr><td><a href="/auth/login">login</a></td></tr>
			<tr><td><a href="{{.WSO2ISMyAccountURL}}">My Account (WSO2IS)</a></td></tr>
		</tbody>
	</table>
</body>
</html>
`))

var authCallbackTemplate = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
<title>example-go-confidential</title>
<style>
body {
	font-family: monospace;
	color: #555;
	background: #e6edf4;
	padding: 1.25rem;
	margin: 0;
}
table {
	background: #fff;
	border: .0625rem solid #c4cdda;
	border-radius: 0 0 .25rem .25rem;
	border-spacing: 0;
	margin-bottom: 1.25rem;
	padding: .75rem 1.25rem;
	text-align: left;
	white-space: pre;
}
table > caption {
	background: #f1f6fb;
	text-align: left;
	font-weight: bold;
	padding: .75rem 1.25rem;
	border: .0625rem solid #c4cdda;
	border-radius: .25rem .25rem 0 0;
	border-bottom: 0;
}
table td, table th {
	padding: .25rem;
}
table > tbody > tr:hover {
	background: #f1f6fb;
}
</style>
</head>
<body>
	<table>
		<caption>Actions</caption>
		<tbody>
			<tr><td><a href="/">home</a></td></tr>
			<tr><td><a href="{{.WSO2ISMyAccountURL}}">My Account (WSO2IS)</a></td></tr>
		</tbody>
	</table>
	<table>
		<caption>User Claims</caption>
		<tbody>
			<tr><th>Issuer</th><td>{{.Issuer}}</td></tr>
			<tr><th>Subject</th><td>{{.Subject}}</td></tr>
			<tr><th>Username</th><td>{{.Username}}</td></tr>
			<tr><th>Email</th><td>{{.Email}}</td></tr>
			<tr><th>EmailVerified</th><td>{{.EmailVerified}}</td></tr>
			<tr><th>GivenName</th><td>{{.GivenName}}</td></tr>
			<tr><th>FamilyName</th><td>{{.FamilyName}}</td></tr>
		</tbody>
	</table>
</body>
</html>
`))

type indexTemplateData struct {
	WSO2ISMyAccountURL string
}

type authCallbackTemplateData struct {
	WSO2ISMyAccountURL string
	Issuer             string
	Subject            string
	Username           string
	Email              string
	EmailVerified      bool
	GivenName          string
	FamilyName         string
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCookie(w http.ResponseWriter, r *http.Request, path, name, value string) {
	c := &http.Cookie{
		Path:     path,
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

func deleteCookie(w http.ResponseWriter, r *http.Request, path, name string) {
	c := &http.Cookie{
		Path:     path,
		Name:     name,
		MaxAge:   -1,
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

// see https://github.com/coreos/go-oidc/blob/v3/example/idtoken/app.go
func main() {
	log.SetFlags(0)

	var listenAddress = flag.String("listen", ":8081", "Listen address.")
	var insecureSkipVerify = flag.Bool("tls-insecure-skip-verify", false, "Skip TLS verification")

	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		log.Fatalf("\nERROR You MUST NOT pass any positional arguments")
	}

	url := os.Getenv("EXAMPLE_URL")
	if url == "" {
		log.Fatalf("ERROR You MUST set the EXAMPLE_URL environment variable")
	}

	wso2isMyAccountURL := os.Getenv("EXAMPLE_WSO2IS_MY_ACCOUNT_URL")
	if wso2isMyAccountURL == "" {
		log.Fatalf("ERROR You MUST set the EXAMPLE_WSO2IS_MY_ACCOUNT_URL environment variable")
	}

	oidcIssuerURL := os.Getenv("EXAMPLE_OIDC_ISSUER_URL")
	if oidcIssuerURL == "" {
		log.Fatalf("ERROR You MUST set the EXAMPLE_OIDC_ISSUER_URL environment variable")
	}

	// read the client id and secret from file or environment variables.
	var oidcClientID string
	var oidcClientSecret string
	oidcClientConfigPath := os.Getenv("EXAMPLE_OIDC_CLIENT_CONFIG_PATH")
	if oidcClientConfigPath != "" {
		var config struct {
			Name         string `json:"name"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		data, err := ioutil.ReadFile(oidcClientConfigPath)
		if err != nil {
			log.Fatalf("ERROR Failed to read from %s: %v", oidcClientConfigPath, err)
		}
		err = json.Unmarshal(data, &config)
		if err != nil {
			log.Fatalf("ERROR Failed to unmarshal JSON from %s: %v", oidcClientConfigPath, err)
		}
		oidcClientID = config.ClientID
		oidcClientSecret = config.ClientSecret
	} else {
		oidcClientID = os.Getenv("EXAMPLE_OIDC_CLIENT_ID")
		oidcClientSecret = os.Getenv("EXAMPLE_OIDC_CLIENT_SECRET")
	}
	if oidcClientID == "" {
		log.Fatalf("ERROR You MUST set the EXAMPLE_OIDC_CLIENT_ID environment variable")
	}
	if oidcClientSecret == "" {
		log.Fatalf("ERROR You MUST set the EXAMPLE_OIDC_CLIENT_SECRET environment variable")
	}

	if *insecureSkipVerify {
		log.Printf("WARNING Disabling TLS verification")
		// NB we also fiddle with oidc.Config bellow.
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	oidcCallbackPath := "/auth/oidc/callback"

	var oidcProvider *oidc.Provider
	var err error

	for {
		oidcProvider, err = oidc.NewProvider(context.TODO(), oidcIssuerURL)
		if err != nil {
			log.Printf("WARNING Failed to initialize OIDC: %v. Retrying in a bit.", err)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}

	log.Printf("OIDC provider Authorization URL: %s", oidcProvider.Endpoint().AuthURL)

	oidcConfig := oauth2.Config{
		ClientID:     oidcClientID,
		ClientSecret: oidcClientSecret,
		RedirectURL:  url + oidcCallbackPath,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s%s\n", r.Method, r.Host, r.URL)
		state, err := randString(16)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		nonce, err := randString(16)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		// create the pkce code verifier.
		// see https://www.rfc-editor.org/rfc/rfc7636
		// see https://condatis.com/news/blog/oauth-confidential-clients/
		codeVerifier, err := randString(32)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		codeChallengeBytes := sha256.Sum256([]byte(codeVerifier))
		codeChallenge := base64.RawURLEncoding.EncodeToString(codeChallengeBytes[:])
		// save the oidc user authentication state as cookies.
		// TODO ciphertext the cookie values?
		setCookie(w, r, oidcCallbackPath, "state", state)
		setCookie(w, r, oidcCallbackPath, "nonce", nonce)
		setCookie(w, r, oidcCallbackPath, "code-verifier", codeVerifier)
		// start the oidc user authentication dance.
		// NB we are adding pkce code challenge because wso2is supports it.
		//    see the code_challenge_methods_supported property at, e.g.:
		// 		https://wso2is.test:9443/oauth2/token/.well-known/openid-configuration
		authCodeURL := oidcConfig.AuthCodeURL(
			state,
			oidc.Nonce(nonce),
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"))
		http.Redirect(w, r, authCodeURL, http.StatusFound)
	})

	http.HandleFunc(oidcCallbackPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s%s\n", r.Method, r.Host, r.URL)

		// verify the state.
		state, err := r.Cookie("state")
		if err != nil {
			http.Error(w, "state not found", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("state") != state.Value {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		// delete the state cookie.
		deleteCookie(w, r, oidcCallbackPath, state.Name)

		// get the code verifier.
		codeVerifier, err := r.Cookie("code-verifier")
		if err != nil {
			http.Error(w, "state not found", http.StatusBadRequest)
			return
		}

		// delete the code verifier cookie.
		deleteCookie(w, r, oidcCallbackPath, codeVerifier.Name)

		// exchange the authorization code with the access token and
		// identity token.
		token, err := oidcConfig.Exchange(
			context.TODO(),
			r.URL.Query().Get("code"),
			oauth2.SetAuthURLParam("code_verifier", codeVerifier.Value))
		if err != nil {
			http.Error(w, "Failed to exchange the authorization code with the access token: "+err.Error(), http.StatusBadRequest)
			return
		}

		unverifiedIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
			return
		}

		// NB in a real program, you should not log these tokens (they should
		// 	  be treated as secrets).
		log.Printf("ID Token: %v", unverifiedIDToken)
		log.Printf("Access Token: %v", token.AccessToken)

		// verify and get the verified id token.
		config := oidc.Config{
			ClientID: oidcClientID,
		}
		if *insecureSkipVerify {
			// TODO verify these.
			config.InsecureSkipSignatureCheck = true
		}
		verifier := oidcProvider.Verifier(&config)
		idToken, err := verifier.Verify(context.TODO(), unverifiedIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// verify the id token nonce.
		nonce, err := r.Cookie("nonce")
		if err != nil {
			http.Error(w, "nonce not found", http.StatusBadRequest)
			return
		}
		if idToken.Nonce != nonce.Value {
			http.Error(w, "nonce did not match", http.StatusBadRequest)
			return
		}

		// delete the nonce cookie.
		deleteCookie(w, r, oidcCallbackPath, nonce.Name)

		// extract the user claims from the id token.
		var claims struct {
			Issuer        string `json:"iss"`
			Subject       string `json:"sub"`
			Username      string `json:"username"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			GivenName     string `json:"given_name"`
			FamilyName    string `json:"family_name"`
		}
		err = idToken.Claims(&claims)
		if err != nil {
			http.Error(w, "Failed to get userinfo claims: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// show the user claims.
		w.Header().Set("Content-Type", "text/html")
		err = authCallbackTemplate.ExecuteTemplate(w, "", authCallbackTemplateData{
			WSO2ISMyAccountURL: wso2isMyAccountURL,
			Issuer:             claims.Issuer,
			Subject:            claims.Subject,
			Username:           claims.Username,
			GivenName:          claims.GivenName,
			FamilyName:         claims.FamilyName,
			Email:              claims.Email,
			EmailVerified:      claims.EmailVerified,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s%s\n", r.Method, r.Host, r.URL)

		if r.URL.Path != "/" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		err = indexTemplate.ExecuteTemplate(w, "", indexTemplateData{
			WSO2ISMyAccountURL: wso2isMyAccountURL,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Printf("Listening at http://%s\n", *listenAddress)

	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatalf("Failed to ListenAndServe: %v", err)
	}
}
