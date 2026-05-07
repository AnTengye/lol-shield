# LCU API Notes

Updated: 2026-05-06

## Source priority

1. Riot Developer Portal: authoritative policy and supported API boundary.
2. `https://lcu.kebs.dev/`: practical LCU endpoint index. Current page shows Client Version `26.05`.
3. `https://www.mingweisamuel.com/lcu-schema/`: useful historical schema, but it warns that Riot removed LCU OpenAPI/Swagger and the static spec may be out of date.

Riot classifies League Client API as local desktop communication used by the League client UI and says it is unsupported for third-party applications. Use it conservatively, keep requests local, and expect endpoints or fields to change.

## Local auth

The app obtains the LCU port and remoting token from the running League client process, then calls:

```text
https://riot:<token>@127.0.0.1:<port>
```

The client uses a self-signed certificate, so local tools usually need TLS verification disabled.

## Endpoints used by lol-shield

### Current summoner

```http
GET /lol-summoner/v1/current-summoner
```

Used for the logged-in user's `puuid`, Riot ID game name, tag line, summoner ID, icon, and level.

### Gameflow phase

```http
GET /lol-gameflow/v1/gameflow-phase
```

Used to detect the current client phase. In this project, `InProgress` means the realtime page can read cached current-game data.

### Gameflow session

```http
GET /lol-gameflow/v1/session
```

Used to inspect current game/session participants and queue metadata.

### Match history by PUUID

```http
GET /lol-match-history/v1/products/lol/{puuid}/matches?begIndex={begin}&endIndex={end}
```

Parameters:

- `puuid`: player PUUID.
- `begIndex`: zero-based first match index.
- `endIndex`: last match index. Treat it as inclusive for LCU match history paging.

For a normal page request:

```text
begin = page * pageSize
end = begin + pageSize - 1
```

The LCU response includes `games.gameCount`, which should be returned to the frontend as `total` so UI pagination can compute real page count.

### Match detail

```http
GET /lol-match-history/v1/games/{gameId}
```

Used by the match detail view. The project rate-limits these requests because asking LCU for many match summaries too quickly can make the client unstable.

### Ranked data

```http
GET /lol-ranked/v1/current-ranked-stats
GET /lol-ranked/v1/ranked-stats/{puuid}
```

Used for local user rank and selected player rank. Cache per PUUID because ranked stats are relatively static during one app session.

### Champion assets

```http
GET /lol-game-data/assets{path}
```

Used to proxy champion icons and skin loading screens through the local backend.

## Pagination note

The previous frontend mixed `totalPages` and `pageSize`: it used the same value both as the page count for the pagination control and as the request `pageSize`. Because the backend only returned an array, the frontend could not know the real number of pages. The fixed contract returns:

```json
{
  "list": [],
  "page": 0,
  "pageSize": 9,
  "total": 42
}
```

`totalPages` should be derived in the UI as `ceil(total / pageSize)`.
