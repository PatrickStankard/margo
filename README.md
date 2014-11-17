margo
=====

A command-line tool for remote execution via SSH.

```bash
go get github.com/PatrickStankard/margo
$GOPATH/bin/margo -config "/tmp/config.sample.json" -job "Example Job"
```

```bash
echo "3" - patrick@example.box.one:22:
3

echo "4" - patrick@example.box.one:22:
4

echo "2" - patrick@example.box.two:22:
2

echo "1" - patrick@example.box.one:22:
1

echo "2" - patrick@example.box.one:22:
2

echo "5" - patrick@example.box.one:22:
5

echo "5" - patrick@example.box.two:22:
5

echo "3" - patrick@example.box.two:22:
3

echo "4" - patrick@example.box.two:22:
4

echo "1" - patrick@example.box.two:22:
1

echo "1" - patrick@example.box.three:22:
1

echo "2" - patrick@example.box.three:22:
2

echo "3" - patrick@example.box.three:22:
3

echo "5" - patrick@example.box.three:22:
5

echo "4" - patrick@example.box.three:22:
4
```
