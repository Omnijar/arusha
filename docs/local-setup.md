## Local setup

Arusha needs

1. Create a network for Arusha (only on first run).

docker network create arusha

2. Start Vault (port 8050).

```
docker run -it -d --cap-add=IPC_LOCK --network arusha --name vault -v ~/vault/data:/vault/file -v ~/vault/config:/vault/config vault server
```

**Note:** config.json should be under ~/vault/config.json and looks something like,

``` js
{
    "listener": {
        "tcp": {
            "address": "0.0.0.0:8050",
            "tls_disable": 1
        }
    },
    "storage": {
        "file": {
            "path": "/vault/file"
        }
    }
}
```

3. Initialize vault (only on first run):

```
docker run -it -d --rm --network arusha --name vault_client -e VAULT_ADDR=http://vault:8050 vault init
```

Keep a copy of the unseal keys and root token somewhere. Export `VAULT_TOKEN` using the root token.

4. Unseal vault using the unseal keys.

```
docker run -it --rm --network arusha --name vault_client -e VAULT_ADDR=http://vault:8050 vault unseal
```

4. Start Keto (port 4466).

```
docker run -it -d --network arusha --name keto -e DATABASE_URL=memory oryd/keto
```

5. Start hydra (port 4444).

```
docker run -it -d --network arusha --name hydra -e OAUTH2_ISSUER_URL=http://localhost/hydra -e OAUTH2_CONSENT_URL=http://localhost/consent -e OAUTH2_LOGIN_URL=${LOGIN_URL} -e DATABASE_URL=memory -e LOG_LEVEL=debug --entrypoint=hydra oryd/hydra:latest-alpine serve --dangerous-force-http
```

6. After exporting `VAULT_TOKEN`, `CALLBACK_URL`, `EMAIL_VERIFY_URL`, `SECRET_VERIFY_URL`, `MAILGUN_DOMAIN` and `MAILGUN_KEY`, launch Arusha (port 54932).

```
docker run -it --rm --name arusha --network arusha -e ARUSHA_CLIENT_CALLBACK_URL=${CALLBACK_URL} -e ARUSHA_EMAIL_VERIFY_URL=${EMAIL_VERIFY_URL} -e ARUSHA_TOKEN_VERIFY_URL=${SECRET_VERIFY_URL} -e ARUSHA_CLUSTER_URL=http://localhost -e HYDRA_PUBLIC_URL=http://localhost/hydra -e HYDRA_PRIVATE_URL=http://hydra:4444 -e KETO_CLUSTER_URL=http://keto:4466 -e VAULT_TOKEN=${VAULT_TOKEN} -e VAULT_ADDR=http://vault:8050 -e MAILGUN_DOMAIN=${MAILGUN_DOMAIN} -e MAILGUN_API_KEY=${MAILGUN_KEY} arusha
```

7. Now, launch the nginx proxy for gating Arusha and Hydra (port 80):

```
docker run -it -d -p 80:80 --network arusha --name nginx -v deploy/nginx/nginx.conf:/etc/nginx/nginx.conf:ro -v deploy/nginx/dev.conf:/etc/nginx/conf.d/dev.conf:ro nginx:alpine
```

8. The first thing Arusha needs is initializing its root client with a set of scopes. This can be done with:

curl -d '[{"method": "POST", "uri": "/some-url/:id", "name": "some-object.create"}]' http://localhost/scopes/init

Note that Arusha won't function properly until its initialized.

---

For resetting vault data, export `VAULT_TOKEN` and run:

```
echo users,emails,reset-tokens,email-tokens,credentials | tr ',' '\n' | while read thing; do docker run --rm --cap-add=IPC_LOCK --network arusha --name vault_client -e VAULT_ADDR=http://vault:8050 -e VAULT_TOKEN=${VAULT_TOKEN} vault sh -c "vault kv list secret/arusha/$thing | tail -n +3 | xargs -i vault kv delete secret/arusha/$thing/{}"; done
```

---

**Note:** For deployment, change occurrences of `localhost:$port` to your preference.
