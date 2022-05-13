#!/usr/bin/env python3
import time
from faker import Faker
import bcrypt
import ulid
import random
import mysql.connector
import csv
import json
import os
import sys
import numpy as np
import traceback

random.seed(42) # fix seed
args = sys.argv
salt = bcrypt.gensalt(rounds=4)

fake = Faker(['en-US', 'ja-JP'])
fake.seed_instance(time.monotonic_ns())
fake_en = Faker(['en-US'])
fake_en.seed_instance(time.monotonic_ns())

tick = 0
counter = ["|", "/", "-", "\\"]
def print_counter():
    global tick
    tick += 1
    print("\r" + counter[tick % len(counter)] + "  " + str(tick), end="")

print('make fake words')
fake_words = fake.words(nb=10000)
print('make fake en words')
fake_words_en = fake_en.words(nb=10000)
print('make fake passwords')
fake_passwords = list()
for i in range(1, 100):
    raw = fake.password(special_chars=False)
    fake_passwords.append([bcrypt.hashpw(raw.encode('utf-8'), salt), raw])
    # print_counter()

# 作成数
cycle = 10 # 分割数
sets = int(args[1]) if len(args) > 1 else 100000 # 最終的に欲しいセット数

# 一周で作る数
user_num = int(sets / cycle) * 1
artist_num = int(sets / cycle) * 1

# 生成数上限
max_album_num_each_artist = 20
max_song_num_each_album = 15
max_playlist_num_each_user = 10000

# アイテム生成数を適当にロングテールで求める
rng = np.random.default_rng()
powers = list(rng.power(1, 10000))
powers_idx = 0
def num_of_items(maxn):
    global powers_idx
    powers_idx += 1
    if powers_idx >= len(powers):
        powers_idx = 0
    return min(int(1 / powers[powers_idx]), maxn)

# 行数
# user      = user
user_rows = user_num
# playlist  = user * playlist/user
# artist    = artist
artist_rows = artist_num
# song      = artist * album * song

print('user_rows', user_rows)
print('artist_rows', artist_rows)
print('1 loop', sum)

# exit()


users = list()
all_users = list()
initial_users = [
    {
        'account': 'isucon',
        'display_name': 'isucon',
        'password_hash': '$2b$11$smSqDIk.wv4UlzxCFDHqeOD922guLSKoeHmjWgfzOttlaNu65xlKW', # isuconpass
        'is_ban': False,
        'created_at': '1970-01-01 00:00:01',
        'last_logined_at': '1970-01-01 00:00:01',
    },
    {
        'account': 'dummy',
        'display_name': 'dummypass',
        'password_hash': '$2b$11$dUin2kiGvI8MBZxXTL2qweX6RTkm5LmHuLnmDAdEMx0YjBHRMa64q', # dummypass
        'is_ban': False,
        'created_at': '1970-01-01 00:00:01',
        'last_logined_at': '1970-01-01 00:00:01',
    },
    {
        'account': 'adminuser',
        'display_name': '管理者',
        'password_hash': '$2b$12$ccW9mADvL5UhZS42HbqspOcYmo/im8ScJ2vkewkxywYRnd/bLVwrC', # adminpass
        'is_ban': False,
        'created_at': '1970-01-01 00:00:01',
        'last_logined_at': '1970-01-01 00:00:01',
    },
]
users_json_data = list()
playlists = list()
artists = list()
songs = list()
playlist_songs = list()
playlist_favorites = list()


def make_playlist_song():
    for p in exist_playlists:
        # print_counter()
        n = int(random.random() * 80)
        song_id_list = random.sample(exist_song_ids, n)
        for i, sid in enumerate(song_id_list):
            plsng = {
                'playlist_id': p['id'],
                'song_id': sid,
                'sort_order': i + 1,
            }
            playlist_songs.append(plsng)

def make_playlist_favorite():
    for u in all_users:
        # print_counter()
        n = u['_favs']
        fav_playlists = random.sample(exist_playlists, n)
        for p in fav_playlists:
            plfv = {
                'playlist_id': p['id'],
                'favorite_user_account': u['account'],
                'created_at': fake.date_time_between(start_date=p['created_at'], end_date='now').isoformat(),
            }
            playlist_favorites.append(plfv)
    for p in exist_playlists:
        n = int(num_of_items(len(all_users)/2))
        fav_user_account_list = random.sample(all_users, n)
        for user in fav_user_account_list:
            plfv = {
                'playlist_id': p['id'],
                'favorite_user_account': user['account'],
                'created_at': fake.date_time_between(start_date=p['created_at'], end_date='now').isoformat(),
            }
            playlist_favorites.append(plfv)

def make_user():
    for user_index in range(1, user_num+1):
        # print_counter()
        user_account = ''.join(random.sample(fake_words_en, 3))
        display_name = ''.join(random.sample(fake_words, 3))
        password = random.choice(fake_passwords)
        created_at = fake.date_time_between(start_date='-1y', end_date='now')
        usr = {
            'account': user_account,
            'display_name': display_name,
            'password_raw': password[1],
            'password_hash': password[0].decode('utf-8'),
            'is_ban': random.random() < 0.05, # 5% ban
            'created_at': created_at.isoformat(),
            'last_logined_at': fake.date_time_between(start_date=created_at, end_date='now').isoformat(),
            '_favs': num_of_items(max_playlist_num_each_user),
        }
        # print('\nplaylist...')
        n = num_of_items(max_playlist_num_each_user)
        for playlist_index in range(1, n):
            playlist_name = ''.join(random.sample(fake_words, 3))
            created_at = fake.date_time_between(start_date='-1y', end_date='now')
            pl = {
                # id = playlist_index (auto inc)
                'ulid': ulid.from_timestamp(created_at).str,
                'name': playlist_name,
                'user_account': user_account,
                'is_public': random.choice([True, False]),
                'created_at': created_at.isoformat(),
                'updated_at': fake.date_time_between(start_date=created_at, end_date='now').isoformat(),
            }
            playlists.append(pl)

        users.append(usr)
        all_users.append(usr)
        users_json_data.append({
            'account': usr['account'],
            'password': usr['password_raw'],
            'is_ban': usr['is_ban'],
            'is_heavy': (n > 100 or usr['_favs'] > 100),
        })


def make_song():
    for artist_index in range(1, artist_num+1):
        # print_counter()
        artist_name = ''.join(random.sample(fake_words, 3))
        ar = {
            # id auto increment
            'ulid': ulid.new().str,
            'name': artist_name
        }
        artists.append(ar)

        n = num_of_items(max_album_num_each_artist)
        for album_index in range(1, n+1):
            album_name = ' '.join(random.sample(fake_words, 2))
            # print('\nalbum', album_name)

            m = num_of_items(max_song_num_each_album)
            for song_index in range(1, m+1):
                song_title = ''.join(random.sample(fake_words, 3))
                sng = {
                    # id auto increment
                    'ulid': ulid.new().str,
                    'title': song_title,
                    'artist_id': artist_index,
                    'artist_name': artist_name,
                    'album': album_name,
                    'track_number': song_index,
                    'is_public': True
                }
                songs.append(sng)


# insert

cnx = mysql.connector.connect(
    user='isucon',
    password='isucon',
    host='localhost',
    database = 'isucon_listen80',
    allow_local_infile=True
)

if cnx.is_connected:
    print("Connected!")

csr = cnx.cursor()

try:
    for t in ['user', 'playlist', 'artist', 'song', 'playlist_song', 'playlist_favorite']:
        csr.execute(f'TRUNCATE {t}')

    for master in range(0, cycle):
        users = initial_users if master == 0 else list()
        playlists = list()
        artists = list()
        songs = list()
        playlist_songs = list()

        print(f"================ master {master} ==============")


        make_user()
        print('user num:', len(users))
        with open('/tmp/user', 'w') as f:
            writer = csv.writer(f, lineterminator='\n')
            rows = list()
            for row in users:
                rows.append([row['account'], row['display_name'], row['password_hash'], ("1" if row['is_ban'] else "0"), row['created_at'], row['last_logined_at'], row['last_logined_at']])
            writer.writerows(rows)
        query_load_user = ("""
            LOAD DATA LOCAL INFILE '/tmp/user'
                IGNORE
                INTO TABLE user
                FIELDS TERMINATED BY ','
                LINES TERMINATED BY '\n'
                (`account`, `display_name`, `password_hash`, `is_ban`, `created_at`, `last_logined_at`);
        """)
        csr.execute(query_load_user)
        cnx.commit()

        print('playlist num:', len(playlists))
        with open('/tmp/playlist', 'w') as f:
            writer = csv.writer(f, lineterminator='\n')
            rows = list()
            for row in playlists:
                rows.append([row['ulid'], row['name'], row['user_account'], int(row['is_public']), row['created_at'], row['updated_at']])
            writer.writerows(rows)
        query_load_playlist = ("""
            LOAD DATA LOCAL INFILE '/tmp/playlist'
                IGNORE
                INTO TABLE playlist
                FIELDS TERMINATED BY ','
                LINES TERMINATED BY '\n'
                (`ulid`, `name`, `user_account`, `is_public`, `created_at`, `updated_at`);
        """)
        csr.execute(query_load_playlist)
        cnx.commit()

        make_song()
        print('artist num:', len(artists))
        with open('/tmp/artist', 'w') as f:
            writer = csv.writer(f, lineterminator='\n')
            rows = list()
            for row in artists:
                rows.append([row['ulid'], row['name']])
            writer.writerows(rows)
        query_load_artist = ("""
            LOAD DATA LOCAL INFILE '/tmp/artist'
                IGNORE
                INTO TABLE artist
                FIELDS TERMINATED BY ','
                LINES TERMINATED BY '\n'
                (`ulid`, `name`);
        """)
        csr.execute(query_load_artist)
        cnx.commit()

        print('song num:', len(songs))
        with open('/tmp/song', 'w') as f:
            writer = csv.writer(f, lineterminator='\n')
            rows = list()
            for row in songs:
                rows.append([row['ulid'], row['title'], row['artist_id'], row['album'], row['track_number'], int(row['is_public'])])
            writer.writerows(rows)
        query_load_song = ("""
            LOAD DATA LOCAL INFILE '/tmp/song'
                IGNORE
                INTO TABLE song
                FIELDS TERMINATED BY ','
                LINES TERMINATED BY '\n'
                (`ulid`, `title`, `artist_id`, `album`, `track_number`, `is_public`);
        """)
        csr.execute(query_load_song)
        cnx.commit()

    try:
        os.remove('users.json')
    except Exception:
        pass
    with open('users.json', 'a') as users_json:
        json.dump(random.sample(users_json_data, min(10000, len(users_json_data))), users_json, indent=2)

    try:
        os.remove('songs.json')
    except Exception:
        pass
    with open('songs.json', 'a') as songs_json:
        json.dump(random.sample(songs, min(10000, len(songs))), songs_json, indent=2)

    query_select_playlist_ids = ("""SELECT id, created_at FROM playlist""")
    csr.execute(query_select_playlist_ids)
    res = csr.fetchall()
    exist_playlists = [{'id': p[0], 'created_at': p[1]} for p in res]
    print('exist_playlists', len(exist_playlists))

    query_select_song_ids = ("""SELECT id FROM song""")
    csr.execute(query_select_song_ids)
    res = csr.fetchall()
    exist_song_ids = [sid[0] for sid in res]
    print('exist_song_ids', len(exist_song_ids))

    make_playlist_song()
    print('playlist song num:', len(playlist_songs))
    with open('/tmp/playlist_song', 'w') as f:
        writer = csv.writer(f, lineterminator='\n')
        rows = list()
        for row in playlist_songs:
            rows.append([row['playlist_id'], row['song_id'], row['sort_order']])
        writer.writerows(rows)

    query_load_playlist_song = ("""
        LOAD DATA LOCAL INFILE '/tmp/playlist_song'
            IGNORE
            INTO TABLE playlist_song
            FIELDS TERMINATED BY ','
            LINES TERMINATED BY '\n'
            (`playlist_id`, `song_id`, `sort_order`);
    """)
    csr.execute(query_load_playlist_song)
    cnx.commit()

    make_playlist_favorite()
    print('playlist favorite num:', len(playlist_favorites))

    with open('/tmp/playlist_favorite', 'w') as f:
        writer = csv.writer(f, lineterminator='\n')
        rows = list()
        for row in playlist_favorites:
            rows.append([row['playlist_id'], row['favorite_user_account'], row['created_at']])
        writer.writerows(rows)

    query_load_playlist_favorite = ("""
        LOAD DATA LOCAL INFILE '/tmp/playlist_favorite'
            IGNORE
            INTO TABLE playlist_favorite
            FIELDS TERMINATED BY ','
            LINES TERMINATED BY '\n'
            (`playlist_id`, `favorite_user_account`, `created_at`);
    """)
    csr.execute(query_load_playlist_favorite)
    cnx.commit()


except Exception as e:
    print(f"Error Occurred: {e}")
    traceback.print_exc()


finally:
    csr.close()
    cnx.close()
