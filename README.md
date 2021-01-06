# Isabelle
Accept app crash ZIPs via Discord or e-mail and post them as Discord messages and GitHub issues

[![Build](https://github.com/AnarkisGaming/isabelle/workflows/Build/badge.svg?event=push)](https://github.com/AnarkisGaming/isabelle/actions) ![Publish Docker image](https://github.com/AnarkisGaming/isabelle/workflows/Publish%20Docker%20image/badge.svg)

## Set up

### Natively
If you are using Go 1.15+, download it using `go get` (this will place the executable in `$GOBIN`):

```
GO111MODULE=on go get -u get.cutie.cafe/isabelle
```

Otherwise, you should be able to use a [stable release](https://github.com/AnarkisGaming/isabelle/releases) or [development release](https://github.com/AnarkisGaming/isabelle/actions). If you're using ancient versions of Linux, you may have to build the app yourself (see below).

Then, download config.json.example and move it to config.json. Fill out all of the fields. After, you can just run `isabelle.exe` or `./isabelle`.

### via Docker
Download config.json.example, rename it to config.json, and fill out all of the fields. Then:

```
docker run -it -v /path/to/config.json:/app/config.json -d anarkisgaming/isabelle
```

### via `ops`
A sample `ops.json` file is provided. Install isabelle, make sure you have a valid `config.json` and then:

```
ops run -n ~/go/bin/isabelle -c ops.json
```

## Building
Clone the repository and then run `GO111MODULE=on go build` to install dependencies and build.

## License
The GNU AGPL. See the `LICENSE` file for details. In short:

```
Isabelle
Accept app crash ZIPs via Discord or e-mail and post them as Discord messages
and GitHub issues

Copyright (C) 2020 Anarkis Gaming/Cutie Café.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```
