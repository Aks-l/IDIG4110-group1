#!/usr/bin/env bash
# Deploy the twin platform onto the SkyHiGh stack (see README.md).
#
# Runs from a laptop on the NTNU network or VPN and reaches every VM through
# k3s-server. Every step is safe to re-run. It never touches the Heat stack or
# the VMs themselves, only what runs on them, and `down` never deletes data.
#
#   ./infra/skyhigh/deploy.sh up     [all|access|data|k8s]
#   ./infra/skyhigh/deploy.sh down   [all|data|k8s]
#   ./infra/skyhigh/deploy.sh status
#
# Addresses default to the current stack outputs; override with environment
# variables if the stack is rebuilt (openstack stack output show twin --all).
set -euo pipefail

FLOATING_IP=${FLOATING_IP:-10.212.168.194}
AGENT1_IP=${AGENT1_IP:-192.168.10.201}
AGENT2_IP=${AGENT2_IP:-192.168.10.187}
DATA_IP=${DATA_IP:-192.168.10.173}

HERE=$(cd "$(dirname "$0")" && pwd)
K8S_DIR="$HERE/../k8s"

SSH_OPTS=(-o BatchMode=yes -o ConnectTimeout=15 -o StrictHostKeyChecking=accept-new)

server() { ssh "${SSH_OPTS[@]}" "ubuntu@$FLOATING_IP" "$@"; }
vm()     { local ip=$1; shift; ssh "${SSH_OPTS[@]}" -J "ubuntu@$FLOATING_IP" "ubuntu@$ip" "$@"; }
to_data() { scp -q "${SSH_OPTS[@]}" -J "ubuntu@$FLOATING_IP" "$@" "ubuntu@$DATA_IP:data-tier/"; }
kubectl_remote() { server sudo k3s kubectl "$@"; }

step() { printf '\n==> %s\n' "$*"; }

# Reads "type key comment" lines on stdin and appends the missing ones.
ADD_KEYS='touch ~/.ssh/authorized_keys
while read -r t b c; do
  case "$t" in ""|\#*) continue ;; esac
  grep -qF "$b" ~/.ssh/authorized_keys || { echo "$t $b $c" >> ~/.ssh/authorized_keys; echo "  added $c on $(hostname)"; }
done'

# --- access: SSH keys and koble ---------------------------------------------

up_access() {
  step "Group keys on k3s-server"
  server "$ADD_KEYS" < "$HERE/authorized_keys"

  step "k3s-server key on the other VMs"
  local key
  key=$(server '[ -f ~/.ssh/id_ed25519 ] || ssh-keygen -q -t ed25519 -N "" -C k3s-server -f ~/.ssh/id_ed25519; cat ~/.ssh/id_ed25519.pub')
  for ip in "$AGENT1_IP" "$AGENT2_IP" "$DATA_IP"; do
    echo "$key" | vm "$ip" "$ADD_KEYS"
  done

  step "koble on k3s-server"
  server 'sudo tee /usr/local/bin/koble >/dev/null && sudo chmod 755 /usr/local/bin/koble' <<EOF
#!/bin/bash
# koble <machine> [command]: SSH from k3s-server to another VM in the twin stack.
# Installed by infra/skyhigh/deploy.sh.
case "\$1" in
  agent1|agent-1|k3s-agent-1) ip=$AGENT1_IP ;;
  agent2|agent-2|k3s-agent-2) ip=$AGENT2_IP ;;
  data|data-vm)               ip=$DATA_IP ;;
  *) echo "usage: koble {agent1|agent2|data} [command]"; exit 1 ;;
esac
shift
exec ssh -o StrictHostKeyChecking=accept-new "ubuntu@\$ip" "\$@"
EOF
  echo "  koble agent1 | koble agent2 | koble data"
}

# --- data: Kafka and databases on data-vm ----------------------------------

up_data() {
  step "Data tier on data-vm"
  vm "$DATA_IP" 'mkdir -p ~/data-tier'
  to_data "$HERE/data-vm/compose.yaml" "$HERE/data-vm/.env.example"
  vm "$DATA_IP" bash -s -- "$DATA_IP" <<'EOF'
set -euo pipefail
cd ~/data-tier
if [ ! -f .env ]; then
  echo "  generating .env (secrets stay on this VM)"
  cp .env.example .env
  sed -i "s/^DATA_VM_IP=.*/DATA_VM_IP=$1/" .env
  cid=$(sudo docker run --rm apache/kafka:4.3.1 /opt/kafka/bin/kafka-storage.sh random-uuid 2>/dev/null | tail -1)
  sed -i "s/^KAFKA_CLUSTER_ID=.*/KAFKA_CLUSTER_ID=$cid/" .env
  for v in INGEST TWIN RULES IDENTITY; do
    sed -i "s/^${v}_DB_PASSWORD=.*/${v}_DB_PASSWORD=$(openssl rand -hex 24)/" .env
  done
  chmod 600 .env
fi
sudo mkdir -p /srv/kafka/data /srv/pgdata/{ingest,twin,rules,identity}
sudo chown 1000:1000 /srv/kafka/data
sudo docker compose up -d --remove-orphans
echo "  waiting for databases to report healthy"
for _ in $(seq 1 30); do
  [ "$(sudo docker compose ps --format '{{.Health}}' | grep -cx healthy)" -ge 4 ] && break
  sleep 3
done
sudo docker compose ps --format 'table {{.Name}}\t{{.Status}}'
EOF
}

down_data() {
  step "Stopping data tier (data on /srv is kept)"
  vm "$DATA_IP" 'cd ~/data-tier 2>/dev/null && sudo docker compose down || echo "  nothing deployed"'
}

# --- k8s: secrets and manifests ----------------------------------------------

manifests() {
  [ -d "$K8S_DIR" ] || return 0
  find "$K8S_DIR" -type f \( -name '*.yaml' -o -name '*.yml' \) | sort
}

up_k8s() {
  step "Secret 'data-tier' in Kubernetes (copied from data-vm, never stored locally)"
  vm "$DATA_IP" 'grep -v "^#" ~/data-tier/.env | grep "="' \
    | server 'sudo k3s kubectl create secret generic data-tier --from-env-file=/dev/stdin --dry-run=client -o yaml | sudo k3s kubectl apply -f -'

  local files
  files=$(manifests)
  if [ -z "$files" ]; then
    step "No manifests in infra/k8s/ yet, skipping"
    return
  fi
  step "Applying manifests from infra/k8s/"
  for f in $files; do cat "$f"; echo '---'; done | kubectl_remote apply -f -
}

down_k8s() {
  local files
  files=$(manifests)
  if [ -z "$files" ]; then
    step "No manifests in infra/k8s/, nothing to remove"
    return
  fi
  step "Removing workloads defined in infra/k8s/"
  for f in $files; do cat "$f"; echo '---'; done | kubectl_remote delete --ignore-not-found -f -
}

# --- status --------------------------------------------------------------------

status() {
  step "k3s nodes"
  kubectl_remote get nodes
  step "Pods"
  kubectl_remote get pods -A -o wide
  step "Data tier"
  vm "$DATA_IP" 'cd ~/data-tier 2>/dev/null && sudo docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" || echo "  not deployed"'
}

# --- main ----------------------------------------------------------------------

cmd=${1:-}
target=${2:-all}

case "$cmd:$target" in
  up:all)    up_access; up_data; up_k8s ;;
  up:access) up_access ;;
  up:data)   up_data ;;
  up:k8s)    up_k8s ;;
  down:all)  down_k8s; down_data ;;
  down:data) down_data ;;
  down:k8s)  down_k8s ;;
  status:*)  status ;;
  *)
    sed -n '2,13p' "$0" | sed 's/^# \{0,1\}//'
    exit 1
    ;;
esac
