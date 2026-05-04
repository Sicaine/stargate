# dhd

Stargate control web app — DHD and SCDC interfaces talking to a hardware API.

## Build

```sh
go build -o dhd .
```

## Run

```sh
./dhd
```

Listens on `:8080` by default. Override with `PORT` and `EXTERNAL_API` env vars:

```sh
PORT=9090 EXTERNAL_API=http://192.168.1.42:9000 ./dhd
```

## Routes

| Path    | Description                  |
|---------|------------------------------|
| `/`     | Landing page                 |
| `/dhd`  | Dial Home Device interface   |
| `/scdc` | Stargate Command Dialing Computer |

## API

| Method | Path              | Description                        |
|--------|-------------------|------------------------------------|
| POST   | `/api/dial`       | Dial with `{"symbols":[1,2,…]}`    |
| POST   | `/api/disconnect` | Disengage active wormhole          |
| POST   | `/api/iris`       | Toggle iris open/close (SCDC only) |
| GET    | `/api/state`      | Current gate and iris state        |

## Credits

Sound effects from [cap_resources](https://github.com/RafaelDeJongh/cap_resources) by RafaelDeJongh.
