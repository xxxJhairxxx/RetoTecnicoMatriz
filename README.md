# Reto técnico — Factorización QR (Go) + Estadísticas (Node.js)

Dos APIs REST que se comunican por HTTP:

```
Navegador ──► Go/Fiber :8080 ──► Node/Express :3000
             (QR + JWT)          (estadísticas)
```

La API en Go recibe una matriz, calcula su factorización QR y pide a la API en Node las
estadísticas de las matrices Q y R. Solo la API en Go está expuesta; la de Node vive en la red
interna de Docker.

## Ejecutar

```bash
docker compose up --build
```

Frontend en <http://localhost:8080>. El `.env` incluido trae credenciales de demo
(`demo` / `Qr-Demo-2026`).

## Pruebas

```bash
cd node-api && npm install && npm test   # 18 pruebas
cd go-api   && go test ./...             # 23 pruebas
```

Sin Go instalado, el `Dockerfile` ya las ejecuta antes de compilar: si `docker compose build`
pasa, la suite está verde.

## Endpoints

**`POST /api/v1/auth/token`**

```json
{ "username": "demo", "password": "Qr-Demo-2026" }
→ { "token": "eyJ…", "expiresAt": "…", "tokenType": "Bearer" }
```

**`POST /api/v1/qr`** · requiere `Authorization: Bearer <token>`

```json
{ "matrix": [[12,-51,4],[6,167,-68],[-4,24,-41]] }
```

```json
{
  "rows": 3, "columns": 3,
  "q": [[0.857,-0.394,-0.331],[0.429,0.903,0.034],[-0.286,0.171,-0.943]],
  "r": [[14,21,-14],[0,175,-70],[0,0,35]],
  "statistics": {
    "max": 175, "min": -70, "sum": 161.44, "average": 8.969, "count": 18,
    "diagonal": { "q": false, "r": false }, "anyDiagonal": false
  }
}
```

Errores: `400` JSON mal formado · `401` token inválido · `422` matriz inválida ·
`502` la API de estadísticas falla.

La API de Node expone `POST /api/v1/statistics`, que recibe `{ q, r }` y devuelve el bloque
`statistics`.

## Decisiones

**QR, no rotación.** El enunciado pide la factorización QR en los requisitos funcionales, pero
menciona "rotación" al describir la arquitectura. Tomé QR porque es lo que dice la sección de
requisitos, y porque la operación adicional habla de *"las matrices"* y de si *"alguna"* es
diagonal: en plural, porque la factorización devuelve dos.

**Gram-Schmidt modificado, escrito a mano.** El algoritmo es el ejercicio, así que no usé
`gonum`. La variante modificada evita la pérdida de ortogonalidad por redondeo. Requiere
`m ≥ n` y rango completo; si no, responde `422`.

**Una sola respuesta.** `POST /qr` devuelve `q`, `r` y `statistics` juntas: el cliente hace una
petición y no necesita conocer la API interna.

**JWT con un secreto compartido.** Go emite los tokens y ambas APIs los verifican. Un único
usuario de demo por variables de entorno; ninguna de las dos arranca sin `JWT_SECRET`.

**Tolerancia al comparar con cero.** `isDiagonal` usa `|x| ≤ 1e-9` porque Q y R salen de un
cálculo en coma flotante: en el ejemplo de arriba `R[0][1]` vale `20.999999999999996`.

**Arquitectura idiomática de cada lenguaje.** En Go, paquetes `internal/` con el dominio sin
dependencias y la interfaz `StatsProvider` declarada en el consumidor. En Node, capas
`routes → controller → domain` con las dependencias inyectadas desde `container.js`. En ambos
casos el dominio no sabe que existe HTTP.

**Las pruebas verifican propiedades.** La de QR comprueba `Q·R = A` y `QᵀQ = I` en vez de
comparar contra números fijos, así sigue siendo válida si cambia el algoritmo.

**Código en inglés, mensajes en español.** Funciones, variables y pruebas en inglés, que es la
convención; los mensajes de error de la API, el README y la interfaz web en español, que es el
idioma de quien evalúa.

## Despliegue

Pensado para Google Cloud Run, un servicio por imagen: ambos son HTTP sin estado y escalan a
cero. La API de Node se despliega con `--ingress internal` para reproducir el aislamiento que
da `docker compose`, y los secretos van en Secret Manager.
