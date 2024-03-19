FROM ubuntu:22.04

ARG QGT_VERSION=0.2.15

RUN mkdir -p /root/.kube/ && \
    echo "fd97de6b91a121428112c52e5fe04a15" > /etc/machine-id && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    pkg-config ca-certificates apt-transport-https gnupg gnupg2 lsb-release software-properties-common jq libc6-dev make curl wget vim dnsutils unzip libsodium-dev && \
    curl -sL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g npm@latest

RUN KUBECTL_VERSION=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt) && \
    curl -s "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" -o /usr/local/bin/kubectl && \
    chmod a+x /usr/local/bin/kubectl

RUN npm install -g quorum-genesis-tool@${QGT_VERSION}

RUN curl -s https://packages.microsoft.com/keys/microsoft.asc | apt-key --keyring /etc/apt/trusted.gpg.d/microsoft.asc.gpg add - && \
    AZ_REPO=$(lsb_release -cs) && \
    echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $AZ_REPO main" | tee /etc/apt/sources.list.d/azure-cli.list && \
    apt-get update && apt-get install -y azure-cli

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && ./aws/install && \
    rm -rf /var/lib/apt/lists/* aws awscliv2.zip
