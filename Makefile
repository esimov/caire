version = 1.0.1
clean:
	rm -f caire
install:
    cd ./caire-${version}/cmd/caire/ && go install
    cp $GOPATH/bin/caire /usr/local/bin
uninstall:
	rm -f /usr/local/bin/caire