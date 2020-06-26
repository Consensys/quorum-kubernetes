
FROM debian:buster-slim

RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends \
        # utilities for keeping Debian CA certificates in sync
		ca-certificates p11-kit \
		# general utilities
		curl wget gettext-base \
		;

RUN set -eux; \
	apt-get install -y --no-install-recommends \
        apt-transport-https gnupg gnupg2 \
        ; \
		# kubectl
		curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl \
		; \
		chmod +x ./kubectl \
		; \
        mv ./kubectl /usr/local/bin/kubectl \
        ; \
        rm -rf /var/lib/apt/lists/*