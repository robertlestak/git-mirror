bin: bin/git-mirror_darwin bin/git-mirror_linux bin/git-mirror_windows
bin: bin/git-mirrord_darwin bin/git-mirrord_linux bin/git-mirrord_windows

bin/git-mirror_darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -o bin/git-mirror_darwin cmd/git-mirror/*.go
	openssl sha512 bin/git-mirror_darwin > bin/git-mirror_darwin.sha512

bin/git-mirror_linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/git-mirror_linux cmd/git-mirror/*.go
	openssl sha512 bin/git-mirror_linux > bin/git-mirror_linux.sha512

bin/git-mirror_windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -o bin/git-mirror_windows cmd/git-mirror/*.go
	openssl sha512 bin/git-mirror_windows > bin/git-mirror_windows.sha512

bin/git-mirrord_darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -o bin/git-mirrord_darwin cmd/git-mirrord/*.go
	openssl sha512 bin/git-mirrord_darwin > bin/git-mirrord_darwin.sha512

bin/git-mirrord_linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/git-mirrord_linux cmd/git-mirrord/*.go
	openssl sha512 bin/git-mirrord_linux > bin/git-mirrord_linux.sha512

bin/git-mirrord_windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -o bin/git-mirrord_windows cmd/git-mirrord/*.go
	openssl sha512 bin/git-mirrord_windows > bin/git-mirrord_windows.sha512