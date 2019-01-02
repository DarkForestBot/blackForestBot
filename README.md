Black Forest bot (aka Dark Forest bot)
===

A telegram bot like werewolfbot with a game like PUBG

Rules
---

1. Every player will get a position `(x, y)` (`x` and `y` are double of players). Every player can shoot at a position every night. Player will be killed when someone shoot at his position.
2. Every shoot would expose one axis of his position.
3. Player would convert to a BEAST when all axes of his position is exposed. When two players kill each other, BEAST will live.
4. Player who abort action at night will convert to BEAST.
5. Two players can unite each other, only of they know.
6. Players who join a union would not expose his position.
7. Player can betray his union at night and kill another one, no position is needed.
8. Player can set a trap at night, player who betrayed you will be killed
9. Only one player live, or all DEAD

Requirements
---
1. Go >= 1.10.3
2. redis >= 3.0.0
3. MySQL >= 5.8 or MariaDB

Build
---
```shell
go get -v git.wetofu.top/tonychee7000/blackForestBot # check the directory name, it must be `blackForestBot`
cd blackForestBot
./build.sh
```

Configuraion File
----
`config.json`

```json
{
    "tgApiToken": "OMITTED",  // You may need a bot for this.
    "debug": true,
    "updateTimeout": 60,
    "database": "bl:bl@tcp(127.0.0.1:3306)/blackforest?charset=utf8mb4&parseTime=True",
    "redis": "redis://localhost:6379/1",
    "adminPassword": "OMITTED",  // Change it.
    "threadLimit": 1024
}
```

Running
----
You may need to upload some gif for running this game. Just run the bot, and input `/admin YOUR_PASSWORD` to make admin mode **ON**.
Then just upload gif to the bot with filename in *win*, *lose*, *start*, *killed*, *trapped*, *beast* and finally use `/admin` without password to make admin mode **OFF**.

Or you may use [official bot](https://t.me/dark_forest_game_bot)

Donation
----

[Patreon](https://www.patreon.com/TonyChyi)
[PayPal](https://paypal.me/tonychee7000)
