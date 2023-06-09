module github.com/m5lapp/go-user-service

go 1.20

require golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1

require github.com/lib/pq v1.10.9

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce // indirect
	golang.org/x/time v0.3.0 // indirect
)

require (
	github.com/kjk/betterguid v0.0.0-20170621091430-c442874ba63a
	github.com/m5lapp/go-service-toolkit v0.0.0-20230620000542-61a2a39348df
	golang.org/x/crypto v0.10.0
)

require (
	github.com/go-mail/mail/v2 v2.3.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)

replace github.com/m5lapp/go-service-toolkit => ../go-service-toolkit
