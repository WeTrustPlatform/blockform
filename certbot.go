package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/WeTrustPlatform/blockform/model"
	"github.com/WeTrustPlatform/blockform/sshcmd"
	"goji.io/pat"
)

func setupHTTPS(ctx context.Context, node model.Node, domain string, email string, onSuccess func(), onError func(error)) {
	cmd := `sudo certbot certonly --nginx -d ` + domain + ` --agree-tos --non-interactive --email ` + email + ` && sudo sed -i -e "s|listen 8545;|server_name ` + domain + `;listen [::]:8545 ssl ipv6only=on;listen 8545 ssl;ssl_certificate /etc/letsencrypt/live/` + domain + `/fullchain.pem;ssl_certificate_key /etc/letsencrypt/live/` + domain + `/privkey.pem;include /etc/letsencrypt/options-ssl-nginx.conf;ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;|g" /etc/nginx/sites-available/default && sudo sed -i -e "s|listen 8546;|server_name ` + domain + `;listen [::]:8546 ssl ipv6only=on;listen 8546 ssl;ssl_certificate /etc/letsencrypt/live/` + domain + `/fullchain.pem;ssl_certificate_key /etc/letsencrypt/live/` + domain + `/privkey.pem;include /etc/letsencrypt/options-ssl-nginx.conf;ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;|g" /etc/nginx/sites-available/default && sudo systemctl restart nginx && sudo systemctl enable certbot.timer`

	_, _, err := sshcmd.Exec(
		os.Getenv("PRIV_KEY"),
		os.Getenv("PASSPHRASE"),
		"blockform",
		node.DomainName,
		cmd,
	)

	if err != nil {
		log.Println(err)
		onError(err)
		return
	}
	onSuccess()
}

func handleCertbot(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(500)
		return
	}

	ID := pat.Param(r, "id")
	domain := r.FormValue("domain")
	email := r.FormValue("email")

	// TODO validate form input

	node := model.Node{}
	db.Find(&node, ID)

	setupHTTPS(context.Background(), node, domain, email,
		func() {
			node.DomainName = domain
			node.HasSSL = true
			db.Save(&node)
			http.Redirect(w, r, "/node/"+ID+"/endpoints", http.StatusSeeOther)
		},
		func(err error) {
			w.Write([]byte(err.Error()))
		})
}
