FILES=isucon_listen80_dump.sql songs.json users.json

dump:
	mysql -uisucon -pisucon isucon_listen80 < ../sql/50_listen80_schema.sql
	python3 build_fake_data.py 30000
	mysqldump -uisucon -pisucon -B isucon_listen80 --no-tablespaces > isucon_listen80_dump.sql
	tar czvf isucon_listen80_dump.tar.gz $(FILES)

clean:
	rm -f isucon_listen80_dump.* $(FILES)
