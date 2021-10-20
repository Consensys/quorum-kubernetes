FROM debian:buster-slim

RUN set -eux; \
    mkdir -p /root/.kube/ ; \
#https://github.com/balena-io-library/base-images/issues/273
    mkdir -p /usr/share/man/man1; \
    apt-get update && apt-get install -y --no-install-recommends \
    pkg-config ca-certificates apt-transport-https gnupg gnupg2 lsb-release software-properties-common jq libc6-dev make curl wget vim unzip openjdk-11-jre libsodium23=1.0.17-1; \
    echo "fd97de6b91a121428112c52e5fe04a15" > /etc/machine-id ; \
# kubectl    
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && chmod a+x ./kubectl && mv ./kubectl /usr/local/bin/kubectl ; \
# helm
    wget -q https://get.helm.sh/helm-v3.5.2-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm && chmod +x /usr/local/bin/helm ;

RUN set -eux; \
# hlf besu
    mkdir -p /opt/besu && wget -q https://hyperledger.jfrog.io/artifactory/besu-binaries/besu/21.7.4/besu-21.7.4.tar.gz -O - | tar xz -C /opt/besu --strip 1 ; \
# tessera
    mkdir -p /opt/tessera && wget -q https://oss.sonatype.org/service/local/repositories/releases/content/net/consensys/quorum/tessera/tessera-dist/21.10.0/tessera-dist-21.10.0.tar -O - | tar x -C /opt/tessera --strip 1 ;

RUN set -eux; \
# az cli
    curl -s https://packages.microsoft.com/keys/microsoft.asc | apt-key --keyring /etc/apt/trusted.gpg.d/microsoft.asc.gpg add - ; \
    echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ buster main" | tee /etc/apt/sources.list.d/azure-cli.list ; \
    apt-get update && apt-get install -y azure-cli ; \
# aws cli
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"; \
    unzip awscliv2.zip && ./aws/install; \
# cleanup
    rm -rf /var/lib/apt/lists/* ;

# istanbul-tools
ADD istanbul-tools_1.3.tar.gz /opt/istanbul-tools/
ENV PATH="/opt/istanbul-tools/bin:/opt/besu/bin:/opt/tessera/bin:${PATH}"
