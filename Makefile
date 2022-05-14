.PHONY: dataset

dataset:
	curl -sL https://github.com/kayac/kayac-isucon-2022/releases/download/v0.0.1/isucon_listen80_dump.tar.gz > isucon_listen80_dump.tar.gz
	# gh release download v0.0.1
	tar xvf isucon_listen80_dump.tar.gz
	cp songs.json users.json bench/data/
	cp songs.json webapp/public/assets/
	mv isucon_listen80_dump.sql sql/90_isucon_listen80_dump.sql
	rm -f songs.json users.json

clean:
	rm -f isucon_listen80_dump.tar.gz bench/data/* webapp/public/assets/songs.json
