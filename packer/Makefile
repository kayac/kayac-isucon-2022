isucon:
	git clone git@github.com:kayac/kayac-isucon-2022.git isucon

isucon.tar.gz: clean isucon
	cd isucon && make dataset
	cd isucon/bench && \
	GOOS=linux GOARCH=amd64 make bench && mv bench bench.amd64 && \
	GOOS=linux GOARCH=arm64 make bench && mv bench bench.arm64
	rm -fr isucon/.git *.tar.gz
	tar cvzf isucon.tar.gz isucon/

clean:
	rm -fr isucon isucon.tar.gz

ami: isucon.tar.gz
	packer build .
