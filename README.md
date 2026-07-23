# Mon Carnet — journal technologie & cuisine (Go)

Un carnet de bord personnel : articles techniques (CI/CD, projets...) et
recettes de cuisine, avec photos, videos, et entrees publiques ou privees.

## Stack

- **Langage** : Go
- **Routeur HTTP** : chi
- **Base de donnees** : PostgreSQL, driver `lib/pq` (`database/sql`)
- **Acces DB** : sqlc — le SQL est ecrit a la main dans `db/queries.sql`, et
  `internal/db/` (genere, ne pas editer a la main) contient le code Go
  type-safe correspondant
- **Templates** : `html/template` de la stdlib (`templates/`)
- **Markdown → HTML** : goldmark
- **Sessions admin** : cookies signes (gorilla/sessions) + bcrypt
- **Conteneurisation** : Docker (build multi-etapes, image finale minimale
  basee sur `distroless`) — pensee pour se transposer vers Kubernetes

## Architecture

```
┌─────────────┐      ┌──────────────┐      ┌────────────────┐
│   Visiteur  │ ───▶ │  Go (chi +   │ ───▶ │  PostgreSQL     │
│  (toi, ta   │      │  html/tmpl)  │      │  (articles,     │
│  copine,    │◀──── │              │◀──── │  recettes)      │
│  internet)  │      │              │      │                 │
└─────────────┘      └──────┬───────┘      └────────────────┘
                             │
                     ┌───────▼────────┐
                     │ static/uploads  │  (photos / videos)
                     └────────────────┘
```

## Demarrage rapide (Docker Compose)

```bash
cp .env.example .env
# edite .env : mets un vrai SECRET_KEY, ton mot de passe admin, etc.

docker compose up -d --build
```

Le site est alors sur http://localhost:8000
L'admin est sur http://localhost:8000/admin (identifiants = ceux mis dans `.env`)

Deux comptes admin sont crees au demarrage si `ADMIN2_USERNAME` /
`ADMIN2_PASSWORD` sont remplis dans `.env` — un pour toi, un pour ta copine.

## Developpement local (sans Docker)

```bash
go mod download
export DATABASE_URL="postgres://journal:journal@localhost:5432/journal?sslmode=disable"
export SECRET_KEY="dev-secret"
go run ./cmd/server
```

Il te faut un PostgreSQL local qui tourne (ou lance juste `docker compose up -d db`
et pointe `DATABASE_URL` vers `localhost:5432`). Les migrations (`migrations/*.sql`)
sont appliquees automatiquement au demarrage (statements idempotents
`CREATE TABLE IF NOT EXISTS`).

## Modifier le schema de la base (workflow sqlc)

1. Modifie `migrations/0001_init.sql` (ou ajoute un `migrations/0002_xxx.sql`).
2. Modifie/ajoute les requetes dans `db/queries.sql`.
3. Regenere le code Go type-safe avec sqlc, via Docker (pas besoin d'installer
   sqlc sur ta machine) :

   ```bash
   docker run --rm -v "$(pwd):/src" -w /src sqlc/sqlc generate
   ```

4. Le code genere apparait dans `internal/db/` — ne le modifie jamais a la
   main, il sera ecrase au prochain `generate`.

## Structure du contenu

- **Categories** : `technologie` et `cuisine` sont creees automatiquement au
  premier lancement (voir `seed()` dans `cmd/server/main.go`). Pour en ajouter
  une troisieme, ajoute une ligne dans la liste `defaults` de cette fonction —
  aucune migration necessaire, la table est deja generique.
- **Entrees** : chaque entree appartient a une categorie et a un type
  (`article` ou `recipe`), et peut etre `published` (brouillon/publie) et
  `is_private` (visible de tous, ou uniquement apres connexion) independamment.
- **Medias** : chaque photo/video est enregistree dans `media`, rattachee a
  une entree. `cover_media_id` designe l'image de couverture.

## Categories

Depuis `/admin`, section "Categories", tu peux :
- **Creer** une categorie (nom + description ; le slug est genere automatiquement)
- **Renommer** / changer la description directement dans le tableau
- **Supprimer** une categorie — bloque tant que des entrees y sont encore
  rattachees, pour eviter de perdre le lien entre une entree et sa categorie

## Historique des versions (façon Git)

Chaque fois qu'une entree publiee est modifiee avec un changement reel de
contenu (titre, texte, categorie, ou details de recette), une nouvelle version
est archivee avec un message de justification (comme un message de commit).
Republier sans rien changer ne cree pas de version inutile.

- **Public** : `/entree/{slug}/historique` liste les versions passees (numero,
  message, date) ; `/entree/{slug}/historique/{n}` affiche le contenu complet
  d'une version donnee. Suit exactement la visibilite de l'entree elle-meme
  (une entree privee garde son historique prive).
- **Admin** : `/admin/entries/{id}/versions` liste l'historique avec un
  bouton "Restaurer" par version — restaurer copie cette ancienne version dans
  l'entree courante et cree une nouvelle version qui documente la
  restauration (rien n'est jamais perdu, comme un `git revert`).

## Le capteur ESP32

Le suivi du capteur (thermometre) n'est plus gere par ce site : tu le geres
sur ton propre site externe, et tu peux simplement mettre un lien vers celui-ci
dans un article de la categorie Technologie (un lien Markdown classique dans
le contenu de l'entree suffit).

## Portabilite & sauvegarde

**1. Le code** — mets ce dossier dans un depot Git des maintenant :

```bash
git init && git add . && git commit -m "Premier commit du carnet"
git remote add origin <ton-depot> && git push -u origin main
```

**2. Les donnees (base + medias)** :

```bash
./backups/backup.sh                                    # sauvegarde
./backups/restore.sh db-xxx.sql uploads-xxx.tar.gz      # restauration
```

Automatise `backup.sh` avec cron (exemple dans le script), et copie
regulierement `backups/` ailleurs que sur la machine qui heberge le site
(NAS, cloud, cle USB...). C'est cette copie "ailleurs" qui constitue la vraie
replication en cas de panne materielle.

> Note : les scripts utilisent le nom de volume `journal-go_uploads`, derive
> du nom du dossier du projet. Si tu renommes le dossier, adapte ce prefixe
> dans `backups/backup.sh` et `backups/restore.sh` (verifie avec
> `docker volume ls`).

**3. Reconstruction complete sur une nouvelle machine** :

```bash
git clone <ton-depot> && cd journal-go
cp .env.example .env   # remets les memes secrets
docker compose up -d
./backups/restore.sh <dump.sql> <uploads.tar.gz>
```

## Vers Kubernetes

L'image Docker (build multi-etapes, binaire Go statique dans une image
`distroless`) est deja pensee pour Kubernetes :

- `web` → un `Deployment` (scalable horizontalement, plusieurs pods derriere
  un `Service`)
- `db` → soit un `StatefulSet` + `PersistentVolumeClaim`, soit une base
  managee externe (plus simple a operer)
- `static/uploads` → un `PersistentVolumeClaim` (ou, si tu passes a plusieurs
  pods, un stockage S3-compatible partage type MinIO)
- Variables d'environnement → `ConfigMap` (valeurs non sensibles) + `Secret`
  (mot de passe DB, `SECRET_KEY`)
- Un `Ingress` pour le nom de domaine/HTTPS

## CI/CD

Le pipeline (`.github/workflows/ci-cd.yml`) a 3 etapes :

1. **build-and-test** (a chaque push et pull request) : `go build`, `go vet`,
   verification du formatage (`gofmt`).
2. **build-and-push** (uniquement sur push vers `main`) : construit l'image
   Docker et la pousse sur GitHub Container Registry (`ghcr.io`), taguee
   `latest` et avec le hash du commit.
3. **deploy** (uniquement sur push vers `main`) : se connecte en SSH a ton
   VPS et relance les conteneurs avec la nouvelle image.

### Mise en place, etape par etape

**1. Commande un VPS chez OVH**

- Va sur [ovhcloud.com](https://www.ovhcloud.com/fr/vps/), choisis l'offre la
  moins chere (VPS-1 ou equivalent, ~5€/mois).
- A la commande, choisis l'image **Ubuntu 26.04 LTS**, une region proche (Gravelines
  ou Strasbourg pour la France), et si l'option est proposee, **ajoute ta cle
  SSH publique** directement pendant la commande (plus simple que de la
  configurer apres).
- Une fois le VPS provisionne (quelques minutes a ~1h), tu recois l'adresse IP
  par email (et le mot de passe root si tu n'as pas mis de cle SSH a la
  commande).

**2. Connecte-toi et installe Docker**

```bash
ssh root@<IP_DU_VPS>
curl -fsSL https://get.docker.com | sh
```

**3. Securise un minimum le serveur (pare-feu)**

OVH ne fournit pas de pare-feu reseau integre comme certains autres
hebergeurs — configure `ufw` directement sur le serveur :

```bash
ufw allow OpenSSH
ufw allow 8000/tcp
ufw enable
```

**4. Cree le dossier de deploiement et copie les fichiers necessaires**

Sur le VPS :
```bash
mkdir ~/journal-go
```

Depuis ta machine, copie `docker-compose.prod.yml` et ton `.env` (rempli
avec de vrais secrets) vers le VPS :
```bash
scp docker-compose.prod.yml .env root@<IP_DU_VPS>:~/journal-go/
```

**5. Rends le package GHCR public** (evite de gerer des identifiants Docker
sur le serveur)

Sur GitHub : ton profil → **Packages** → le package cree apres le premier
push → **Package settings** → change la visibilite en **Public**.

**6. Cree une cle SSH dediee au deploiement** (si tu n'en as pas deja mis
une a la commande du VPS)

Sur ta machine :
```bash
ssh-keygen -t ed25519 -f deploy_key -N ""
ssh-copy-id -i deploy_key.pub root@<IP_DU_VPS>
```

**7. Ajoute 3 secrets dans ton depot GitHub**

Settings → Secrets and variables → Actions → New repository secret :

| Nom | Valeur |
|---|---|
| `VPS_HOST` | L'adresse IP de ton VPS OVH |
| `VPS_USER` | `root` (ou l'utilisateur SSH que tu utilises) |
| `VPS_SSH_KEY` | Le contenu complet de `deploy_key` (la cle **privee**) |

**8. Adapte `docker-compose.prod.yml`**

Remplace `<TON-COMPTE-GITHUB>/<TON-DEPOT>` par tes vraies valeurs, commite,
repousse.

**9. Pousse un changement et regarde le pipeline tourner**

Onglet **Actions** de ton depot : les 3 etapes doivent s'executer, la
derniere redemarrant le site sur ton VPS avec la nouvelle image. Le site est
alors accessible sur `http://<IP_DU_VPS>:8000`.

### Etapes suivantes (pas urgentes)

- **Nom de domaine + HTTPS** : ajoute un reverse proxy (Caddy, tres simple —
  gere le HTTPS automatiquement) devant le conteneur `web`, pointe ton
  domaine vers l'IP du VPS.
- **Sauvegardes** : adapte `backups/backup.sh` pour tourner directement sur
  le VPS (cron), et copie les sauvegardes ailleurs (ton PC, un stockage
  externe) — c'est ça qui garantit la vraie replication en cas de probleme
  serveur.
