# Isabelle
Accept and post about After the Collapse crashes in our development channels

## Set up
If you are using Go 1.14+, download it using `go get` (this will place the executable in `$GOBIN`):

```
go get -u get.cutie.cafe/isabelle
```

Otherwise, you should be able to use a stable release from the Releases section. NB: if you're using ancient versions of Linux, you may have to build the app yourself.

Then, download config.json.example and move it to config.json. Fill out all the fields. After, you can just run `isabelle.exe` or `./isabelle`.

## Advanced/extra steps
We're dealing with a lot of user-generated content here, so maaaaybe you should run Isabelle inside something like [ops](https://ops.city) or Docker just in case. A sample ops.json file has been included. Use it like this:

```
ops run -n ~/go/bin/isabelle -c ops.json
```

## Building
Clone the repository and then run `go build`. A `go.mod` and `go.sum` have been included, so dependencies should resolve just fine.

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