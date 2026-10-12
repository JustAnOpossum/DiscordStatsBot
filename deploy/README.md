# Production deploy (NetBird + GHCR)

Pipeline: push to `master` → `docker-image.yml` builds and pushes the image
to GHCR → `deploy.yml` fires (`workflow_run`), joins the NetBird network as
an ephemeral peer, SSHes to the server, copies `docker-compose.prod.yml`,
and runs `docker compose pull && up -d`.

Nothing sensitive is in this repo: the server host, SSH key, and bot
secrets all live in the GitHub **production** environment and on the
server itself.

## One-time GitHub setup

1. **Environment**: Settings → Environments → *Create environment* →
   name it exactly `production`. Enable **required reviewers** (yourself)
   so every deploy waits for your approval — the key protection for a
   public repo.
2. **Environment secrets** (on the `production` environment, not repo-level):

   | Secret | Value |
   |---|---|
   | `NETBIRD_SETUP_KEY` | NetBird setup key — **Ephemeral ON**, own group, access policy allowing only SSH to the bot server |
   | `DEPLOY_HOST` | Server's NetBird IP (`100.x.y.z`) or `.netbird.cloud` name |
   | `DEPLOY_USER` | SSH user on the server (dedicated `deploy` user, not root) |
   | `DEPLOY_DIR` | e.g. `/opt/discordstatsbot` |
   | `DEPLOY_SSH_KEY` | Private half of a dedicated ed25519 deploy key |
   | `DEPLOY_KNOWN_HOSTS` | Optional: `ssh-keyscan <netbird-ip>` output to pin the host key |

## One-time server setup

```bash
# 1. NetBird peer, tagged/grouped so only the deploy key's group can SSH it.

# 2. Dedicated deploy user with the public half of the deploy key:
sudo useradd -m -s /bin/bash deploy
sudo install -d -m 700 -o deploy /home/deploy/.ssh
echo '<public key>' | sudo tee /home/deploy/.ssh/authorized_keys
sudo chmod 600 /home/deploy/.ssh/authorized_keys

# 3. App dir + bot secrets (never committed):
sudo mkdir -p /opt/discordstatsbot
sudo chown deploy /opt/discordstatsbot
# create /opt/discordstatsbot/.env.prod from .env.example (BOT_TOKEN, BING_KEY, ...)

# 4. GHCR pull auth — pick ONE:
#    a) Set the GHCR package to PUBLIC (anonymous pulls, zero server creds), or
#    b) fine-grained PAT with read:packages on this repo only, then:
docker login ghcr.io -u <github-user>   # store the PAT
```

## Notes

- First deploy uses `StrictHostKeyChecking=accept-new` (TOFU). Pin the
  host key afterwards with `DEPLOY_KNOWN_HOSTS` for strict checking.
- The workflow copies the compose file on every run, so compose changes
  deploy with the code — no manual server edits needed.
- `workflow_run` jobs never run for fork PRs; secrets are only reachable
  from pushes to `master` by people with write access.
- The root `docker-compose.yml` (with `build:`) stays for local dev;
  production uses this `deploy/docker-compose.prod.yml` (image pull only).
