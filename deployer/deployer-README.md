# deployer

Compose stack kezelő CLI tool – egy helyről kezeled az összes stack-odat.

## Használat

```bash
deployer status              # összes stack állapota
deployer status media        # csak a media stack
deployer up                  # összes stack indítása
deployer up media            # csak a media stack indítása
deployer down                # összes stack leállítása
deployer down monitoring     # csak a monitoring leállítása
```

## Konfiguráció

`deployer.yaml`:

```yaml
# runtime: docker | podman
runtime: docker
stacks:
  - name: media
    path: /srv/stacks/media
  - name: monitoring
    path: /srv/stacks/monitoring
  - name: tools
    path: /srv/stacks/tools
```

## Telepítés

```bash
go build -o deployer
mv deployer ~/.local/bin/
```

## Hogyan működik

1. Beolvassa a `deployer.yaml`-t – ebből tudja melyik stack hol van és milyen runtime-ot használ
2. A CLI argumentumból kapja az akciót (`status`, `up`, `down`) és opcionálisan a stack nevét
3. Az adott runtime-mal (`docker`/`podman`) futtatja a `compose` parancsokat a stack mappájából

## Megtanult Go koncepciók

- `os/exec` – shell parancs futtatás Go-ból
- `cmd.Dir` – parancs futtatása adott mappából
- `cmd.Stdout = os.Stdout` – kimenet átfolyatása a terminálba
- `os.Args` – CLI argumentumok kezelése
- YAML konfig beolvasás struct-ba
- Slice szűrés (target stack kiválasztása)
