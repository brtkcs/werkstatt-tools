**Import** – behúzzuk az `fmt`-t (kiíratás), `os`-t (fájlolvasás), és az első külső csomagot: `gopkg.in/yaml.v3` ami YAML fájlokat tud értelmezni.

**Két struct** – ez most újdonság. A `Host` egy sablon ami leírja hogy egy szerver hogyan néz ki: van neve, címe, portja. A `Config` pedig egy lista Host-okból. A backtick-es részek (`` `yaml:"name"` ``) megmondják a YAML csomagnak: "amikor a YAML-ban `name`-et látsz, azt a `Name` mezőbe töltsd".

**`func main()`** elindulunk:

`os.ReadFile("hosts.yaml")` – beolvassa az egész fájlt egyszerre egy `data` változóba. Ez különbözik az envcheck-től ahol soronként olvastunk – itt az egészet egyben kell mert a YAML parser úgy várja.

`yaml.Unmarshal(data, &config)` – ez a lényeg. Fogja a nyers YAML szöveget és "beletölti" a config struct-ba. Az `&` azt jelenti: "itt van a config memóriacíme, ide írd bele". Utána a `config.Hosts` már egy Go slice tele Host struct-okkal, amiken végig tudsz menni.

A `for` ciklus végigmegy a hostokon és kiírja mindegyiket: név → cím:port.

Tehát: fájl beolvasás → YAML-ból Go struct → kiíratás. Lényegében egy konfig fájlt tettünk "érthetővé" a Go számára.
