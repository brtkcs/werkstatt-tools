Fentről lefelé:

**Konstansok és struct** – a színek ismerősek. A `Host` struct most IP-t és egy port listát tárol (nem bool mint előbb, hanem `[]int` – melyik portok nyitottak).

**`commonPorts`** – a portok amiket ellenőrzünk. Nem az összes 65535-öt, csak a gyakoriakat – SSH, DNS, web, adatbázis, stb.

**`scanHost` függvény** – kap egy IP-t, és végigscanneli rajta a commonPorts-ot. Portonként egy goroutine indul, mind egyszerre próbálkozik. Ha egy port nyitott, hozzáadja a `host.Ports` slice-hoz. A `Mutex` azért kell mert 11 goroutine egyszerre akar írni ugyanabba a listába – a Lock/Unlock biztosítja hogy egyszerre csak egy írhat. A végén rendezi a portokat szám szerint és visszaadja a Host-ot.

**`isAlive` függvény** – egyszerű kérdés: "válaszol-e ez az IP bármelyik porton?" Ha igen, él. Ha egyik porton sem, halott.

**`func main()`** két fázisban dolgozik:

**1. fázis – discovery:** 254 goroutine egyszerre scanneli a subnet összes IP-jét. Amelyik válaszol, bekerül az `aliveIPs` channel-be. Összegyűjtjük és rendezzük őket.

**2. fázis – port scan:** végigmegyünk az élő IP-ken, mindegyikre futtatjuk a `scanHost`-ot. Kiírjuk az IP-t és mellé a nyitott portokat lila színnel.

Tehát: először megkeresi ki van a hálózaton, aztán megnézi minek van nyitva az ajtaja. Két kör, egyre részletesebb.
