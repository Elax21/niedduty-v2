# Deploy nach Produktion (niedduty.de)

Wo alles liegt und wie ein neuer Stand live geht. Stand: 15.08.2026.

## Aufbau

| Was | Wo |
|---|---|
| Host | `v20.ovh.emserver.de` — **SSH-Port 22** (nicht 6041 wie sonst; `~/.ssh/config` setzt global 6041, deshalb `-p 22` mitgeben) |
| Container | Proxmox **CT 213** „niedduty-v2", intern `10.0.13.1` |
| Dienst | systemd `niedduty`, läuft als User `niedduty` auf Port **8213** |
| Binary | `/usr/local/bin/niedduty` (Frontend ist eingebettet) |
| Env | `/etc/niedduty/niedduty.env` (`EnvironmentFile` der Unit) |
| Datenbank | PostgreSQL **im selben Container**, DB `niedduty2` |
| Reverse-Proxy | CT 200 „v20-proxy" (`10.0.0.1`), nginx → `http://10.0.13.1:8213`, Let's-Encrypt für `niedduty.de` + `www` |
| Backup | `/root/bin/doBackup.sh` (Cron 3:00), siehe `deploy/doBackup.sh` |

**Im Container gibt es kein Go, Node oder Git.** Gebaut wird auf dem eigenen
Rechner, hochgeschoben wird nur das fertige Binary.

## Ablauf

```bash
# 0. Vorher: alles committet und gepusht, `go build ./...` und `npx vue-tsc -b` grün.

# 1. Frontend ins Binary bauen (schreibt nach internal/web/dist)
cd frontend && npm run build && cd ..

# 2. Binary für Linux bauen
export PATH=$PATH:~/.local/go/bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/niedduty ./cmd/server

# 3. Datenbank sichern (kostet Sekunden, spart Nerven)
ssh -p 22 v20.ovh.emserver.de 'pct exec 213 -- bash -lc \
  "su - postgres -c \"pg_dump niedduty2\" > /root/pre-deploy-$(date +%F-%H%M).sql"'

# 4. Binary auf den Host und in den Container
scp -P 22 /tmp/niedduty v20.ovh.emserver.de:/tmp/niedduty-new
ssh -p 22 v20.ovh.emserver.de 'pct push 213 /tmp/niedduty-new /usr/local/bin/niedduty.new --perms 755'

# 5. Tauschen und neu starten (alte Version bleibt als .bak liegen)
ssh -p 22 v20.ovh.emserver.de 'pct exec 213 -- bash -lc "
  cp -a /usr/local/bin/niedduty /usr/local/bin/niedduty.bak &&
  systemctl stop niedduty &&
  mv /usr/local/bin/niedduty.new /usr/local/bin/niedduty &&
  systemctl start niedduty && sleep 4 && systemctl is-active niedduty"'

# 6. Nachsehen
ssh -p 22 v20.ovh.emserver.de 'pct exec 213 -- journalctl -u niedduty -n 20 --no-pager'
curl -s -o /dev/null -w '%{http_code}\n' https://niedduty.de/
```

Schema-Änderungen brauchen keinen Extra-Schritt: `AutoMigrate` legt neue
Tabellen und Spalten beim Start an, einmalige Daten-Migrationen laufen aus
`internal/store/seed.go` mit (Merker in `settings`, z. B.
`migration.perUnitMinuten`). Im Log steht, was passiert ist.

## Zurückrollen

```bash
ssh -p 22 v20.ovh.emserver.de 'pct exec 213 -- bash -lc "
  systemctl stop niedduty &&
  mv /usr/local/bin/niedduty.bak /usr/local/bin/niedduty &&
  systemctl start niedduty"'
```

Datenbank zurück (nur wenn wirklich nötig, überschreibt den aktuellen Stand):

```bash
ssh -p 22 v20.ovh.emserver.de 'pct exec 213 -- bash -lc "
  systemctl stop niedduty &&
  su - postgres -c \"dropdb niedduty2 && createdb -O niedduty niedduty2\" &&
  su - postgres -c \"psql niedduty2\" < /root/pre-deploy-JJJJ-MM-TT-HHMM.sql &&
  systemctl start niedduty"'
```

## Nach dem Deploy

- `frontend/src/lib/changelog.ts` gepflegt? Dann sehen alle beim nächsten
  Öffnen einmalig die Notiz „Neu in der Kabine".
- `RELEASE.md` bekommt pro Änderung eine Zeile (✨ 🐛 💄 ⚙️).

## Stolpersteine

- **Port 22, nicht 6041.** Sonst „Connection refused".
- **`v20` allein löst nicht auf** — immer `v20.ovh.emserver.de`.
- Die interne IP `10.0.13.1` ist per SSH erreichbar, der Weg über
  `pct exec 213` vom Host ist aber der eingespielte.
- Binary statisch bauen (`CGO_ENABLED=0`), der Container hat eine andere
  libc-Version als Fedora.
