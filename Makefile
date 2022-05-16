.PHONY: dataset

dataset: isucon_listen80_dump.tar.gz
	tar xvf isucon_listen80_dump.tar.gz
	cp songs.json users.json bench/data/
	cp songs.json webapp/public/assets/
	mv isucon_listen80_dump.sql sql/90_isucon_listen80_dump.sql
	rm -f songs.json users.json

isucon_listen80_dump.tar.gz:
	curl -sLO https://github.com/kayac/kayac-isucon-2022/releases/download/v0.0.1/isucon_listen80_dump.tar.gz

clean:
	rm -f isucon_listen80_dump.tar.gz bench/data/* webapp/public/assets/songs.json
