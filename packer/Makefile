isucon:
	git clone git@github.com:kayac/kayac-isucon-2022.git isucon

isucon.tar.gz: clean isucon
	cd isucon && make dataset && rm *.tar.gz
	cd isucon/bench && make bench
	rm -fr isucon/.git
	tar cvzf isucon.tar.gz isucon/

clean:
	rm -fr isucon isucon.tar.gz

ami: isucon.tar.gz
	packer build .
