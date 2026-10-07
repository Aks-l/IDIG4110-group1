# SkyHiGh infrastructure

[`stack.yaml`](stack.yaml) is a Heat template that creates the whole cloud
environment for the prototype in the group's SkyHiGh project:

| Resource | Purpose |
| --- | --- |
| `twin-net`, `twin-subnet` (192.168.10.0/24), `twin-router` | Private network routed to `ntnu-internal` |
| `sg-edge` | SSH, HTTP, HTTPS and ICMP into `k3s-server` |
| `sg-cluster` | All traffic between the k3s nodes |
| `sg-data` | PostgreSQL (5432) and Kafka (9092) from the cluster, SSH from `k3s-server` |
| `k3s-server` | k3s control plane, Traefik ingress, SSH jump host; the only floating IP |
| `k3s-agent-1`, `k3s-agent-2` | k3s workers for the microservices |
| `data-vm` | Docker host for Kafka and PostgreSQL |
| `pg-data`, `kafka-data` | 20 GB SSD volumes mounted at `/srv/pgdata` and `/srv/kafka` on `data-vm` |

All four VMs use `gx3.2c8r`, which takes the project's full quota of 8 vCPUs
and 32 GB RAM.

## Prerequisites

- An application credential for the project. In Horizon, go to Identity →
  Application Credentials, then download `clouds.yaml` to
  `~/.config/openstack/clouds.yaml`. Never commit it.
- The OpenStack CLI: `pip install python-openstackclient python-heatclient`
- The NTNU network or VPN. Floating IPs on `ntnu-internal` are not reachable
  from outside NTNU.

## Usage

Creating or updating the stack needs a login that can create Keystone trusts.
Heat uses a trust to act on your behalf. A normal (restricted) application
credential cannot create trusts, and the CLI then fails with
`ERROR: Internal Error`. Create the stack in Horizon instead:

1. Go to **Project → Orchestration → Stacks → Launch Stack**.
2. Under **Template Source**, choose **File**, then select `infra/skyhigh/stack.yaml`.
3. Name the stack `twin` and leave **Rollback On Failure** unticked, so a
   failed resource stays visible.
4. Enter your password, set `key_name` to your keypair, and leave the other
   parameters at their defaults.

The CLI command below works only with password authentication (an OpenRC file)
or an *unrestricted* application credential:

```bash
openstack --os-cloud openstack stack create -t infra/skyhigh/stack.yaml --parameter key_name=<keypair> --wait twin
```

Reading the stack works with any credential. Show the floating IP, private IPs
and SSH commands:

```bash
openstack --os-cloud openstack stack output show twin --all
```

Apply changes to the template. This has the same login requirement as create,
or you can use **Change Stack Template** on the stack in Horizon:

```bash
openstack --os-cloud openstack stack update -t infra/skyhigh/stack.yaml --parameter key_name=<keypair> --wait twin
```

Delete everything, either with the command below or with **Delete Stacks** in
Horizon. This also deletes the `pg-data` and `kafka-data` volumes, so snapshot
them first if they hold anything you need:

```bash
openstack --os-cloud openstack stack delete twin
```

> **Warning:** Preview every update first with `openstack stack update --dry-run ...`.
> Heat replaces a VM whenever its boot script or image changes. For small
> changes such as a single security group rule, it is often safer to add the
> rule directly with `openstack security group rule create` and update the
> template to match. That is how the 5433–5435 rule was added on 2026-10-05.

## Data tier on data-vm

[`data-vm/compose.yaml`](data-vm/compose.yaml) runs Kafka (port 9092) and one
TimescaleDB container per service: `ingest_db` 5432, `twin_db` 5433,
`rules_db` 5434 and `identity_db` 5435. All ports are bound to the VM's private
IP and reachable only from the k3s cluster. Data lives on the Cinder volumes at
`/srv/kafka/data` and `/srv/pgdata/<service>`.

The live `.env` with generated passwords exists only on data-vm (mode 600).

## Deploying

[`deploy.sh`](deploy.sh) puts everything on top of the stack. Run it from Git
Bash or a Linux/macOS shell on the NTNU network or VPN. It reaches every VM
through k3s-server and is safe to re-run.

Deploy or update everything:

```bash
./infra/skyhigh/deploy.sh up
```

Deploy one part only:

```bash
./infra/skyhigh/deploy.sh up data
```

The parts you can name are:

- `access` installs the keys from [`authorized_keys`](authorized_keys) on
  k3s-server, gives k3s-server its own key on the other VMs, and installs
  `koble`.
- `data` deploys Kafka and the databases on data-vm and generates `.env` there
  the first time.
- `k8s` creates the `data-tier` Secret in Kubernetes from data-vm's `.env`, then
  applies every manifest under `infra/k8s/`.

Stop the services and data containers. Volumes and data are kept:

```bash
./infra/skyhigh/deploy.sh down
```

Show nodes, pods and data containers:

```bash
./infra/skyhigh/deploy.sh status
```

Workloads read database credentials and `DATA_VM_IP` from the `data-tier`
Secret, so passwords never pass through Git or a laptop.

If the stack is rebuilt, its IPs change. Set `FLOATING_IP`, `AGENT1_IP`,
`AGENT2_IP` and `DATA_IP` in the environment or at the top of the script, using
the values from `openstack stack output show twin --all`.

## Getting around

From `k3s-server`, `koble agent1`, `koble agent2` and `koble data` SSH onto the
other VMs, using k3s-server's own key. You can add a command after the name,
for example `koble data docker ps`.

## After the VMs boot

- Each VM's boot script logs to `/var/log/twin-bootstrap.log`.
- On `k3s-server`, `kubectl get nodes` should list all three nodes as `Ready`
  a few minutes after the stack completes.
- To give teammates access, add their public keys to `~/.ssh/authorized_keys`
  on the VMs.
