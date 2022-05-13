.PHONY: dataset

dataset:
	# gh release download v1.0.0
	# curl -sL https://github.com/kayac/kayac-isucon-2022/release/...
	tar xvf isucon_listen80_dump.tar.gz
	cp songs.json users.json bench/data/
	cp songs.json webapp/public/assets/
	mv isucon_listen80_dump.sql sql/90_isucon_listen80_dump.sql
	rm -f songs.json users.json
